package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config dibaca dari environment. Nilai default hanya aman untuk development
// lokal; mode production menolak secret default (lihat Validate).
type Config struct {
	Env               string
	Host              string
	Port              string
	DatabasePath      string
	JWTSecret         string
	JWTExpiryHrs      int
	FrontendURL       string
	AuthRateLimit      int
	APIRateLimit       int
	GitHubClientID     string
	GitHubClientSecret string
	ResendAPIKey       string
}

const devJWTSecret = "dev-only-secret-change-me-in-production"

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// LoadConfig membaca environment tanpa efek samping lain, supaya mudah diuji.
func LoadConfig() Config {
	expiry, err := strconv.Atoi(envOr("JWT_EXPIRY_HOURS", "24"))
	if err != nil || expiry < 1 || expiry > 720 {
		expiry = 24
	}
	return Config{
		Env:                envOr("APP_ENV", "development"),
		Host:               envOr("HOST", "127.0.0.1"),
		Port:               envOr("PORT", "8080"),
		DatabasePath:       envOr("DATABASE_PATH", "data/vibelog.db"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		JWTExpiryHrs:       expiry,
		FrontendURL:        envOr("FRONTEND_URL", "http://127.0.0.1:4173"),
		AuthRateLimit:      20,
		APIRateLimit:       300,
		GitHubClientID:     strings.TrimSpace(os.Getenv("GITHUB_CLIENT_ID")),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		ResendAPIKey:       strings.TrimSpace(os.Getenv("RESEND_API_KEY")),
	}
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
