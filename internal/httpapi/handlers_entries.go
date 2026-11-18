package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/enpidi26/vibelog/internal/database"
	"github.com/enpidi26/vibelog/internal/domain"
)

func entryFilterFromQuery(r *http.Request) domain.EntryFilter {
	q := r.URL.Query()
	f := domain.EntryFilter{
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
	if f.Outcome != "" && !domain.IsValidOutcome(f.Outcome) {
		f.Outcome = ""
	}
	return f
}

func (s *Server) handleListEntries(w http.ResponseWriter, r *http.Request) {
	entries, err := database.ListEntries(s.db, userIDFromContext(r.Context()), entryFilterFromQuery(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load entries")
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (s *Server) handleCreateEntry(w http.ResponseWriter, r *http.Request) {
	var in domain.EntryInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	e, err := database.CreateEntry(s.db, userIDFromContext(r.Context()), in)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create entry")
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (s *Server) handleGetEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_id", "id tidak valid")
		return
	}
	e, err := database.GetEntry(s.db, userIDFromContext(r.Context()), id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "entry not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load entry")
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) handlePatchEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_id", "id tidak valid")
		return
	}
	var in domain.EntryInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	e, err := database.UpdateEntry(s.db, userIDFromContext(r.Context()), id, in)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "entry or project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to update entry")
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) handleDeleteEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_id", "id tidak valid")
		return
	}
	if err := database.DeleteEntry(s.db, userIDFromContext(r.Context()), id); errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "entry not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete entry")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id})
}

func (s *Server) handleDuplicateEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_id", "id tidak valid")
		return
	}
	e, err := database.DuplicateEntry(s.db, userIDFromContext(r.Context()), id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "entry not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to duplicate entry")
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	var projectID *int64
	if v := strings.TrimSpace(r.URL.Query().Get("project_id")); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil && id >= 0 {
			projectID = &id
		}
	}
	summ, err := database.EntrySummary(s.db, userIDFromContext(r.Context()), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "gagal memuat ringkasan")
		return
	}
	writeJSON(w, http.StatusOK, summ)
}
