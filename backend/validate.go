package main

import (
	"errors"
	"strings"
)

// Hasil (outcome) yang diizinkan untuk satu entri jurnal.
var validOutcomes = map[string]bool{
	"success":     true,
	"partial":     true,
	"failed":      true,
	"in_progress": true,
}

func IsValidOutcome(v string) bool { return validOutcomes[v] }

// NormalizeTags merapikan tag: trim, huruf kecil, buang duplikat dan kosong,
// maksimal 10 tag, masing-masing maksimal 30 karakter. Tag boleh berisi
// huruf, angka, spasi, strip, dan underscore; karakter lain dibuang.
func NormalizeTags(raw []string) ([]string, error) {
	seen := map[string]bool{}
	out := []string{}
	for _, t := range raw {
		cleaned := strings.ToLower(strings.TrimSpace(t))
		cleaned = strings.Join(strings.Fields(cleaned), " ")
		if cleaned == "" {
			continue
		}
		var b strings.Builder
		for _, r := range cleaned {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' || r == '_' {
				b.WriteRune(r)
			}
		}
		cleaned = strings.TrimSpace(b.String())
		if cleaned == "" || seen[cleaned] {
			continue
		}
		seen[cleaned] = true
		out = append(out, cleaned)
	}
	if len(out) > 10 {
		return nil, errors.New("maksimal 10 tag")
	}
	for _, t := range out {
		if len(t) > 30 {
			return nil, errors.New("setiap tag maksimal 30 karakter")
		}
	}
	return out, nil
}

// ParseTagsInput menerima tag dari string "a, b" atau array, dipakai handler
// supaya frontend boleh mengirim keduanya. Di API ini input resmi adalah
// array; string hanya diterima sebagai kemudahan dan dipecah koma.
func ParseTagsInput(raw []string) ([]string, error) {
	expanded := []string{}
	for _, t := range raw {
		for _, part := range strings.Split(t, ",") {
			expanded = append(expanded, part)
		}
	}
	return NormalizeTags(expanded)
}

type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (in RegisterInput) Validate() error {
	if strings.TrimSpace(in.Name) == "" || len(in.Name) > 80 {
		return errors.New("nama wajib diisi, maksimal 80 karakter")
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" || len(email) > 120 || !strings.Contains(email, "@") {
		return errors.New("email tidak valid")
	}
	if len(in.Password) < 8 || len(in.Password) > 128 {
		return errors.New("password minimal 8 karakter, maksimal 128 karakter")
	}
	return nil
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (in LoginInput) Validate() error {
	if strings.TrimSpace(in.Email) == "" || strings.TrimSpace(in.Password) == "" {
		return errors.New("email dan password wajib diisi")
	}
	return nil
}

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
		return errors.New("nama proyek wajib diisi, maksimal 80 karakter")
	}
	if len(in.Description) > 500 {
		return errors.New("deskripsi maksimal 500 karakter")
	}
	if in.Color != "" && !validColors[in.Color] {
		return errors.New("warna proyek tidak dikenal")
	}
	return nil
}

type EntryInput struct {
	ProjectID   *int64   `json:"project_id"`
	Title       string   `json:"title"`
	PromptText  string   `json:"prompt_text"`
	Model       string   `json:"model"`
	Tool        string   `json:"tool"`
	Outcome     string   `json:"outcome"`
	Rating      *int     `json:"rating"`
	ResultNotes string   `json:"result_notes"`
	Tags        []string `json:"tags"`
}

func (in EntryInput) Validate() error {
	if strings.TrimSpace(in.Title) == "" || len(in.Title) > 120 {
		return errors.New("judul wajib diisi, maksimal 120 karakter")
	}
	if strings.TrimSpace(in.PromptText) == "" || len(in.PromptText) > 10000 {
		return errors.New("prompt wajib diisi, maksimal 10000 karakter")
	}
	if strings.TrimSpace(in.Model) == "" || len(in.Model) > 60 {
		return errors.New("model wajib diisi, maksimal 60 karakter")
	}
	if strings.TrimSpace(in.Tool) == "" || len(in.Tool) > 60 {
		return errors.New("tool wajib diisi, maksimal 60 karakter")
	}
	if !IsValidOutcome(in.Outcome) {
		return errors.New("outcome harus success, partial, failed, atau in_progress")
	}
	if in.Rating != nil && (*in.Rating < 1 || *in.Rating > 5) {
		return errors.New("rating harus 1 sampai 5")
	}
	if len(in.ResultNotes) > 10000 {
		return errors.New("catatan hasil maksimal 10000 karakter")
	}
	if _, err := ParseTagsInput(in.Tags); err != nil {
		return err
	}
	return nil
}
