// Command vibelog menjalankan server VibeLog: API, SPA, dan SQLite dalam
// satu binary.
package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enpidi26/vibelog/internal/config"
	"github.com/enpidi26/vibelog/internal/database"
	"github.com/enpidi26/vibelog/internal/email"
	"github.com/enpidi26/vibelog/internal/httpapi"
)

// version diisi saat build: -ldflags "-X main.version=1.2.3".
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "backup":
			if err := runBackup(os.Args[2:]); err != nil {
				log.Fatalf("backup failed: %v", err)
			}
			return
		case "version", "-v", "--version":
			println("vibelog " + version)
			return
		case "help", "-h", "--help":
			usage()
			return
		default:
			os.Stderr.WriteString("unknown command " + os.Args[1] + "\n\n")
			usage()
			os.Exit(2)
		}
	}

	cfg := config.Load()
	if err := cfg.EnsureJWTSecret(); err != nil {
		log.Fatalf("jwt secret: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	mailer, err := email.New(email.Config{
		Driver:       cfg.EmailDriver,
		From:         cfg.EmailFrom,
		SMTPHost:     cfg.SMTPHost,
		SMTPPort:     cfg.SMTPPort,
		SMTPUsername: cfg.SMTPUsername,
		SMTPPassword: cfg.SMTPPassword,
		SMTPTLS:      cfg.SMTPTLS,
		ResendAPIKey: cfg.ResendAPIKey,
	})
	if err != nil {
		log.Fatalf("email configuration: %v", err)
	}

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	srv := httpapi.New(httpapi.Deps{Config: cfg, DB: db, Mailer: mailer})

	cleanupDone := make(chan struct{})
	go startTokenCleanup(cleanupDone, db)
	defer close(cleanupDone)

	httpServer := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("vibelog %s listening on %s (env=%s, base_url=%s, email=%s, registration=%t)",
			version, cfg.Addr(), cfg.Env, cfg.BaseURL, mailer.Name(), cfg.AllowRegistration)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

// startTokenCleanup menghapus token magic link kadaluarsa sekali sejam.
func startTokenCleanup(done <-chan struct{}, db *sql.DB) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := database.CleanupExpiredTokens(db); err != nil {
				log.Printf("token cleanup: %v", err)
			}
		}
	}
}
