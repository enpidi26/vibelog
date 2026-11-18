// Package email mengirim pesan lewat console, SMTP, atau Resend.
package email

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"time"
)

// Mailer adalah satu-satunya jalan keluar email.
type Mailer interface {
	Send(to, subject, htmlBody, textBody string) error
	Name() string
}

var ErrDisabled = errors.New("email delivery is disabled on this instance")

// Config dibangun dari config.Config oleh caller supaya paket ini tidak
// bergantung pada paket config.
type Config struct {
	Driver       string
	From         string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPTLS      string
	ResendAPIKey string
}

func New(cfg Config) (Mailer, error) {
	from, err := mail.ParseAddress(cfg.From)
	if err != nil {
		return nil, fmt.Errorf("EMAIL_FROM is not a valid address: %w", err)
	}

	switch cfg.Driver {
	case "off":
		return offMailer{}, nil
	case "console", "":
		return consoleMailer{}, nil
	case "smtp":
		return &smtpMailer{
			host:     cfg.SMTPHost,
			port:     cfg.SMTPPort,
			username: cfg.SMTPUsername,
			password: cfg.SMTPPassword,
			tlsMode:  cfg.SMTPTLS,
			from:     from,
		}, nil
	case "resend":
		return &resendMailer{apiKey: cfg.ResendAPIKey, from: from}, nil
	default:
		return nil, fmt.Errorf("unknown EMAIL_DRIVER %q", cfg.Driver)
	}
}

type consoleMailer struct{}

func (consoleMailer) Name() string { return "console" }

func (consoleMailer) Send(to, subject, _, textBody string) error {
	log.Printf("[email:console] to=%s subject=%q\n%s", to, subject, textBody)
	return nil
}

type offMailer struct{}

func (offMailer) Name() string { return "off" }

func (offMailer) Send(_, _, _, _ string) error { return ErrDisabled }

// HTTPClient dipakai panggilan keluar supaya tidak menggantung selamanya bila
// layanan pihak ketiga lambat.
var HTTPClient = &http.Client{Timeout: 15 * time.Second}
