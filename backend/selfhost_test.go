package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type captureMailer struct {
	to, subject, html, text string
	err                     error
}

func (m *captureMailer) Send(to, subject, html, text string) error {
	if m.err != nil {
		return m.err
	}
	m.to, m.subject, m.html, m.text = to, subject, html, text
	return nil
}

func (m *captureMailer) Name() string { return "capture" }

func magicTokenFromMail(t *testing.T, text string) string {
	t.Helper()
	idx := strings.Index(text, "/auth/verify?token=")
	if idx < 0 {
		t.Fatalf("magic link not found in email: %q", text)
	}
	raw := text[idx+len("/auth/verify?token="):]
	raw = strings.Fields(raw)[0]
	token, err := url.QueryUnescape(raw)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return token
}

func TestMigrationsAreVersionedAndIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "migrate.db")
	for i := 0; i < 2; i++ {
		db, err := OpenDB(path)
		if err != nil {
			t.Fatalf("open %d: %v", i, err)
		}
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
			t.Fatalf("count migrations: %v", err)
		}
		if count != len(migrations) {
			t.Fatalf("expected %d migrations, got %d", len(migrations), count)
		}
		db.Close()
	}

	db, err := sql.Open("sqlite", sqliteDSN(path))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cols, err := tableColumnsDB(db, "users")
	if err != nil {
		t.Fatal(err)
	}
	for _, col := range []string{"email_verified", "email_verified_at"} {
		if !cols[col] {
			t.Fatalf("users.%s missing", col)
		}
	}
}

func tableColumnsDB(db *sql.DB, table string) (map[string]bool, error) {
	rows, err := db.Query(`SELECT name FROM pragma_table_info(?)`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out[name] = true
	}
	return out, rows.Err()
}

