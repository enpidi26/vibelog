package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/enpidi26/vibelog/internal/database"
	"github.com/enpidi26/vibelog/internal/export"
)

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "md"
	}
	if format != "md" && format != "json" {
		writeError(w, http.StatusBadRequest, "validation_error", "format harus md atau json")
		return
	}
	filter := entryFilterFromQuery(r)
	filter.All = true // ekspor memuat seluruh riwayat, bukan 100 terbaru
	entries, err := database.ListEntries(s.db, userIDFromContext(r.Context()), filter)
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
	_, _ = w.Write([]byte(export.Markdown(entries)))
}
