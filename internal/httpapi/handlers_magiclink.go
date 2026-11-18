package httpapi

import (
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/enpidi26/vibelog/internal/auth"
	"github.com/enpidi26/vibelog/internal/database"
	"github.com/enpidi26/vibelog/internal/domain"
	"github.com/enpidi26/vibelog/internal/email"
)

const magicLinkTTL = 15 * time.Minute

// handleRequestMagicLink handles POST /api/auth/email/request.
//
// Respons selalu sama supaya endpoint ini tidak bisa dipakai menebak email
// mana yang terdaftar. Akun baru hanya dibuat kalau registrasi diizinkan dan
// email benar-benar berhasil dikirim.
func (s *Server) handleRequestMagicLink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	emailAddr, ok := domain.NormalizeEmail(req.Email)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_email", "Enter a valid email address")
		return
	}
	generic := func() {
		writeJSON(w, http.StatusOK, map[string]string{
			"message": "If that email can sign in, a link is on its way.",
		})
	}

	if s.mailer.Name() == "off" {
		writeError(w, http.StatusServiceUnavailable, "email_disabled",
			"Email sign-in is not configured on this instance. Use a password instead.")
		return
	}

	user, err := database.GetUserByEmail(s.db, emailAddr)
	created := false
	if errors.Is(err, domain.ErrNotFound) {
		if !s.cfg.AllowRegistration {
			generic()
			return
		}
		name := strings.Split(emailAddr, "@")[0]
		if len(name) > 80 {
			name = name[:80]
		}
		user, err = database.CreateUser(s.db, name, emailAddr, "")
		if errors.Is(err, domain.ErrEmailTaken) {
			user, err = database.GetUserByEmail(s.db, emailAddr)
		}
		if err != nil {
			log.Printf("magic link: create user: %v", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "Could not start sign-in")
			return
		}
		created = true
	} else if err != nil {
		log.Printf("magic link: lookup user: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not start sign-in")
		return
	}

	token, err := auth.RandomToken(32)
	if err != nil {
		s.rollbackMagicUser(created, user.ID)
		log.Printf("magic link: generate token: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not start sign-in")
		return
	}
	if err := database.IssueMagicToken(s.db, user.ID, token, magicLinkTTL); err != nil {
		s.rollbackMagicUser(created, user.ID)
		log.Printf("magic link: store token: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not start sign-in")
		return
	}

	magicURL := fmt.Sprintf("%s/auth/verify?token=%s", s.cfg.BaseURL, token)
	subject, htmlBody, textBody := magicLinkMessage(magicURL)
	if err := s.mailer.Send(emailAddr, subject, htmlBody, textBody); err != nil {
		s.rollbackMagicUser(created, user.ID)
		log.Printf("magic link: send email to %s: %v", emailAddr, err)
		if errors.Is(err, email.ErrDisabled) {
			writeError(w, http.StatusServiceUnavailable, "email_disabled", "Email sign-in is not configured on this instance.")
			return
		}
		writeError(w, http.StatusBadGateway, "email_failed", "The email could not be sent. Try again in a moment.")
		return
	}
	generic()
}

func (s *Server) rollbackMagicUser(created bool, userID int64) {
	if !created {
		return
	}
	if err := database.DeleteUser(s.db, userID); err != nil {
		log.Printf("magic link: rollback user %d: %v", userID, err)
	}
}

func magicLinkMessage(magicURL string) (subject, htmlBody, textBody string) {
	subject = "Your VibeLog sign-in link"
	textBody = "Sign in to VibeLog:\n\n" + magicURL + "\n\nThis link expires in 15 minutes. If you did not request it, ignore this email.\n"
	htmlBody = fmt.Sprintf(`<div style="font-family:system-ui,sans-serif;max-width:560px;margin:0 auto;line-height:1.5">
  <h2 style="margin:0 0 12px">Sign in to VibeLog</h2>
  <p>Click the button below. The link expires in 15 minutes.</p>
  <p style="margin:24px 0"><a href="%s" style="display:inline-block;background:#111;color:#fff;padding:12px 20px;text-decoration:none;border-radius:4px">Sign in</a></p>
  <p style="color:#555;font-size:14px">Or paste this URL into your browser:<br><code style="word-break:break-all">%s</code></p>
  <p style="color:#777;font-size:13px">If you did not request this email, you can ignore it.</p>
</div>`, html.EscapeString(magicURL), html.EscapeString(magicURL))
	return subject, htmlBody, textBody
}

// handleVerifyMagicLink handles POST /api/auth/email/verify with {"token": "..."}.
// Verifikasi lewat POST supaya pemindai email/prefetch tidak membakar tautan
// sekali pakai sebelum pemiliknya sempat mengklik.
func (s *Server) handleVerifyMagicLink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	token := strings.TrimSpace(req.Token)
	if len(token) < 32 || len(token) > 128 {
		writeError(w, http.StatusBadRequest, "invalid_token", "This sign-in link is invalid or has expired.")
		return
	}

	userID, err := database.ConsumeMagicToken(s.db, token)
	switch {
	case errors.Is(err, database.ErrTokenInvalid):
		writeError(w, http.StatusBadRequest, "invalid_token", "This sign-in link is invalid or has expired.")
		return
	case errors.Is(err, database.ErrTokenUsed):
		writeError(w, http.StatusBadRequest, "token_used", "This sign-in link has already been used.")
		return
	case errors.Is(err, database.ErrTokenExpired):
		writeError(w, http.StatusBadRequest, "token_expired", "This sign-in link has expired. Request a new one.")
		return
	case err != nil:
		log.Printf("magic link verify: consume token: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not sign you in")
		return
	}

	user, err := database.GetUserByID(s.db, userID)
	if err != nil {
		log.Printf("magic link verify: load user: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not sign you in")
		return
	}
	jwtToken, err := s.issueToken(user)
	if err != nil {
		log.Printf("magic link verify: issue JWT: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not sign you in")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": jwtToken,
		"user":  publicUser(user),
	})
}
