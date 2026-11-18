// Package domain berisi tipe data inti dan aturan validasi yang dipakai
// database maupun HTTP layer.
package domain

import "errors"

var (
	// ErrNotFound dikembalikan saat resource tidak ada atau bukan milik user.
	ErrNotFound = errors.New("not found")
	// ErrEmailTaken dikembalikan saat email sudah terdaftar.
	ErrEmailTaken = errors.New("email already registered")
)
