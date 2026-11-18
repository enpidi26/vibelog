package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
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
			fmt.Println("vibelog " + version)
			return
		case "help", "-h", "--help":
			usage()
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
			usage()
			os.Exit(2)
		}
	}

	cfg := LoadConfig()
	if err := cfg.EnsureJWTSecret(); err != nil {
		log.Fatalf("jwt secret: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	mailer, err := NewMailer(cfg)
	if err != nil {
		log.Fatalf("email configuration: %v", err)
	}

	db, err := OpenDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	srv := newServer(cfg, db)
	srv.mailer = mailer

	cleanupDone := make(chan struct{})
	go startOTPCleanup(cleanupDone, db)
	defer close(cleanupDone)

	httpServer := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           srv.routes(),
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

func usage() {
	fmt.Print(`vibelog — prompt journal

Usage:
  vibelog              start the server (env-configured)
  vibelog backup [dir] write a consistent SQLite backup into dir
  vibelog version      print the version
  vibelog help         show this message

Environment:
  BASE_URL                 public URL, e.g. https://vibelog.example.com
  DATABASE_PATH            SQLite file (default data/vibelog.db)
  JWT_SECRET               optional; otherwise generated next to the database
  ALLOW_REGISTRATION       true|false (default true)
  EMAIL_DRIVER             console|smtp|resend|off (default console)
  SMTP_HOST/SMTP_PORT/...  SMTP settings when EMAIL_DRIVER=smtp
  RESEND_API_KEY           when EMAIL_DRIVER=resend
`)
}

func newServer(cfg Config, db *sql.DB) *server {
	return &server{
		cfg:       cfg,
		db:        db,
		authLimit: NewRateLimiter(cfg.AuthRateLimit, time.Minute, cfg.TrustedProxies),
		apiLimit:  NewRateLimiter(cfg.APIRateLimit, time.Minute, cfg.TrustedProxies),
		mailer:    consoleMailer{},
		static:    spaHandler(),
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /livez", s.handleLive)

	mux.Handle("POST /api/auth/register", rateLimit(s.authLimit, http.HandlerFunc(s.handleRegister)))
	mux.Handle("POST /api/auth/login", rateLimit(s.authLimit, http.HandlerFunc(s.handleLogin)))
	mux.Handle("GET /api/auth/github", rateLimit(s.authLimit, http.HandlerFunc(s.handleGithubLogin)))
	mux.Handle("GET /api/auth/github/callback", http.HandlerFunc(s.handleGithubCallback))
	mux.Handle("POST /api/auth/email/request", rateLimit(s.authLimit, http.HandlerFunc(s.handleRequestMagicLink)))
	mux.Handle("POST /api/auth/email/verify", rateLimit(s.authLimit, http.HandlerFunc(s.handleVerifyMagicLink)))

	protected := func(h http.HandlerFunc) http.Handler {
		return rateLimit(s.apiLimit, authRequired(s.cfg, h))
	}
	mux.Handle("GET /api/auth/me", protected(s.handleMe))
	mux.Handle("GET /api/projects", protected(s.handleListProjects))
	mux.Handle("POST /api/projects", protected(s.handleCreateProject))
	mux.Handle("PATCH /api/projects/{id}", protected(s.handlePatchProject))
	mux.Handle("DELETE /api/projects/{id}", protected(s.handleDeleteProject))
	mux.Handle("GET /api/entries/summary", protected(s.handleSummary))
	mux.Handle("GET /api/entries/export", protected(s.handleExport))
	mux.Handle("GET /api/entries", protected(s.handleListEntries))
	mux.Handle("POST /api/entries", protected(s.handleCreateEntry))
	mux.Handle("GET /api/entries/{id}", protected(s.handleGetEntry))
	mux.Handle("PATCH /api/entries/{id}", protected(s.handlePatchEntry))
	mux.Handle("DELETE /api/entries/{id}", protected(s.handleDeleteEntry))
	mux.Handle("POST /api/entries/{id}/duplicate", protected(s.handleDuplicateEntry))

	// API yang tidak dikenal tetap JSON; selain itu SPA (kalau frontend
	// di-embed) supaya refresh di /app tidak 404.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/health" || r.URL.Path == "/livez" {
			writeError(w, http.StatusNotFound, "not_found", "route not found")
			return
		}
		if s.static != nil {
			s.static.ServeHTTP(w, r)
			return
		}
		writeError(w, http.StatusNotFound, "not_found", "route not found")
	})

	handler := securityHeaders(logRequests(cors(s.cfg, mux)))
	handler = recoverer(handler)
	return requestTimeout(30*time.Second, handler)
}
