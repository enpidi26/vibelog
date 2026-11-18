package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/enpidi26/vibelog/internal/config"
	"github.com/enpidi26/vibelog/internal/database"
	"github.com/enpidi26/vibelog/internal/domain"
)

func TestRegistrationCanBeDisabled(t *testing.T) {
	srv, mailer := testServer(t, func(cfg *config.Config) {
		cfg.AllowRegistration = false
	})

	rec := doJSON(t, srv, "POST", "/api/auth/register", "", domain.RegisterInput{
		Name: "Ana", Email: "ana@test.id", Password: "password123",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("registration should be 403, got %d: %s", rec.Code, rec.Body.String())
	}

	// Email tak dikenal tidak boleh membuat akun, dan respons harus generik.
	rec = doJSON(t, srv, "POST", "/api/auth/email/request", "", map[string]string{"email": "baru@test.id"})
	if rec.Code != http.StatusOK {
		t.Fatalf("magic request should be generic 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if mailer.to != "" {
		t.Fatal("no email should be sent when registration is closed")
	}
	if _, err := database.GetUserByEmail(srv.db, "baru@test.id"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("user must not be created, got err=%v", err)
	}
}

func TestMagicLinkRequestVerifyAndReplay(t *testing.T) {
	srv, mailer := testServer(t, nil)

	rec := doJSON(t, srv, "POST", "/api/auth/email/request", "", map[string]string{"email": "Magic@test.id"})
	if rec.Code != http.StatusOK {
		t.Fatalf("request magic link: %d %s", rec.Code, rec.Body.String())
	}
	if mailer.to != "magic@test.id" {
		t.Fatalf("email should be normalized, got %q", mailer.to)
	}
	token := magicTokenFromMail(t, mailer.text)

	// GET tidak boleh membakar token.
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
	srv, mailer := testServer(t, nil)

	rec := doJSON(t, srv, "POST", "/api/auth/email/request", "", map[string]string{"email": "magic@test.id"})
	if rec.Code != http.StatusOK {
		t.Fatalf("first request: %d", rec.Code)
	}
	first := magicTokenFromMail(t, mailer.text)

	rec = doJSON(t, srv, "POST", "/api/auth/email/request", "", map[string]string{"email": "magic@test.id"})
	if rec.Code != http.StatusOK {
		t.Fatalf("second request: %d", rec.Code)
	}
	second := magicTokenFromMail(t, mailer.text)
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

func TestEmailDisabledRejectsMagicLink(t *testing.T) {
	srv, _ := testServer(t, nil)
	srv.mailer = offTestMailer{}
	rec := doJSON(t, srv, "POST", "/api/auth/email/request", "", map[string]string{"email": "a@b.test"})
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("email off should be 503, got %d: %s", rec.Code, rec.Body.String())
	}
}
