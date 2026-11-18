package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// Mailer adalah satu-satunya jalan keluar email. Self-host bisa memakai
// console (link di log), SMTP milik sendiri, atau Resend untuk cloud.
type Mailer interface {
	Send(to, subject, htmlBody, textBody string) error
	Name() string
}

var ErrEmailDisabled = errors.New("email delivery is disabled on this instance")

func NewMailer(cfg Config) (Mailer, error) {
	from, err := mail.ParseAddress(cfg.EmailFrom)
	if err != nil {
		return nil, fmt.Errorf("EMAIL_FROM is not a valid address: %w", err)
	}

	switch cfg.EmailDriver {
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
		return &resendMailer{
			apiKey: cfg.ResendAPIKey,
			from:   from,
		}, nil
	default:
		return nil, fmt.Errorf("unknown EMAIL_DRIVER %q", cfg.EmailDriver)
	}
}

// ---- console ----

type consoleMailer struct{}

func (consoleMailer) Name() string { return "console" }

func (consoleMailer) Send(to, subject, _, textBody string) error {
	log.Printf("[email:console] to=%s subject=%q\n%s", to, subject, textBody)
	return nil
}

// ---- off ----

type offMailer struct{}

func (offMailer) Name() string { return "off" }

func (offMailer) Send(_, _, _, _ string) error { return ErrEmailDisabled }

// ---- SMTP ----

type smtpMailer struct {
	host     string
	port     int
	username string
	password string
	tlsMode  string
	from     *mail.Address
}

func (m *smtpMailer) Name() string { return "smtp" }

func (m *smtpMailer) Send(to, subject, htmlBody, textBody string) error {
	addr := net.JoinHostPort(m.host, strconv.Itoa(m.port))
	tlsConfig := &tls.Config{ServerName: m.host, MinVersion: tls.VersionTLS12}

	var client *smtp.Client
	var err error
	switch m.tlsMode {
	case "implicit":
		conn, derr := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", addr, tlsConfig)
		if derr != nil {
			return fmt.Errorf("smtp tls dial: %w", derr)
		}
		client, err = smtp.NewClient(conn, m.host)
	default:
		client, err = smtp.Dial(addr)
		if err == nil && m.tlsMode == "starttls" {
			if ok, _ := client.Extension("STARTTLS"); !ok {
				client.Close()
				return errors.New("smtp server does not offer STARTTLS")
			}
			err = client.StartTLS(tlsConfig)
		}
	}
	if err != nil {
		return fmt.Errorf("smtp connect: %w", err)
	}
	defer client.Close()

	if m.username != "" {
		if ok, _ := client.Extension("AUTH"); !ok {
			return errors.New("smtp server does not offer AUTH")
		}
		auth := smtp.PlainAuth("", m.username, m.password, m.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(m.from.Address); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(buildMIMEMessage(m.from, to, subject, htmlBody)); err != nil {
		return fmt.Errorf("smtp write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp finish body: %w", err)
	}
	return client.Quit()
}

func buildMIMEMessage(from *mail.Address, to, subject, htmlBody string) []byte {
	var body strings.Builder
	qp := quotedprintable.NewWriter(&body)
	_, _ = qp.Write([]byte(htmlBody))
	_ = qp.Close()

	headers := []string{
		"From: " + from.String(),
		"To: " + to,
		"Subject: " + mime.QEncoding.Encode("utf-8", subject),
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		`Content-Type: text/html; charset="utf-8"`,
		"Content-Transfer-Encoding: quoted-printable",
	}
	// quotedprintable.Writer sudah memakai CRLF; jangan dikonversi dua kali.
	return []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + body.String())
}

// ---- Resend ----

type resendMailer struct {
	apiKey string
	from   *mail.Address
}

func (m *resendMailer) Name() string { return "resend" }

func (m *resendMailer) Send(to, subject, htmlBody, _ string) error {
	payload := map[string]any{
		"from":    m.from.String(),
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return fmt.Errorf("resend API error: %d %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}

// httpClient dipakai semua panggilan keluar supaya tidak ada request yang
// menggantung selamanya bila layanan pihak ketiga lambat.
var httpClient = &http.Client{Timeout: 15 * time.Second}
