package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config dibaca dari environment. Nilai default hanya aman untuk development
// lokal; mode production menolak secret default (lihat Validate).
type Config struct {
	Env          string
	Host         string
	Port         string
	DatabasePath string

	JWTSecret     string
	JWTSecretPath string
	JWTExpiryHrs  int
	BaseURL       string

	AuthRateLimit int
	APIRateLimit  int

	AllowRegistration bool

	EmailDriver string
	EmailFrom   string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPTLS      string // starttls | implicit | none

	ResendAPIKey string

	GitHubClientID     string
	GitHubClientSecret string

	TrustedProxies []*net.IPNet
}

const devJWTSecret = "dev-only-secret-change-me-in-production"

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	return fallback
}

// LoadConfig membaca environment tanpa efek samping lain, supaya mudah diuji.
// Resolusi secret dari file dilakukan terpisah lewat EnsureJWTSecret.
func LoadConfig() Config {
	port := envOr("PORT", "8080")

	expiry, err := strconv.Atoi(envOr("JWT_EXPIRY_HOURS", "24"))
	if err != nil || expiry < 1 || expiry > 720 {
		expiry = 24
	}

	smtpPort, err := strconv.Atoi(envOr("SMTP_PORT", "587"))
	if err != nil || smtpPort < 1 || smtpPort > 65535 {
		smtpPort = 587
	}

	// BASE_URL adalah nama baru; FRONTEND_URL tetap dibaca supaya instalasi
	// lama tidak putus saat upgrade.
	baseURL := envOr("BASE_URL", envOr("FRONTEND_URL", "http://localhost:"+port))
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")

	dbPath := envOr("DATABASE_PATH", "data/vibelog.db")

	return Config{
		Env:                envOr("APP_ENV", "development"),
		Host:               envOr("HOST", "127.0.0.1"),
		Port:               port,
		DatabasePath:       dbPath,
		JWTSecret:          strings.TrimSpace(os.Getenv("JWT_SECRET")),
		JWTSecretPath:      envOr("JWT_SECRET_PATH", filepath.Join(filepath.Dir(dbPath), "jwt_secret")),
		JWTExpiryHrs:       expiry,
		BaseURL:            baseURL,
		AuthRateLimit:      20,
		APIRateLimit:       300,
		AllowRegistration:  envBool("ALLOW_REGISTRATION", true),
		EmailDriver:        strings.ToLower(envOr("EMAIL_DRIVER", "console")),
		EmailFrom:          envOr("EMAIL_FROM", "VibeLog <no-reply@localhost>"),
		SMTPHost:           strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:           smtpPort,
		SMTPUsername:       os.Getenv("SMTP_USERNAME"),
		SMTPPassword:       os.Getenv("SMTP_PASSWORD"),
		SMTPTLS:            strings.ToLower(envOr("SMTP_TLS", defaultSMTPTLS(smtpPort))),
		ResendAPIKey:       strings.TrimSpace(os.Getenv("RESEND_API_KEY")),
		GitHubClientID:     strings.TrimSpace(os.Getenv("GITHUB_CLIENT_ID")),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		TrustedProxies:     ParseTrustedProxies(envOr("TRUSTED_PROXIES", "127.0.0.1/32,::1/128")),
	}
}

func defaultSMTPTLS(port int) string {
	if port == 465 {
		return "implicit"
	}
	return "starttls"
}

// EnsureJWTSecret memakai JWT_SECRET bila diisi. Kalau kosong, secret dibuat
// sekali dan disimpan di JWT_SECRET_PATH (default: sebelah file database)
// dengan mode 0600. Self-hoster tidak perlu menyetel secret apa pun, dan
// secret tetap stabil antar restart.
func (c *Config) EnsureJWTSecret() error {
	if strings.TrimSpace(c.JWTSecret) != "" {
		return nil
	}
	if c.JWTSecretPath == "" {
		return errors.New("JWT_SECRET is empty and JWT_SECRET_PATH is not set")
	}

	if b, err := os.ReadFile(c.JWTSecretPath); err == nil {
		secret := strings.TrimSpace(string(b))
		if len(secret) >= 32 {
			c.JWTSecret = secret
			return nil
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read jwt secret: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(c.JWTSecretPath), 0o700); err != nil {
		return fmt.Errorf("create secret directory: %w", err)
	}
	buf := make([]byte, 48)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Errorf("generate jwt secret: %w", err)
	}
	secret := hex.EncodeToString(buf)
	if err := os.WriteFile(c.JWTSecretPath, []byte(secret+"\n"), 0o600); err != nil {
		return fmt.Errorf("write jwt secret: %w", err)
	}
	c.JWTSecret = secret
	return nil
}

// Validate menolak konfigurasi yang tidak aman untuk production sebelum
// server mendengarkan port apa pun.
func (c Config) Validate() error {
	if c.Env == "production" {
		if c.JWTSecret == "" {
			return errors.New("JWT_SECRET must be set in production mode")
		}
		if c.JWTSecret == devJWTSecret {
			return errors.New("JWT_SECRET is still using the development default value")
		}
		if len(c.JWTSecret) < 32 {
			return errors.New("JWT_SECRET must be at least 32 characters in production mode")
		}
	}
	if c.Port == "" {
		return errors.New("PORT cannot be empty")
	}
	if c.DatabasePath == "" {
		return errors.New("DATABASE_PATH cannot be empty")
	}
	if c.BaseURL == "" {
		return errors.New("BASE_URL cannot be empty")
	}
	if u, err := url.Parse(c.BaseURL); err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("BASE_URL must be an absolute URL, got %q", c.BaseURL)
	}
	switch c.EmailDriver {
	case "console", "smtp", "resend", "off":
	case "":
		return errors.New("EMAIL_DRIVER cannot be empty")
	default:
		return fmt.Errorf("EMAIL_DRIVER must be console, smtp, resend or off, got %q", c.EmailDriver)
	}
	if c.EmailDriver == "smtp" {
		if c.SMTPHost == "" {
			return errors.New("SMTP_HOST is required when EMAIL_DRIVER=smtp")
		}
		switch c.SMTPTLS {
		case "starttls", "implicit", "none":
		default:
			return errors.New("SMTP_TLS must be starttls, implicit or none")
		}
		if (c.SMTPUsername != "" || c.SMTPPassword != "") && c.SMTPTLS == "none" {
			return errors.New("SMTP credentials require SMTP_TLS=starttls or implicit")
		}
	}
	if c.EmailDriver == "resend" && c.ResendAPIKey == "" {
		return errors.New("RESEND_API_KEY is required when EMAIL_DRIVER=resend")
	}
	return nil
}

// EffectiveJWTSecret mengembalikan secret yang dipakai; di development,
// secret default dipakai supaya ./scripts/run.sh jalan tanpa setup.
func (c Config) EffectiveJWTSecret() string {
	if c.JWTSecret != "" {
		return c.JWTSecret
	}
	return devJWTSecret
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// BaseURLIsHTTPS dipakai untuk menandai cookie Secure.
func (c Config) BaseURLIsHTTPS() bool {
	return strings.HasPrefix(strings.ToLower(c.BaseURL), "https://")
}
