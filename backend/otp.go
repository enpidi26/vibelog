package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/smtp"
	"strings"
)

// GenerateOTP membuat kode OTP 6 digit secara acak.
func GenerateOTP() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// EmailConfig berisi konfigurasi SMTP untuk mengirim email.
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// SendOTPEmail mengirim email berisi kode OTP.
func SendOTPEmail(cfg EmailConfig, toEmail, toName, code string) error {
	if cfg.SMTPHost == "" {
		// Mode development: log OTP ke console saja
		fmt.Printf("\n=== OTP EMAIL ===\nTo: %s <%s>\nCode: %s\n=================\n", toName, toEmail, code)
		return nil
	}

	subject := "VibeLog Verification Code"
	body := fmt.Sprintf(`Hi %s,

Thanks for signing up for VibeLog!

Your verification code is: %s

This code is valid for 15 minutes. Do not share it with anyone.

If you didn't sign up for VibeLog, you can safely ignore this email.

Best,
The VibeLog Team`, toName, code)

	from := cfg.FromEmail
	if cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromEmail)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, toEmail, subject, body)

	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort)

	return smtp.SendMail(addr, auth, cfg.FromEmail, []string{toEmail}, []byte(msg))
}

// LoadEmailConfig membaca konfigurasi email dari environment.
func LoadEmailConfig() EmailConfig {
	return EmailConfig{
		SMTPHost:     strings.TrimSpace(envOr("SMTP_HOST", "")),
		SMTPPort:     envOr("SMTP_PORT", "587"),
		SMTPUsername: strings.TrimSpace(envOr("SMTP_USERNAME", "")),
		SMTPPassword: envOr("SMTP_PASSWORD", ""),
		FromEmail:    envOr("SMTP_FROM_EMAIL", "noreply@vibelog.app"),
		FromName:     envOr("SMTP_FROM_NAME", "VibeLog"),
	}
}
