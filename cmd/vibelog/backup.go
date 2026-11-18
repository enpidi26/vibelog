package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/enpidi26/vibelog/internal/config"
	"github.com/enpidi26/vibelog/internal/database"
)

// runBackup membuat salinan database ke <dir>, default di sebelah database.
//
//	vibelog backup [direktori]
func runBackup(args []string) error {
	cfg := config.Load()

	dir := filepath.Join(filepath.Dir(cfg.DatabasePath), "backups")
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		dir = args[0]
	}
	keep := 7
	if v, err := strconv.Atoi(strings.TrimSpace(os.Getenv("BACKUP_KEEP"))); err == nil && v > 0 {
		keep = v
	}

	dest, err := database.Backup(cfg.DatabasePath, dir, keep)
	if err != nil {
		return err
	}
	fmt.Println(dest)
	return nil
}
