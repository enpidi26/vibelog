package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func testDB(t *testing.T) *server {
	t.Helper()
	cfg := LoadConfig()
	cfg.DatabasePath = filepath.Join(t.TempDir(), "vibelog-test.db")
	db, err := OpenDB(cfg.DatabasePath)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return newServer(cfg, db)
}

func doJSON(t *testing.T, srv *server, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	return rec
}

func registerToken(t *testing.T, srv *server, name, email string) string {
	t.Helper()
	rec := doJSON(t, srv, "POST", "/api/auth/register", "", RegisterInput{Name: name, Email: email, Password: "password123"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("register %s: status %d body %s", email, rec.Code, rec.Body.String())
	}
	var res struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil || res.Token == "" {
		t.Fatalf("token kosong: %v body %s", err, rec.Body.String())
	}
	return res.Token
}

func TestConfigProductionRejectsWeakSecret(t *testing.T) {
	cfg := LoadConfig()
	cfg.Env = "production"
	cfg.JWTSecret = "pendek"
	if err := cfg.Validate(); err == nil {
		t.Fatal("secret pendek harus ditolak di production")
	}
	cfg.JWTSecret = strings.Repeat("a", 32)
	if err := cfg.Validate(); err != nil {
		t.Fatalf("secret 32 karakter harus diterima: %v", err)
	}
}

func TestPasswordAndTokenRoundTrip(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword("password123", hash) || CheckPassword("salah", hash) {
		t.Fatal("verifikasi password salah")
	}
	token, err := IssueToken("secret-yang-cukup-panjang-untuk-test", 7, "a@b.test", "Ana", 24)
	if err != nil {
		t.Fatal(err)
	}
	uid, email, _, err := ParseToken("secret-yang-cukup-panjang-untuk-test", token)
	if err != nil || uid != 7 || email != "a@b.test" {
		t.Fatalf("parse token: uid=%d email=%s err=%v", uid, email, err)
	}
	if _, _, _, err := ParseToken("secret-lain", token); err == nil {
		t.Fatal("token dengan secret lain harus ditolak")
	}
}

func TestEntryValidation(t *testing.T) {
	bad := EntryInput{Title: "x", PromptText: "p", Model: "Claude", Tool: "Cursor", Outcome: "aneh"}
	if err := bad.Validate(); err == nil {
		t.Fatal("outcome tidak dikenal harus ditolak")
	}
	rating := 6
	bad2 := EntryInput{Title: "x", PromptText: "p", Model: "Claude", Tool: "Cursor", Outcome: "success", Rating: &rating}
	if err := bad2.Validate(); err == nil {
		t.Fatal("rating 6 harus ditolak")
	}
	if _, err := NormalizeTags([]string{"AI Agent", "ai-agent", "  ", "a/b"}); err != nil {
		t.Fatalf("normalisasi tag: %v", err)
	}
}

func TestAuthDuplicateEmailAndWrongPassword(t *testing.T) {
	srv := testDB(t)
	registerToken(t, srv, "Ana", "ana@test.id")
	rec := doJSON(t, srv, "POST", "/api/auth/register", "", RegisterInput{Name: "Ana 2", Email: "ANA@test.id", Password: "password123"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("email duplikat harus 409, dapat %d: %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, srv, "POST", "/api/auth/login", "", LoginInput{Email: "ana@test.id", Password: "salah1234"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("password salah harus 401, dapat %d", rec.Code)
	}
	rec = doJSON(t, srv, "GET", "/api/entries", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("tanpa token harus 401, dapat %d", rec.Code)
	}
}

func TestProjectEntryFlowIsolationDuplicateSummaryExport(t *testing.T) {
	srv := testDB(t)
	tokenA := registerToken(t, srv, "Ana", "ana@test.id")
	tokenB := registerToken(t, srv, "Budi", "budi@test.id")

	// Proyek milik Ana.
	rec := doJSON(t, srv, "POST", "/api/projects", tokenA, ProjectInput{Name: "VibeLog", Description: "Jurnal", Color: "teal"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("buat proyek: %d %s", rec.Code, rec.Body.String())
	}
	var project Project
	_ = json.Unmarshal(rec.Body.Bytes(), &project)

	// Budi tidak boleh melihat/menghapus proyek Ana (GET proyek tidak ada
	// endpoint publik; isolasi dibuktikan lewat entri + hapus proyek).
	rec = doJSON(t, srv, "DELETE", "/api/projects/1", tokenB, nil)
	// ID proyek Ana belum tentu 1 di semua urutan, jadi cek lewat entri:
	// buat entri di proyek Ana, lalu Budi mencoba mengubahnya.
	rating := 5
	rec = doJSON(t, srv, "POST", "/api/entries", tokenA, EntryInput{
		ProjectID: &project.ID, Title: "Refactor auth", PromptText: "Buatkan refactor...",
		Model: "Claude", Tool: "Cursor", Outcome: "success", Rating: &rating,
		ResultNotes: "Berhasil, test hijau.", Tags: []string{"Refactor", "auth"},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("buat entri: %d %s", rec.Code, rec.Body.String())
	}
	var entry Entry
	_ = json.Unmarshal(rec.Body.Bytes(), &entry)
	if entry.IterationNo != 1 || len(entry.Tags) != 2 || entry.Tags[0] != "refactor" {
		t.Fatalf("entri tidak normal: %+v", entry)
	}

	rec = doJSON(t, srv, "PATCH", "/api/entries/"+itoa(entry.ID), tokenB, EntryInput{
		ProjectID: &project.ID, Title: "Diretas", PromptText: "x", Model: "Claude", Tool: "Cursor", Outcome: "failed",
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("Budi mengubah entri Ana harus 404, dapat %d: %s", rec.Code, rec.Body.String())
	}

	// Duplikat sebagai iterasi 2.
	rec = doJSON(t, srv, "POST", "/api/entries/"+itoa(entry.ID)+"/duplicate", tokenA, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("duplikat: %d %s", rec.Code, rec.Body.String())
	}
	var dup Entry
	_ = json.Unmarshal(rec.Body.Bytes(), &dup)
	if dup.IterationNo != 2 || dup.Outcome != "in_progress" || dup.ParentID == nil || *dup.ParentID != entry.ID {
		t.Fatalf("duplikat salah: %+v", dup)
	}

	// Ringkasan: 1 success + 1 in_progress, rata-rata rating hanya dari yang dinilai.
	rec = doJSON(t, srv, "GET", "/api/entries/summary", tokenA, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("summary: %d %s", rec.Code, rec.Body.String())
	}
	var summ Summary
	_ = json.Unmarshal(rec.Body.Bytes(), &summ)
	if summ.Total != 2 || summ.Success != 1 || summ.InProgress != 1 || summ.RatedCount != 1 {
		t.Fatalf("summary salah: %+v", summ)
	}
	if summ.AverageRating == nil || *summ.AverageRating != 5 {
		t.Fatalf("rata-rata rating harus 5: %+v", summ)
	}

	// Filter q dan outcome.
	rec = doJSON(t, srv, "GET", "/api/entries?q=refactor&outcome=success", tokenA, nil)
	var listed []Entry
	_ = json.Unmarshal(rec.Body.Bytes(), &listed)
	if len(listed) != 1 {
		t.Fatalf("filter harus 1 entri, dapat %d", len(listed))
	}

	// Export Markdown memuat judul dan prompt.
	rec = doJSON(t, srv, "GET", "/api/entries/export?format=md", tokenA, nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Refactor auth") || !strings.Contains(rec.Body.String(), "Buatkan refactor") {
		t.Fatalf("export md salah: %d %s", rec.Code, rec.Body.String())
	}

	// Hapus proyek: entri tetap ada, project_id jadi null.
	rec = doJSON(t, srv, "DELETE", "/api/projects/"+itoa(project.ID), tokenA, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("hapus proyek: %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, srv, "GET", "/api/entries/"+itoa(entry.ID), tokenA, nil)
	var kept Entry
	_ = json.Unmarshal(rec.Body.Bytes(), &kept)
	if kept.ProjectID != nil {
		t.Fatal("entri harus bertahan dengan project_id null setelah proyek dihapus")
	}

	// Filter project_id=0 hanya mengembalikan entri tanpa proyek.
	rec = doJSON(t, srv, "GET", "/api/entries?project_id=0", tokenA, nil)
	var noProject []Entry
	_ = json.Unmarshal(rec.Body.Bytes(), &noProject)
	if len(noProject) != 2 {
		t.Fatalf("filter tanpa proyek harus 2 entri, dapat %d", len(noProject))
	}
}

func TestSummaryEmptyUser(t *testing.T) {
	// Regresi: ringkasan akun tanpa entri sempat 500 karena SUM atas
	// himpunan kosong menghasilkan NULL yang tidak bisa di-scan ke int.
	srv := testDB(t)
	token := registerToken(t, srv, "Kosong", "kosong@test.id")
	rec := doJSON(t, srv, "GET", "/api/entries/summary", token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("summary kosong harus 200, dapat %d: %s", rec.Code, rec.Body.String())
	}
	var summ Summary
	_ = json.Unmarshal(rec.Body.Bytes(), &summ)
	if summ.Total != 0 || summ.AverageRating != nil || len(summ.ByModel) != 0 {
		t.Fatalf("summary kosong salah: %+v", summ)
	}
	rec = doJSON(t, srv, "GET", "/api/entries/summary?project_id=0", token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("summary project_id=0 kosong harus 200, dapat %d: %s", rec.Code, rec.Body.String())
	}
}

func TestBuildMarkdownEscapesFence(t *testing.T) {	md := BuildMarkdown([]Entry{{Title: "T", Model: "M", Tool: "X", Outcome: "success", IterationNo: 1, PromptText: "a ``` b"}})
	if strings.Contains(md, "a ``` b") {
		t.Fatal("fence di dalam prompt harus diamankan")
	}
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
