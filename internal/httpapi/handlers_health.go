package httpapi

import (
	"context"
	"net/http"
	"time"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.db.PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "unavailable", "db": "error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "vibelog", "db": "ok"})
}

func (s *Server) handleLive(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

// handleMeta memberi tahu frontend fitur opsional mana yang aktif, supaya
// tombol GitHub/magic link tidak ditampilkan saat instance belum
// mengonfigurasinya.
func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	emailLogin := s.mailer != nil && s.mailer.Name() != "off"
	writeJSON(w, http.StatusOK, map[string]any{
		"email_login_enabled":  emailLogin,
		"email_delivery":       s.mailer.Name(),
		"github_login_enabled": s.cfg.GitHubClientID != "" && s.cfg.GitHubClientSecret != "",
		"registration_enabled": s.cfg.AllowRegistration,
	})
}
