package database

import (
	"testing"

	"github.com/enpidi26/vibelog/internal/domain"
)

func TestProjectEntryFlowIsolationDuplicateSummary(t *testing.T) {
	db, _ := testDB(t)
	ana := testUser(t, db, "ana@test.id")
	budi := testUser(t, db, "budi@test.id")

	project, err := CreateProject(db, ana.ID, domain.ProjectInput{Name: "VibeLog", Description: "Jurnal", Color: "teal"})
	if err != nil {
		t.Fatal(err)
	}
	if err := DeleteProject(db, budi.ID, project.ID); err != domain.ErrNotFound {
		t.Fatalf("Budi menghapus proyek Ana harus ErrNotFound, got %v", err)
	}

	rating := 5
	entry, err := CreateEntry(db, ana.ID, domain.EntryInput{
		ProjectID: &project.ID, Title: "Refactor auth", PromptText: "Buatkan refactor...",
		Model: "Claude", Tool: "Cursor", Outcome: domain.OutcomeSuccess, Rating: &rating,
		ResultNotes: "Berhasil, test hijau.", Tags: []string{"Refactor", "auth"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if entry.IterationNo != 1 || len(entry.Tags) != 2 || entry.Tags[0] != "refactor" {
		t.Fatalf("entri tidak normal: %+v", entry)
	}

	if _, err := UpdateEntry(db, budi.ID, entry.ID, domain.EntryInput{
		ProjectID: &project.ID, Title: "Diretas", PromptText: "x", Model: "Claude", Tool: "Cursor", Outcome: domain.OutcomeFailed,
	}); err != domain.ErrNotFound {
		t.Fatalf("Budi mengubah entri Ana harus ErrNotFound, got %v", err)
	}

	dup, err := DuplicateEntry(db, ana.ID, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if dup.IterationNo != 2 || dup.Outcome != domain.OutcomeInProgress || dup.ParentID == nil || *dup.ParentID != entry.ID {
		t.Fatalf("duplikat salah: %+v", dup)
	}

	summ, err := EntrySummary(db, ana.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if summ.Total != 2 || summ.Success != 1 || summ.InProgress != 1 || summ.RatedCount != 1 {
		t.Fatalf("summary salah: %+v", summ)
	}
	if summ.AverageRating == nil || *summ.AverageRating != 5 {
		t.Fatalf("rata-rata rating harus 5: %+v", summ)
	}

	listed, err := ListEntries(db, ana.ID, domain.EntryFilter{Query: "refactor", Outcome: domain.OutcomeSuccess})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 {
		t.Fatalf("filter harus 1 entri, dapat %d", len(listed))
	}

	if err := DeleteProject(db, ana.ID, project.ID); err != nil {
		t.Fatal(err)
	}
	kept, err := GetEntry(db, ana.ID, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if kept.ProjectID != nil {
		t.Fatal("entri harus bertahan dengan project_id null setelah proyek dihapus")
	}

	noProject, err := ListEntries(db, ana.ID, domain.EntryFilter{ProjectID: ptrInt64(0)})
	if err != nil {
		t.Fatal(err)
	}
	if len(noProject) != 2 {
		t.Fatalf("filter tanpa proyek harus 2 entri, dapat %d", len(noProject))
	}
}

func TestSummaryEmptyUser(t *testing.T) {
	db, _ := testDB(t)
	u := testUser(t, db, "kosong@test.id")
	summ, err := EntrySummary(db, u.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if summ.Total != 0 || summ.AverageRating != nil || len(summ.ByModel) != 0 {
		t.Fatalf("summary kosong salah: %+v", summ)
	}
	if _, err := EntrySummary(db, u.ID, ptrInt64(0)); err != nil {
		t.Fatalf("summary project_id=0: %v", err)
	}
}

func TestExportIncludesAllEntries(t *testing.T) {
	db, _ := testDB(t)
	u := testUser(t, db, "ana@test.id")
	for i := 0; i < 250; i++ {
		if _, err := db.Exec(`INSERT INTO entries
			(user_id, title, prompt_text, model, tool, outcome, result_notes, tags, iteration_no, created_at, updated_at)
			VALUES (?, 'entry', 'p', 'Claude', 'Cursor', 'success', '', '', 1, ?, ?)`,
			u.ID, utcNow(), utcNow()); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := ListEntries(db, u.ID, domain.EntryFilter{All: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 250 {
		t.Fatalf("export must include every entry, got %d", len(entries))
	}
}

func TestSearchAndTagFiltersEscapeWildcards(t *testing.T) {
	db, _ := testDB(t)
	u := testUser(t, db, "ana@test.id")
	if _, err := CreateEntry(db, u.ID, domain.EntryInput{
		Title: "Refactor", PromptText: "prompt", Model: "Claude", Tool: "Cursor",
		Outcome: domain.OutcomeSuccess, Tags: []string{"vibe coding"},
	}); err != nil {
		t.Fatal(err)
	}

	listed, err := ListEntries(db, u.ID, domain.EntryFilter{Query: "%"})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 0 {
		t.Fatalf("q=%% must be literal, got %d entries", len(listed))
	}
	listed, err = ListEntries(db, u.ID, domain.EntryFilter{Tag: "_"})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 0 {
		t.Fatalf("tag=_ must be literal, got %d entries", len(listed))
	}
	listed, err = ListEntries(db, u.ID, domain.EntryFilter{Tag: "vibe coding"})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 {
		t.Fatalf("tag filter must match, got %d entries", len(listed))
	}
}

func TestTagFilterTreatsUnderscoreLiterally(t *testing.T) {
	db, _ := testDB(t)
	u := testUser(t, db, "ana@test.id")
	for _, tag := range []string{"vibe_coding", "vibexcoding"} {
		if _, err := CreateEntry(db, u.ID, domain.EntryInput{
			Title: "Entry " + tag, PromptText: "p", Model: "Claude", Tool: "Cursor",
			Outcome: domain.OutcomeSuccess, Tags: []string{tag},
		}); err != nil {
			t.Fatal(err)
		}
	}
	listed, err := ListEntries(db, u.ID, domain.EntryFilter{Tag: "vibe_coding"})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || len(listed[0].Tags) != 1 || listed[0].Tags[0] != "vibe_coding" {
		t.Fatalf("underscore must be literal, got %+v", listed)
	}
}

func ptrInt64(v int64) *int64 { return &v }
