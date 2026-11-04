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
)

func main() {
	cfg := LoadConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}
	db, err := OpenDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	srv := newServer(cfg, db)
	httpServer := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           srv.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("vibelog-api berjalan di http://%s (env=%s)", cfg.Addr(), cfg.Env)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("mematikan server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func newServer(cfg Config, db *sql.DB) *server {
	return &server{
		cfg:       cfg,
		db:        db,
		authLimit: NewRateLimiter(cfg.AuthRateLimit, time.Minute),
		apiLimit:  NewRateLimiter(cfg.APIRateLimit, time.Minute),
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)

	mux.Handle("POST /api/auth/register", rateLimit(s.authLimit, http.HandlerFunc(s.handleRegister)))
	mux.Handle("POST /api/auth/login", rateLimit(s.authLimit, http.HandlerFunc(s.handleLogin)))
	mux.Handle("POST /api/auth/verify-otp", rateLimit(s.authLimit, http.HandlerFunc(s.handleVerifyOTP)))
	mux.Handle("POST /api/auth/resend-otp", rateLimit(s.authLimit, http.HandlerFunc(s.handleResendOTP)))

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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
	})

	handler := securityHeaders(cors(s.cfg, mux))
	handler = recoverer(handler)
	return requestTimeout(30*time.Second, handler)
}
