package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/enpidi26/vibelog/internal/domain"
)

func TestProjectEntryFlowIsolationDuplicateSummaryExport(t *testing.T) {
	srv, _ := testServer(t, nil)
	tokenA := registerToken(t, srv, "Ana", "ana@test.id")
	tokenB := registerToken(t, srv, "Budi", "budi@test.id")

	rec := doJSON(t, srv, "POST", "/api/projects", tokenA, domain.ProjectInput{Name: "VibeLog", Description: "Jurnal", Color: "teal"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("buat proyek: %d %s", rec.Code, rec.Body.String())
	}
	var project domain.Project
	_ = json.Unmarshal(rec.Body.Bytes(), &project)

	rating := 5
	rec = doJSON(t, srv, "POST", "/api/entries", tokenA, domain.EntryInput{
		ProjectID: &project.ID, Title: "Refactor auth", PromptText: "Buatkan refactor...",
		Model: "Claude", Tool: "Cursor", Outcome: domain.OutcomeSuccess, Rating: &rating,
		ResultNotes: "Berhasil, test hijau.", Tags: []string{"Refactor", "auth"},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("buat entri: %d %s", rec.Code, rec.Body.String())
	}
	var entry domain.Entry
	_ = json.Unmarshal(rec.Body.Bytes(), &entry)
	if entry.IterationNo != 1 || len(entry.Tags) != 2 || entry.Tags[0] != "refactor" {
		t.Fatalf("entri tidak normal: %+v", entry)
	}

	rec = doJSON(t, srv, "PATCH", "/api/entries/"+itoa(entry.ID), tokenB, domain.EntryInput{
		Title: "Diretas", PromptText: "x", Model: "Claude", Tool: "Cursor", Outcome: domain.OutcomeFailed,
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("Budi mengubah entri Ana harus 404, dapat %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, srv, "POST", "/api/entries/"+itoa(entry.ID)+"/duplicate", tokenA, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("duplikat: %d %s", rec.Code, rec.Body.String())
	}
	var dup domain.Entry
	_ = json.Unmarshal(rec.Body.Bytes(), &dup)
	if dup.IterationNo != 2 || dup.Outcome != domain.OutcomeInProgress || dup.ParentID == nil || *dup.ParentID != entry.ID {
		t.Fatalf("duplikat salah: %+v", dup)
	}

	rec = doJSON(t, srv, "GET", "/api/entries/summary", tokenA, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("summary: %d %s", rec.Code, rec.Body.String())
	}
	var summ domain.Summary
	_ = json.Unmarshal(rec.Body.Bytes(), &summ)
	if summ.Total != 2 || summ.Success != 1 || summ.InProgress != 1 || summ.RatedCount != 1 {
		t.Fatalf("summary salah: %+v", summ)
	}
	if summ.AverageRating == nil || *summ.AverageRating != 5 {
		t.Fatalf("rata-rata rating harus 5: %+v", summ)
	}

	rec = doJSON(t, srv, "GET", "/api/entries?q=refactor&outcome=success", tokenA, nil)
	var listed []domain.Entry
	_ = json.Unmarshal(rec.Body.Bytes(), &listed)
	if len(listed) != 1 {
		t.Fatalf("filter harus 1 entri, dapat %d", len(listed))
	}

	rec = doJSON(t, srv, "GET", "/api/entries/export?format=md", tokenA, nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Refactor auth") || !strings.Contains(rec.Body.String(), "Buatkan refactor") {
		t.Fatalf("export md salah: %d %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, srv, "DELETE", "/api/projects/"+itoa(project.ID), tokenA, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("hapus proyek: %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, srv, "GET", "/api/entries/"+itoa(entry.ID), tokenA, nil)
	var kept domain.Entry
	_ = json.Unmarshal(rec.Body.Bytes(), &kept)
	if kept.ProjectID != nil {
		t.Fatal("entri harus bertahan dengan project_id null setelah proyek dihapus")
	}

	rec = doJSON(t, srv, "GET", "/api/entries?project_id=0", tokenA, nil)
	var noProject []domain.Entry
	_ = json.Unmarshal(rec.Body.Bytes(), &noProject)
	if len(noProject) != 2 {
		t.Fatalf("filter tanpa proyek harus 2 entri, dapat %d", len(noProject))
	}
}

func TestSummaryEmptyUser(t *testing.T) {
	srv, _ := testServer(t, nil)
	token := registerToken(t, srv, "Kosong", "kosong@test.id")
	rec := doJSON(t, srv, "GET", "/api/entries/summary", token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("summary kosong harus 200, dapat %d: %s", rec.Code, rec.Body.String())
	}
	var summ domain.Summary
	_ = json.Unmarshal(rec.Body.Bytes(), &summ)
	if summ.Total != 0 || summ.AverageRating != nil || len(summ.ByModel) != 0 {
		t.Fatalf("summary kosong salah: %+v", summ)
	}
	rec = doJSON(t, srv, "GET", "/api/entries/summary?project_id=0", token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("summary project_id=0 kosong harus 200, dapat %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHealthReportsDatabase(t *testing.T) {
	srv, _ := testServer(t, nil)
	rec := doJSON(t, srv, "GET", "/health", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("health: %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"db":"ok"`) {
		t.Fatalf("health body missing db status: %s", rec.Body.String())
	}
	rec = doJSON(t, srv, "GET", "/livez", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("livez: %d", rec.Code)
	}
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
