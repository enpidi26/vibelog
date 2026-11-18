package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionRejectsWeakSecret(t *testing.T) {
	base := Config{Env: "production", BaseURL: "https://vibelog.example.com", Port: "8080", DatabasePath: "x.db", EmailDriver: "console"}
	base.JWTSecret = "pendek"
	if err := base.Validate(); err == nil {
		t.Fatal("secret pendek harus ditolak di production")
	}
	base.JWTSecret = strings.Repeat("a", 32)
	if err := base.Validate(); err != nil {
		t.Fatalf("secret 32 karakter harus diterima: %v", err)
	}
}

func TestValidateRejectsBadBaseURL(t *testing.T) {
	cfg := Config{Env: "development", BaseURL: "bukan-url", Port: "8080", DatabasePath: "x.db", EmailDriver: "console"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("BASE_URL relatif harus ditolak")
	}
}

func TestValidateEmailDrivers(t *testing.T) {
	base := Config{Env: "development", BaseURL: "http://localhost:8080", Port: "8080", DatabasePath: "x.db"}

	base.EmailDriver = "smtp"
	if err := base.Validate(); err == nil {
		t.Fatal("smtp tanpa host harus ditolak")
	}
	base.SMTPHost = "smtp.example.com"
	base.SMTPTLS = "starttls"
	base.SMTPUsername = "user"
	base.SMTPPassword = "pass"
	if err := base.Validate(); err != nil {
		t.Fatalf("smtp valid ditolak: %v", err)
	}
	base.SMTPTLS = "none"
	if err := base.Validate(); err == nil {
		t.Fatal("kredensial SMTP tanpa TLS harus ditolak")
	}

	base = Config{Env: "development", BaseURL: "http://localhost:8080", Port: "8080", DatabasePath: "x.db", EmailDriver: "resend"}
	if err := base.Validate(); err == nil {
		t.Fatal("resend tanpa API key harus ditolak")
	}
}

func TestEnsureJWTSecretGeneratedAndPersisted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jwt_secret")
	a := Config{JWTSecretPath: path}
	if err := a.EnsureJWTSecret(); err != nil {
		t.Fatal(err)
	}
	if len(a.JWTSecret) < 32 {
		t.Fatalf("secret too short: %d", len(a.JWTSecret))
	}
	b := Config{JWTSecretPath: path}
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
