package database

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrTokenInvalid = errors.New("invalid or expired token")
	ErrTokenUsed    = errors.New("token already used")
	ErrTokenExpired = errors.New("token expired")
)

// IssueMagicToken menyimpan token magic link dan membatalkan token lama user
// tersebut, sehingga hanya tautan terbaru yang berlaku.
func IssueMagicToken(db *sql.DB, userID int64, token string, ttl time.Duration) error {
	if _, err := db.Exec(`DELETE FROM otp_codes WHERE user_id = ? AND purpose = 'magic_link'`, userID); err != nil {
		return err
	}
	expiresAt := time.Now().UTC().Add(ttl).Format(time.RFC3339Nano)
	_, err := db.Exec(
		`INSERT INTO otp_codes (user_id, code, purpose, expires_at, created_at) VALUES (?, ?, 'magic_link', ?, ?)`,
		userID, token, expiresAt, utcNow(),
	)
	return err
}

// ConsumeMagicToken memakai token sekali pakai dan menandai email user
// terverifikasi dalam satu transaksi. ErrTokenInvalid/Used/Expired membedakan
// alasan penolakan untuk pesan API.
func ConsumeMagicToken(db *sql.DB, token string) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var otpID, userID int64
	var expiresAt string
	var used int
	err = tx.QueryRow(`SELECT id, user_id, expires_at, used FROM otp_codes
		WHERE code = ? AND purpose = 'magic_link'
		ORDER BY id DESC LIMIT 1`, token).
		Scan(&otpID, &userID, &expiresAt, &used)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrTokenInvalid
	}
	if err != nil {
		return 0, err
	}
	if used != 0 {
		return 0, ErrTokenUsed
	}
	if time.Now().UTC().After(parseTime(expiresAt)) {
		return 0, ErrTokenExpired
	}

	res, err := tx.Exec(`UPDATE otp_codes SET used = 1 WHERE id = ? AND used = 0`, otpID)
	if err != nil {
		return 0, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return 0, ErrTokenUsed
	}
	if _, err := tx.Exec(`UPDATE users SET email_verified = 1, email_verified_at = ? WHERE id = ?`, utcNow(), userID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return userID, nil
}

// CleanupExpiredTokens menghapus token yang sudah kadaluarsa lebih dari 24 jam.
func CleanupExpiredTokens(db *sql.DB) error {
	cutoff := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339Nano)
	_, err := db.Exec(`DELETE FROM otp_codes WHERE expires_at < ?`, cutoff)
	return err
}
