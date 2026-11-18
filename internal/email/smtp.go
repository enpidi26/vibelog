package email

import (
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type smtpMailer struct {
	host     string
	port     int
	username string
	password string
	tlsMode  string
	from     *mail.Address
}

func (m *smtpMailer) Name() string { return "smtp" }

func (m *smtpMailer) Send(to, subject, htmlBody, _ string) error {
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
