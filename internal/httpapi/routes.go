package httpapi

import (
	"net/http"
	"strings"
	"time"
)

// Handler mengembalikan seluruh routing, termasuk middleware global.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /livez", s.handleLive)
	mux.HandleFunc("GET /api/meta", s.handleMeta)

	mux.Handle("POST /api/auth/register", s.rateLimit(s.authLimit, http.HandlerFunc(s.handleRegister)))
	mux.Handle("POST /api/auth/login", s.rateLimit(s.authLimit, http.HandlerFunc(s.handleLogin)))
	mux.Handle("GET /api/auth/github", s.rateLimit(s.authLimit, http.HandlerFunc(s.handleGithubLogin)))
	mux.Handle("GET /api/auth/github/callback", http.HandlerFunc(s.handleGithubCallback))
	mux.Handle("POST /api/auth/email/request", s.rateLimit(s.authLimit, http.HandlerFunc(s.handleRequestMagicLink)))
	mux.Handle("POST /api/auth/email/verify", s.rateLimit(s.authLimit, http.HandlerFunc(s.handleVerifyMagicLink)))

	protected := func(h http.HandlerFunc) http.Handler {
		return s.rateLimit(s.apiLimit, s.authRequired(h))
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
