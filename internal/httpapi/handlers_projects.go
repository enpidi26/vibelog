package httpapi

import (
	"errors"
	"net/http"

	"github.com/enpidi26/vibelog/internal/database"
	"github.com/enpidi26/vibelog/internal/domain"
)

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := database.ListProjects(s.db, userIDFromContext(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load projects")
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var in domain.ProjectInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	p, err := database.CreateProject(s.db, userIDFromContext(r.Context()), in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create project")
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) handlePatchProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_id", "id tidak valid")
		return
	}
	var in domain.ProjectInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := in.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	p, err := database.UpdateProject(s.db, userIDFromContext(r.Context()), id, in)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to update project")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_id", "id tidak valid")
		return
	}
	if err := database.DeleteProject(s.db, userIDFromContext(r.Context()), id); errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "project not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete project")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id})
}
