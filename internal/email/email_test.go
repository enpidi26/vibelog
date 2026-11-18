package email

import (
	"net/mail"
	"strings"
	"testing"
)

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

func TestNewMailers(t *testing.T) {
	from := "VibeLog <no-reply@example.com>"

	m, err := New(Config{Driver: "console", From: from})
	if err != nil || m.Name() != "console" {
		t.Fatalf("console: %v %v", m, err)
	}
	m, err = New(Config{Driver: "off", From: from})
	if err != nil || m.Name() != "off" {
		t.Fatalf("off: %v %v", m, err)
	}
	m, err = New(Config{Driver: "smtp", From: from, SMTPHost: "smtp.example.com", SMTPPort: 587, SMTPTLS: "starttls"})
	if err != nil || m.Name() != "smtp" {
		t.Fatalf("smtp: %v %v", m, err)
	}
	if _, err := New(Config{Driver: "smtp", From: "not-an-address"}); err == nil {
		t.Fatal("EMAIL_FROM tidak valid harus ditolak")
	}
	if _, err := New(Config{Driver: "unknown", From: from}); err == nil {
		t.Fatal("driver tidak dikenal harus ditolak")
	}
}

func TestOffMailerErrors(t *testing.T) {
	if err := (offMailer{}).Send("a@b.test", "s", "h", "t"); err != ErrDisabled {
		t.Fatalf("want ErrDisabled, got %v", err)
	}
}
