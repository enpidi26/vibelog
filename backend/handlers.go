package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type server struct {
	cfg       Config
	db        *sql.DB
	authLimit *RateLimiter
	apiLimit  *RateLimiter
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "vibelog-api"})
}

func (s *server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var in RegisterInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	hash, err := HashPassword(in.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to process password")
		return
	}
	user, err := CreateUser(s.db, in.Name, in.Email, hash)
	if errors.Is(err, ErrEmailTaken) {
		writeError(w, http.StatusConflict, "email_taken", "email already registered")
		return
	}
	if err != nil {
		fmt.Printf("register: create user failed: %v\n", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create account")
		return
	}

	token, err := IssueToken(s.cfg.EffectiveJWTSecret(), user.ID, user.Email, user.Name, s.cfg.JWTExpiryHrs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create token")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"token": token,
		"user":  publicUser(user),
	})
}

func (s *server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var in LoginInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	user, err := GetUserByEmail(s.db, in.Email)
	if errors.Is(err, ErrNotFound) || !CheckPassword(in.Password, user.PasswordHash) {
		// Pesan sama untuk email tidak ada dan password salah, supaya
		// penyerang tidak bisa menebak akun yang terdaftar.
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "incorrect email or password")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to sign in")
		return
	}
	
	// Cek apakah email sudah diverifikasi
	if !user.EmailVerified {
		writeError(w, http.StatusForbidden, "email_not_verified", "email not verified. Please check your email for the verification code.")
		return
	}
	
	token, err := IssueToken(s.cfg.EffectiveJWTSecret(), user.ID, user.Email, user.Name, s.cfg.JWTExpiryHrs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  publicUser(user),
	})
}

func publicUser(u User) map[string]any {
	return map[string]any{
		"id":             u.ID,
		"name":           u.Name,
		"email":          u.Email,
		"email_verified": u.EmailVerified,
		"created_at":     u.CreatedAt,
	}
}

func (s *server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserByID(s.db, userIDFromContext(r.Context()))
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load profile")
		return
	}
	writeJSON(w, http.StatusOK, publicUser(user))
}

// ---- Projects ----

func (s *server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := ListProjects(s.db, userIDFromContext(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load projects")
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (s *server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var in ProjectInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	p, err := CreateProject(s.db, userIDFromContext(r.Context()), in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create project")
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (s *server) handlePatchProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_id", "id tidak valid")
		return
	}
	var in ProjectInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	p, err := UpdateProject(s.db, userIDFromContext(r.Context()), id, in)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to update project")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_id", "id tidak valid")
		return
	}
	if err := DeleteProject(s.db, userIDFromContext(r.Context()), id); errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete project")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id})
}

// ---- Entries ----

func entryFilterFromQuery(r *http.Request) EntryFilter {
	q := r.URL.Query()
	f := EntryFilter{
		Query:   strings.TrimSpace(q.Get("q")),
		Model:   strings.TrimSpace(q.Get("model")),
		Tool:    strings.TrimSpace(q.Get("tool")),
		Outcome: strings.TrimSpace(q.Get("outcome")),
		Tag:     strings.TrimSpace(q.Get("tag")),
		Limit:   100,
	}
	if v := q.Get("project_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil && id >= 0 {
			f.ProjectID = &id
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Limit = n
		}
	}
	if f.Outcome != "" && !IsValidOutcome(f.Outcome) {
		f.Outcome = ""
	}
	return f
}

func (s *server) handleListEntries(w http.ResponseWriter, r *http.Request) {
	entries, err := ListEntries(s.db, userIDFromContext(r.Context()), entryFilterFromQuery(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load entries")
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (s *server) handleCreateEntry(w http.ResponseWriter, r *http.Request) {
	var in EntryInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	e, err := CreateEntry(s.db, userIDFromContext(r.Context()), in)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create entry")
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (s *server) handleGetEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_id", "id tidak valid")
		return
	}
	e, err := GetEntry(s.db, userIDFromContext(r.Context()), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "entry not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load entries")
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *server) handlePatchEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_id", "id tidak valid")
		return
	}
	var in EntryInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	e, err := UpdateEntry(s.db, userIDFromContext(r.Context()), id, in)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "entry or project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to update entry")
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *server) handleDeleteEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_id", "id tidak valid")
		return
	}
	if err := DeleteEntry(s.db, userIDFromContext(r.Context()), id); errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "entry not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete entry")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id})
}

func (s *server) handleDuplicateEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_id", "id tidak valid")
		return
	}
	e, err := DuplicateEntry(s.db, userIDFromContext(r.Context()), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "entry not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to duplicate entry")
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (s *server) handleSummary(w http.ResponseWriter, r *http.Request) {
	var projectID *int64
	if v := strings.TrimSpace(r.URL.Query().Get("project_id")); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil && id >= 0 {
			projectID = &id
		}
	}
	summ, err := EntrySummary(s.db, userIDFromContext(r.Context()), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "gagal memuat ringkasan")
		return
	}
	writeJSON(w, http.StatusOK, summ)
}

func (s *server) handleExport(w http.ResponseWriter, r *http.Request) {
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "md"
	}
	if format != "md" && format != "json" {
		writeError(w, http.StatusBadRequest, "validation_error", "format harus md atau json")
		return
	}
	entries, err := ListEntries(s.db, userIDFromContext(r.Context()), entryFilterFromQuery(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "gagal mengekspor entri")
		return
	}
	stamp := time.Now().Format("20060102")
	if format == "json" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="vibelog-export-%s.json"`, stamp))
		_ = json.NewEncoder(w).Encode(entries)
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="vibelog-export-%s.md"`, stamp))
	_, _ = w.Write([]byte(BuildMarkdown(entries)))
}
