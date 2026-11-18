package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enpidi26/vibelog/internal/config"
	"github.com/enpidi26/vibelog/internal/database"
	"github.com/enpidi26/vibelog/internal/domain"
	"github.com/enpidi26/vibelog/internal/email"
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

type offTestMailer struct{}

func (offTestMailer) Send(_, _, _, _ string) error { return email.ErrDisabled }
func (offTestMailer) Name() string                 { return "off" }

func testServer(t *testing.T, mutate func(*config.Config)) (*Server, *captureMailer) {
	t.Helper()
	cfg := config.Load()
	cfg.DatabasePath = filepath.Join(t.TempDir(), "vibelog-test.db")
	cfg.BaseURL = "http://example.test"
	cfg.AuthRateLimit = 1000
	cfg.APIRateLimit = 1000
	cfg.AllowRegistration = true
	if mutate != nil {
		mutate(&cfg)
	}
	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	mailer := &captureMailer{}
	srv := New(Deps{Config: cfg, DB: db, Mailer: mailer})
	return srv, mailer
}

func doJSON(t *testing.T, srv *Server, method, path, token string, body any) *httptest.ResponseRecorder {
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
	srv.Handler().ServeHTTP(rec, req)
	return rec
}

func registerToken(t *testing.T, srv *Server, name, emailAddr string) string {
	t.Helper()
	rec := doJSON(t, srv, "POST", "/api/auth/register", "", domain.RegisterInput{Name: name, Email: emailAddr, Password: "password123"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("register %s: status %d body %s", emailAddr, rec.Code, rec.Body.String())
	}
	var res struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil || res.Token == "" {
		t.Fatalf("token kosong: %v body %s", err, rec.Body.String())
	}
	return res.Token
}

func magicTokenFromMail(t *testing.T, text string) string {
	t.Helper()
	idx := strings.Index(text, "/auth/verify?token=")
	if idx < 0 {
		t.Fatalf("magic link not found in email: %q", text)
	}
	token := text[idx+len("/auth/verify?token="):]
	return strings.Fields(token)[0]
}
