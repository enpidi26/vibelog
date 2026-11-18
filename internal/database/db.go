// Package database menyimpan seluruh akses SQLite VibeLog.
package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/enpidi26/vibelog/internal/domain"

	_ "modernc.org/sqlite"
)

// sqliteDSN memakai driver SQLite murni-Go. Tanpa CGO, binary bisa
// di-cross-compile dan tidak butuh gcc di mesin self-host.
func sqliteDSN(path string) string {
	return "file:" + filepath.ToSlash(path) + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"
}

// Open membuka SQLite, mengaktifkan WAL + foreign keys, dan menjalankan
// migrasi.
func Open(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", sqliteDSN(path))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func utcNow() string { return time.Now().UTC().Format(time.RFC3339Nano) }

// escapeLike meng-escape karakter khusus LIKE supaya pencarian "%", "_"
// atau "\" tidak berubah menjadi wildcard.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func likePattern(s string) string {
	return "%" + escapeLike(strings.ToLower(s)) + "%"
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t.UTC()
}

func tagsToText(tags []string) string { return strings.Join(tags, ", ") }

func textToTags(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	tags, _ := domain.NormalizeTags(strings.Split(s, ","))
	if tags == nil {
		return []string{}
	}
	return tags
}

func nullableInt64(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	n := v.Int64
	return &n
}

func nullableInt(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	n := int(v.Int64)
	return &n
}
