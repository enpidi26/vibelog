package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// runBackup membuat salinan konsisten database dengan VACUUM INTO, lalu
// menyisakan sejumlah arsip terbaru. Aman dijalankan saat server hidup.
//
//	vibelog backup [direktori]
func runBackup(args []string) error {
	cfg := LoadConfig()

	if _, err := os.Stat(cfg.DatabasePath); err != nil {
		return fmt.Errorf("database not found at %s: %w", cfg.DatabasePath, err)
	}

	dir := filepath.Join(filepath.Dir(cfg.DatabasePath), "backups")
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		dir = args[0]
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create backup directory: %w", err)
	}

	dest := filepath.Join(dir, "vibelog-"+time.Now().Format("20060102-150405")+".db")
	db, err := sql.Open("sqlite3", cfg.DatabasePath+"?_busy_timeout=5000")
	if err != nil {
		return err
	}
	defer db.Close()

	quoted := strings.ReplaceAll(filepath.ToSlash(dest), "'", "''")
	if _, err := db.Exec(fmt.Sprintf("VACUUM INTO '%s'", quoted)); err != nil {
		return fmt.Errorf("vacuum into backup: %w", err)
	}
	if err := os.Chmod(dest, 0o600); err != nil {
		return fmt.Errorf("set backup permissions: %w", err)
	}

	keep := 7
	if v, err := strconv.Atoi(strings.TrimSpace(os.Getenv("BACKUP_KEEP"))); err == nil && v > 0 {
		keep = v
	}
	if err := rotateBackups(dir, keep); err != nil {
		return err
	}

	fmt.Println(dest)
	return nil
}

func rotateBackups(dir string, keep int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	names := []string{}
	for _, e := range entries {
		n := e.Name()
		if !e.IsDir() && strings.HasPrefix(n, "vibelog-") && strings.HasSuffix(n, ".db") {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	for len(names) > keep {
		oldest := filepath.Join(dir, names[0])
		if err := os.Remove(oldest); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove old backup %s: %w", oldest, err)
		}
		names = names[1:]
	}
	return nil
}
