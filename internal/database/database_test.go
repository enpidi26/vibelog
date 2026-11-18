package database

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enpidi26/vibelog/internal/domain"
)

func testDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "vibelog-test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, path
}

func testUser(t *testing.T, db *sql.DB, email string) domain.User {
	t.Helper()
	u, err := CreateUser(db, "Test", email, "hash")
	if err != nil {
		t.Fatalf("CreateUser(%s): %v", email, err)
	}
	return u
}

func TestMigrationsAreVersionedAndIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "migrate.db")
	for i := 0; i < 2; i++ {
		db, err := Open(path)
		if err != nil {
			t.Fatalf("open %d: %v", i, err)
		}
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
			t.Fatalf("count migrations: %v", err)
		}
		if count != len(migrations) {
			t.Fatalf("expected %d migrations, got %d", len(migrations), count)
		}
		db.Close()
	}
}

func TestMigrationAddsEmailVerificationToOldDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")
	raw, err := sql.Open("sqlite", sqliteDSN(path))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`); err != nil {
		t.Fatal(err)
	}
	raw.Close()

	db, err := Open(path)
	if err != nil {
		t.Fatalf("upgrade old database: %v", err)
	}
	defer db.Close()
	cols, err := tableColumnsDB(db, "users")
	if err != nil {
		t.Fatal(err)
	}
	for _, col := range []string{"email_verified", "email_verified_at"} {
		if !cols[col] {
			t.Fatalf("users.%s missing after migration", col)
		}
	}
}

func tableColumnsDB(db *sql.DB, table string) (map[string]bool, error) {
	rows, err := db.Query(`SELECT name FROM pragma_table_info(?)`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out[name] = true
	}
	return out, rows.Err()
}

func TestDatabasePragmas(t *testing.T) {
	db, _ := testDB(t)

	var mode string
	if err := db.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if strings.ToLower(mode) != "wal" {
		t.Fatalf("journal_mode = %q, want wal", mode)
	}
	var fk int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatal(err)
	}
	if fk != 1 {
		t.Fatalf("foreign_keys = %d, want 1", fk)
	}
	var busy int
	if err := db.QueryRow(`PRAGMA busy_timeout`).Scan(&busy); err != nil {
		t.Fatal(err)
	}
	if busy != 5000 {
		t.Fatalf("busy_timeout = %d, want 5000", busy)
	}
}

func TestCreateUserRejectsDuplicateEmail(t *testing.T) {
	db, _ := testDB(t)
	testUser(t, db, "ana@test.id")
	if _, err := CreateUser(db, "Ana 2", "ANA@test.id", "hash"); err != domain.ErrEmailTaken {
		t.Fatalf("want ErrEmailTaken, got %v", err)
	}
}

func TestMarkAndDeleteUser(t *testing.T) {
	db, _ := testDB(t)
	u := testUser(t, db, "ana@test.id")
	if err := MarkEmailVerified(db, u.ID); err != nil {
		t.Fatal(err)
	}
	got, err := GetUserByID(db, u.ID)
	if err != nil || !got.EmailVerified || got.EmailVerifiedAt == nil {
		t.Fatalf("user should be verified: %+v err=%v", got, err)
	}
	if err := DeleteUser(db, u.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := GetUserByID(db, u.ID); err != domain.ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
