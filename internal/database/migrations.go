package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Setiap perubahan skema adalah satu migrasi bernomor yang dijalankan sekali
// dan dicatat di schema_migrations. Database self-host hidup bertahun-tahun
// dan harus bisa naik versi tanpa kehilangan data.
type migration struct {
	version int
	name    string
	up      func(*sql.Tx) error
}

var migrations = []migration{
	{version: 1, name: "initial schema", up: migrateInitialSchema},
	{version: 2, name: "email verification columns", up: migrateEmailVerifiedColumns},
	{version: 3, name: "otp lookup index", up: func(tx *sql.Tx) error {
		_, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_otp_code ON otp_codes(code, purpose)`)
		return err
	}},
}

func migrate(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TEXT NOT NULL
		)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, m := range migrations {
		var applied int
		err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, m.version).Scan(&applied)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", m.version, err)
		}
		if applied > 0 {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", m.version, err)
		}
		if err := m.up(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d (%s): %w", m.version, m.name, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)`,
			m.version, m.name, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.version, err)
		}
	}
	return nil
}

func migrateInitialSchema(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			email_verified INTEGER NOT NULL DEFAULT 0,
			email_verified_at TEXT,
			created_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			color TEXT NOT NULL DEFAULT 'teal',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			project_id INTEGER REFERENCES projects(id) ON DELETE SET NULL,
			title TEXT NOT NULL,
			prompt_text TEXT NOT NULL,
			model TEXT NOT NULL,
			tool TEXT NOT NULL,
			outcome TEXT NOT NULL,
			rating INTEGER,
			result_notes TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '',
			iteration_no INTEGER NOT NULL DEFAULT 1,
			parent_id INTEGER REFERENCES entries(id) ON DELETE SET NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS otp_codes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			code TEXT NOT NULL,
			purpose TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			used INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_projects_user ON projects(user_id, name);
		CREATE INDEX IF NOT EXISTS idx_entries_user_created ON entries(user_id, created_at DESC, id DESC);
		CREATE INDEX IF NOT EXISTS idx_entries_user_project ON entries(user_id, project_id);
		CREATE INDEX IF NOT EXISTS idx_entries_user_outcome ON entries(user_id, outcome);
		CREATE INDEX IF NOT EXISTS idx_otp_user_purpose ON otp_codes(user_id, purpose, expires_at);
	`)
	return err
}

// Database lama (sebelum fitur verifikasi email) belum punya dua kolom ini.
func migrateEmailVerifiedColumns(tx *sql.Tx) error {
	cols, err := tableColumns(tx, "users")
	if err != nil {
		return err
	}
	if !cols["email_verified"] {
		if _, err := tx.Exec(`ALTER TABLE users ADD COLUMN email_verified INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}
	if !cols["email_verified_at"] {
		if _, err := tx.Exec(`ALTER TABLE users ADD COLUMN email_verified_at TEXT`); err != nil {
			return err
		}
	}
	return nil
}

func tableColumns(tx *sql.Tx, table string) (map[string]bool, error) {
	rows, err := tx.Query(`SELECT name FROM pragma_table_info(?)`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		cols[name] = true
	}
	return cols, rows.Err()
}