func TestRegistrationCanBeDisabled(t *testing.T) {
	srv := testDB(t)
	srv.cfg.AllowRegistration = false
	capture := &captureMailer{}
	srv.mailer = capture

	rec := doJSON(t, srv, "POST", "/api/auth/register", "", RegisterInput{Name: "Ana", Email: "ana@test.id", Password: "password123"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("registration should be 403, got %d: %s", rec.Code, rec.Body.String())
	}

	// Email tak dikenal tidak boleh membuat akun, dan respons harus generik.
	rec = doJSON(t, srv, "POST", "/api/auth/email/request", "", map[string]string{"email": "baru@test.id"})
	if rec.Code != http.StatusOK {
		t.Fatalf("magic request should be generic 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if capture.to != "" {
		t.Fatal("no email should be sent when registration is closed")
	}
	if _, err := GetUserByEmail(srv.db, "baru@test.id"); err == nil {
		t.Fatal("user must not be created when registration is closed")
	}
}

func TestMagicLinkRequestVerifyAndReplay(t *testing.T) {
	srv := testDB(t)
	capture := &captureMailer{}
	srv.mailer = capture

	rec := doJSON(t, srv, "POST", "/api/auth/email/request", "", map[string]string{"email": "Magic@test.id"})
	if rec.Code != http.StatusOK {
		t.Fatalf("request magic link: %d %s", rec.Code, rec.Body.String())
	}
	if capture.to != "magic@test.id" {
		t.Fatalf("email should be normalized, got %q", capture.to)
	}
	token := magicTokenFromMail(t, capture.text)

	// GET tidak boleh membakar token (404/405 sama-sama aman).
	rec = doJSON(t, srv, "GET", "/api/auth/email/verify?token="+token, "", nil)
	if rec.Code == http.StatusOK {
		t.Fatalf("GET verify must not succeed, got %d", rec.Code)
	}

	rec = doJSON(t, srv, "POST", "/api/auth/email/verify", "", map[string]string{"token": token})
	if rec.Code != http.StatusOK {
		t.Fatalf("verify: %d %s", rec.Code, rec.Body.String())
	}
	var auth struct {
		Token string `json:"token"`
		User  struct {
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &auth); err != nil {
		t.Fatal(err)
	}
	if auth.Token == "" || auth.User.Email != "magic@test.id" || !auth.User.EmailVerified {
		t.Fatalf("bad auth payload: %+v", auth)
	}

	rec = doJSON(t, srv, "POST", "/api/auth/email/verify", "", map[string]string{"token": token})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("replay should be 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMagicLinkSecondRequestInvalidatesFirstToken(t *testing.T) {
	srv := testDB(t)
	capture := &captureMailer{}
	srv.mailer = capture

	rec := doJSON(t, srv, "POST", "/api/auth/email/request", "", map[string]string{"email": "magic@test.id"})
	if rec.Code != http.StatusOK {
		t.Fatalf("first request: %d", rec.Code)
	}
	first := magicTokenFromMail(t, capture.text)

	rec = doJSON(t, srv, "POST", "/api/auth/email/request", "", map[string]string{"email": "magic@test.id"})
	if rec.Code != http.StatusOK {
		t.Fatalf("second request: %d", rec.Code)
	}
	second := magicTokenFromMail(t, capture.text)
	if first == second {
		t.Fatal("tokens should differ")
	}

	rec = doJSON(t, srv, "POST", "/api/auth/email/verify", "", map[string]string{"token": first})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("old token should be invalid, got %d", rec.Code)
	}
	rec = doJSON(t, srv, "POST", "/api/auth/email/verify", "", map[string]string{"token": second})
	if rec.Code != http.StatusOK {
		t.Fatalf("new token should work, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEnsureJWTSecretGeneratedAndPersisted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jwt_secret")
	a := Config{}
	a.JWTSecretPath = path
	if err := a.EnsureJWTSecret(); err != nil {
		t.Fatal(err)
	}
	if len(a.JWTSecret) < 32 {
		t.Fatalf("secret too short: %d", len(a.JWTSecret))
	}
	b := Config{}
	b.JWTSecretPath = path
	if err := b.EnsureJWTSecret(); err != nil {
		t.Fatal(err)
	}
	if b.JWTSecret != a.JWTSecret {
		t.Fatal("secret must be stable across restarts")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("secret file must not be empty")
	}
}

func TestExplicitJWTSecretWins(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jwt_secret")
	cfg := Config{JWTSecret: strings.Repeat("x", 40), JWTSecretPath: path}
	if err := cfg.EnsureJWTSecret(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("secret file should not be written when JWT_SECRET is set")
	}
}

func TestRotateBackupsKeepsNewest(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"vibelog-20260101-000000.db", "vibelog-20260102-000000.db", "vibelog-20260103-000000.db"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := rotateBackups(dir, 2); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 backups, got %d", len(entries))
	}
	if entries[0].Name() != "vibelog-20260102-000000.db" || entries[1].Name() != "vibelog-20260103-000000.db" {
		t.Fatalf("wrong backups kept: %v %v", entries[0].Name(), entries[1].Name())
	}
}

func TestSearchAndTagFiltersEscapeWildcards(t *testing.T) {
	srv := testDB(t)
	token := registerToken(t, srv, "Ana", "ana@test.id")
	rec := doJSON(t, srv, "POST", "/api/entries", token, EntryInput{
		Title: "Refactor", PromptText: "prompt", Model: "Claude", Tool: "Cursor",
		Outcome: "success", Tags: []string{"vibe coding"},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create entry: %d %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, srv, "GET", "/api/entries?q=%25", token, nil)
	var listed []Entry
	_ = json.Unmarshal(rec.Body.Bytes(), &listed)
	if len(listed) != 0 {
		t.Fatalf("q=%% must be literal, got %d entries", len(listed))
	}

	rec = doJSON(t, srv, "GET", "/api/entries?tag=_", token, nil)
	listed = nil
	_ = json.Unmarshal(rec.Body.Bytes(), &listed)
	if len(listed) != 0 {
		t.Fatalf("tag=_ must be literal, got %d entries", len(listed))
	}

	rec = doJSON(t, srv, "GET", "/api/entries?tag=vibe%20coding", token, nil)
	listed = nil
	_ = json.Unmarshal(rec.Body.Bytes(), &listed)
	if len(listed) != 1 {
		t.Fatalf("tag filter must match, got %d entries", len(listed))
	}
}

func TestHealthReportsDatabase(t *testing.T) {
	srv := testDB(t)
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("health: %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"db":"ok"`) {
		t.Fatalf("health body missing db status: %s", rec.Body.String())
	}
}

func TestEmailDisabledRejectsMagicLink(t *testing.T) {
	srv := testDB(t)
	srv.mailer = offMailer{}
	rec := doJSON(t, srv, "POST", "/api/auth/email/request", "", map[string]string{"email": "a@b.test"})
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("email off should be 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestExportIncludesAllEntries(t *testing.T) {
	srv := testDB(t)
	token := registerToken(t, srv, "Ana", "ana@test.id")
	for i := 0; i < 250; i++ {
		if _, err := srv.db.Exec(`INSERT INTO entries
			(user_id, title, prompt_text, model, tool, outcome, result_notes, tags, iteration_no, created_at, updated_at)
			VALUES (1, ?, 'p', 'Claude', 'Cursor', 'success', '', '', 1, ?, ?)`,
			"entry", utcNow(), utcNow()); err != nil {
			t.Fatal(err)
		}
	}
	rec := doJSON(t, srv, "GET", "/api/entries/export?format=json", token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("export: %d %s", rec.Code, rec.Body.String())
	}
	var entries []Entry
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 250 {
		t.Fatalf("export must include every entry, got %d", len(entries))
	}
}

func TestMIMEMessageUsesCRLF(t *testing.T) {
	from, err := mail.ParseAddress("VibeLog <no-reply@example.com>")
	if err != nil {
		t.Fatal(err)
	}
	msg := string(buildMIMEMessage(from, "user@test.id", "Halo", "<p>Hi</p>"))
	if strings.Contains(msg, "\r\r\n") {
		t.Fatal("MIME message must not contain doubled CR")
	}
	for _, want := range []string{"From: ", "To: user@test.id", "Subject: ", "Content-Type: text/html"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("missing %q in message: %q", want, msg)
		}
	}
}

func TestTagFilterTreatsUnderscoreLiterally(t *testing.T) {
	srv := testDB(t)
	token := registerToken(t, srv, "Ana", "ana@test.id")
	for i, tag := range []string{"vibe_coding", "vibexcoding"} {
		rec := doJSON(t, srv, "POST", "/api/entries", token, EntryInput{
			Title: "Entry " + tag, PromptText: "p", Model: "Claude", Tool: "Cursor",
			Outcome: "success", Tags: []string{tag},
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("create %d: %d %s", i, rec.Code, rec.Body.String())
		}
	}
	rec := doJSON(t, srv, "GET", "/api/entries?tag=vibe_coding", token, nil)
	var listed []Entry
	_ = json.Unmarshal(rec.Body.Bytes(), &listed)
	if len(listed) != 1 || len(listed[0].Tags) != 1 || listed[0].Tags[0] != "vibe_coding" {
		t.Fatalf("underscore must be literal, got %+v", listed)
	}
}

func TestDatabasePragmas(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pragmas.db")
	db, err := OpenDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var mode string
	if err := db.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if strings.ToLower(mode) != "wal" {
		t.Fatalf("journal_mode = %q, want wal", mode)
	}
	var fk int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatal(err)
	}
	if fk != 1 {
		t.Fatalf("foreign_keys = %d, want 1", fk)
	}
	var busy int
	if err := db.QueryRow(`PRAGMA busy_timeout`).Scan(&busy); err != nil {
		t.Fatal(err)
	}
	if busy != 5000 {
		t.Fatalf("busy_timeout = %d, want 5000", busy)
	}
}
