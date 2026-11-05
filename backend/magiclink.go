package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// sendMagicLink sends a magic link email via Resend API.
func sendMagicLink(apiKey, toEmail, magicURL string) error {
	if apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	payload := map[string]interface{}{
		"from":    "VibeLog <onboarding@resend.dev>",
		"to":      []string{toEmail},
		"subject": "Your VibeLog Login Link",
		"html": fmt.Sprintf(`
			<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
				<h2>Login to VibeLog</h2>
				<p>Click the button below to log in to your account. This link will expire in 15 minutes.</p>
				<a href="%s" style="display: inline-block; background: #0070f3; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0;">
					Login to VibeLog
				</a>
				<p style="color: #666; font-size: 14px;">If you didn't request this email, you can safely ignore it.</p>
				<p style="color: #666; font-size: 14px;">Or copy and paste this URL: <br><code>%s</code></p>
			</div>
		`, magicURL, magicURL),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", strings.NewReader(string(body)))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error: %d %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// generateSecureToken generates a cryptographically secure random token.
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// handleRequestMagicLink handles POST /api/auth/email/request
func (s *server) handleRequestMagicLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		writeError(w, http.StatusBadRequest, "invalid_email", "Invalid email address")
		return
	}

	// Check if user exists, if not create a new user
	var user User
	err := s.db.QueryRow(`SELECT id, name, email, email_verified FROM users WHERE email = ?`, email).
		Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified)

	if err == sql.ErrNoRows {
		// Create new user with email as name initially
		name := strings.Split(email, "@")[0]
		result, err := s.db.Exec(
			`INSERT INTO users (name, email, password_hash, email_verified, created_at) VALUES (?, ?, '', 0, ?)`,
			name, email, utcNow(),
		)
		if err != nil {
			fmt.Printf("create user for magic link: %v\n", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create account")
			return
		}
		userID, _ := result.LastInsertId()
		user.ID = userID
		user.Email = email
		user.Name = name
	} else if err != nil {
		fmt.Printf("lookup user for magic link: %v\n", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Database error")
		return
	}

	// Generate token
	token, err := generateSecureToken(32)
	if err != nil {
		fmt.Printf("generate magic link token: %v\n", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to generate token")
		return
	}

	// Store token (expires in 15 minutes)
	expiresAt := time.Now().UTC().Add(15 * time.Minute).Format(time.RFC3339Nano)
	_, err = s.db.Exec(
		`INSERT INTO otp_codes (user_id, code, purpose, expires_at, created_at) VALUES (?, ?, 'magic_link', ?, ?)`,
		user.ID, token, expiresAt, utcNow(),
	)
	if err != nil {
		fmt.Printf("store magic link token: %v\n", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to store token")
		return
	}

	// Send email
	magicURL := fmt.Sprintf("%s/auth/verify?token=%s", s.cfg.FrontendURL, token)
	if err := sendMagicLink(s.cfg.ResendAPIKey, email, magicURL); err != nil {
		fmt.Printf("send magic link email: %v\n", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to send email")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Magic link sent to your email",
	})
}

// handleVerifyMagicLink handles GET /api/auth/email/verify?token=xxx
func (s *server) handleVerifyMagicLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "missing_token", "Token is required")
		return
	}

	// Verify token
	var userID int64
	var expiresAt, used string
	err := s.db.QueryRow(
		`SELECT user_id, expires_at, used FROM otp_codes 
		 WHERE code = ? AND purpose = 'magic_link' 
		 ORDER BY created_at DESC LIMIT 1`,
		token,
	).Scan(&userID, &expiresAt, &used)

	if err == sql.ErrNoRows {
		writeError(w, http.StatusBadRequest, "invalid_token", "Invalid or expired token")
		return
	}
	if err != nil {
		fmt.Printf("verify magic link token: %v\n", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Database error")
		return
	}

	// Check if already used
	if used != "0" {
		writeError(w, http.StatusBadRequest, "token_used", "Token has already been used")
		return
	}

	// Check expiration
	expiry := parseTime(expiresAt)
	if time.Now().UTC().After(expiry) {
		writeError(w, http.StatusBadRequest, "token_expired", "Token has expired")
		return
	}

	// Mark token as used
	_, err = s.db.Exec(`UPDATE otp_codes SET used = 1 WHERE code = ? AND purpose = 'magic_link'`, token)
	if err != nil {
		fmt.Printf("mark magic link token as used: %v\n", err)
	}

	// Mark email as verified
	now := utcNow()
	_, err = s.db.Exec(`UPDATE users SET email_verified = 1, email_verified_at = ? WHERE id = ?`, now, userID)
	if err != nil {
		fmt.Printf("mark email verified: %v\n", err)
	}

	// Get user details
	var user User
	err = s.db.QueryRow(`SELECT id, name, email FROM users WHERE id = ?`, userID).
		Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		fmt.Printf("fetch user after magic link verify: %v\n", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch user")
		return
	}

	// Generate JWT
	jwtToken, err := IssueToken(s.cfg.EffectiveJWTSecret(), user.ID, user.Email, user.Name, s.cfg.JWTExpiryHrs)
	if err != nil {
		fmt.Printf("generate JWT after magic link: %v\n", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": jwtToken,
		"user":  publicUser(user),
	})
}
