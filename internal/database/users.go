package database

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/enpidi26/vibelog/internal/domain"
)

func CreateUser(db *sql.DB, name, email, passwordHash string) (domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	res, err := db.Exec(`INSERT INTO users (name, email, password_hash, email_verified, created_at) VALUES (?, ?, ?, 0, ?)`,
		strings.TrimSpace(name), email, passwordHash, utcNow())
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.User{}, domain.ErrEmailTaken
		}
		return domain.User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.User{}, err
	}
	return GetUserByID(db, id)
}

func GetUserByID(db *sql.DB, id int64) (domain.User, error) {
	var u domain.User
	var created string
	var verified int
	var verifiedAt sql.NullString
	err := db.QueryRow(`SELECT id, name, email, password_hash, email_verified, email_verified_at, created_at FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &verified, &verifiedAt, &created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	u.EmailVerified = verified == 1
	if verifiedAt.Valid {
		t := parseTime(verifiedAt.String)
		u.EmailVerifiedAt = &t
	}
	u.CreatedAt = parseTime(created)
	return u, nil
}

func GetUserByEmail(db *sql.DB, email string) (domain.User, error) {
	var u domain.User
	var created string
	var verified int
	var verifiedAt sql.NullString
	err := db.QueryRow(`SELECT id, name, email, password_hash, email_verified, email_verified_at, created_at FROM users WHERE email = ?`,
		strings.ToLower(strings.TrimSpace(email))).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &verified, &verifiedAt, &created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	u.EmailVerified = verified == 1
	if verifiedAt.Valid {
		t := parseTime(verifiedAt.String)
		u.EmailVerifiedAt = &t
	}
	u.CreatedAt = parseTime(created)
	return u, nil
}

// MarkEmailVerified menandai email user sebagai terverifikasi.
func MarkEmailVerified(db *sql.DB, userID int64) error {
	_, err := db.Exec(`UPDATE users SET email_verified = 1, email_verified_at = ? WHERE id = ?`, utcNow(), userID)
	return err
}

// DeleteUser dipakai untuk membatalkan akun yang baru dibuat ketika email
// magic link gagal dikirim.
func DeleteUser(db *sql.DB, userID int64) error {
	_, err := db.Exec(`DELETE FROM users WHERE id = ?`, userID)
	return err
}
