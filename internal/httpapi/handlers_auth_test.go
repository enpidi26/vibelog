package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/enpidi26/vibelog/internal/domain"
)

func TestAuthDuplicateEmailAndWrongPassword(t *testing.T) {
	srv, _ := testServer(t, nil)
	registerToken(t, srv, "Ana", "ana@test.id")

	rec := doJSON(t, srv, "POST", "/api/auth/register", "", domain.RegisterInput{Name: "Ana 2", Email: "ANA@test.id", Password: "password123"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("email duplikat harus 409, dapat %d: %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, srv, "POST", "/api/auth/login", "", domain.LoginInput{Email: "ana@test.id", Password: "salah1234"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("password salah harus 401, dapat %d", rec.Code)
	}
	rec = doJSON(t, srv, "GET", "/api/entries", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("tanpa token harus 401, dapat %d", rec.Code)
	}

	rec = doJSON(t, srv, "POST", "/api/auth/login", "", domain.LoginInput{Email: "ana@test.id", Password: "password123"})
	if rec.Code != http.StatusOK {
		t.Fatalf("login benar harus 200, dapat %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMeEndpoint(t *testing.T) {
	srv, _ := testServer(t, nil)
	token := registerToken(t, srv, "Ana", "ana@test.id")
	rec := doJSON(t, srv, "GET", "/api/auth/me", token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("me: %d %s", rec.Code, rec.Body.String())
	}
	var user map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &user); err != nil {
		t.Fatal(err)
	}
	if user["email"] != "ana@test.id" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestMetaReflectsEnabledFeatures(t *testing.T) {
	srv, _ := testServer(t, nil)

	rec := doJSON(t, srv, "GET", "/api/meta", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("meta: %d %s", rec.Code, rec.Body.String())
	}
	var meta map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &meta); err != nil {
		t.Fatal(err)
	}
	if meta["email_login_enabled"] != true || meta["registration_enabled"] != true || meta["github_login_enabled"] != false || meta["email_delivery"] != "capture" {
		t.Fatalf("unexpected meta: %+v", meta)
	}

	srv.cfg.AllowRegistration = false
	srv.mailer = offTestMailer{}
	rec = doJSON(t, srv, "GET", "/api/meta", "", nil)
	_ = json.Unmarshal(rec.Body.Bytes(), &meta)
	if meta["email_login_enabled"] != false || meta["registration_enabled"] != false || meta["email_delivery"] != "off" {
		t.Fatalf("features should be disabled: %+v", meta)
	}
}
