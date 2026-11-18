package domain

import (
	"errors"
	"strings"
	"time"
)

// Project mengelompokkan entri jurnal.
type Project struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"-"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	EntryCount  int       `json:"entry_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProjectInput adalah payload pembuatan/pengubahan proyek.
type ProjectInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

var validColors = map[string]bool{
	"teal": true, "amber": true, "violet": true, "rose": true,
	"sky": true, "lime": true, "slate": true,
}

func (in ProjectInput) Validate() error {
	if strings.TrimSpace(in.Name) == "" || len(in.Name) > 80 {
		return errors.New("project name is required, max 80 characters")
	}
	if len(in.Description) > 500 {
		return errors.New("description must be at most 500 characters")
	}
	if in.Color != "" && !validColors[in.Color] {
		return errors.New("unknown project color")
	}
	return nil
}
