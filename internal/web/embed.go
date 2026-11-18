// Package web membawa hasil build frontend ke dalam binary Go, supaya
// self-host cukup menjalankan satu container dan satu port.
package web

import (
	"embed"
	"io/fs"
)

// Files berisi frontend/dist. Saat build lokal tanpa frontend, hanya .gitkeep
// yang ada dan server tidak menyalakan mode SPA.
//
//go:embed all:dist
var Files embed.FS

// FS mengembalikan isi dist dan apakah index.html benar-benar ada.
func FS() (fs.FS, bool) {
	sub, err := fs.Sub(Files, "dist")
	if err != nil {
		return nil, false
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return nil, false
	}
	return sub, true
}
