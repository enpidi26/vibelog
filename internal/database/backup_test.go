package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupCreatesConsistentCopy(t *testing.T) {
	db, path := testDB(t)
	testUser(t, db, "ana@test.id")

	dir := filepath.Join(t.TempDir(), "backups")
	dest, err := Backup(path, dir, 7)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("backup file must not be empty")
	}

	// Backup harus bisa dibuka dan memuat user.
	backupDB, err := Open(dest)
	if err != nil {
		t.Fatalf("open backup: %v", err)
	}
	defer backupDB.Close()
	var count int
	if err := backupDB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("backup should contain 1 user, got %d", count)
	}
}

func TestRotateBackupsKeepsNewest(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"vibelog-20260101-000000.db", "vibelog-20260102-000000.db", "vibelog-20260103-000000.db"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := rotateBackups(dir, 2); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 backups, got %d", len(entries))
	}
	if entries[0].Name() != "vibelog-20260102-000000.db" || entries[1].Name() != "vibelog-20260103-000000.db" {
		t.Fatalf("wrong backups kept: %v %v", entries[0].Name(), entries[1].Name())
	}
}
