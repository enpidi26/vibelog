package httpapi

import (
	"errors"
	"log"
	"net/http"

	"github.com/enpidi26/vibelog/internal/auth"
	"github.com/enpidi26/vibelog/internal/database"
	"github.com/enpidi26/vibelog/internal/domain"
)

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.AllowRegistration {
		writeError(w, http.StatusForbidden, "registration_closed", "Registration is closed on this instance.")
		return
	}
	var in domain.RegisterInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to process password")
		return
	}
	user, err := database.CreateUser(s.db, in.Name, in.Email, hash)
	if errors.Is(err, domain.ErrEmailTaken) {
		writeError(w, http.StatusConflict, "email_taken", "email already registered")
		return
	}
	if err != nil {
		log.Printf("register: create user failed: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create account")
		return
	}

	token, err := s.issueToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create token")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"token": token,
		"user":  publicUser(user),
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var in domain.LoginInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	user, err := database.GetUserByEmail(s.db, in.Email)
	if errors.Is(err, domain.ErrNotFound) {
		// Pesan sama untuk email tidak ada dan password salah, supaya
		// penyerang tidak bisa menebak akun yang terdaftar.
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "incorrect email or password")
		return
	}
	if err != nil {
		log.Printf("login: lookup user: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to sign in")
		return
	}
	if !auth.CheckPassword(in.Password, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "incorrect email or password")
		return
	}

	token, err := s.issueToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  publicUser(user),
	})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, err := database.GetUserByID(s.db, userIDFromContext(r.Context()))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load profile")
		return
	}
	writeJSON(w, http.StatusOK, publicUser(user))
}

func (s *Server) issueToken(user domain.User) (string, error) {
	return auth.IssueToken(s.cfg.EffectiveJWTSecret(), user.ID, user.Email, user.Name, s.cfg.JWTExpiryHrs)
}

func publicUser(u domain.User) map[string]any {
	return map[string]any{
		"id":             u.ID,
		"name":           u.Name,
		"email":          u.Email,
		"email_verified": u.EmailVerified,
		"created_at":     u.CreatedAt,
	}
}
