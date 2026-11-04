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
// maximum of 10 tags, masing-masing maksimal 30 karakter. Tag boleh berisi
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
		return nil, errors.New("maximum of 10 tags")
	}
	for _, t := range out {
		if len(t) > 30 {
			return nil, errors.New("each tag must be at most 30 characters")
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
		return errors.New("name is required, max 80 characters")
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" || len(email) > 120 || !strings.Contains(email, "@") {
		return errors.New("invalid email")
	}
	if len(in.Password) < 8 || len(in.Password) > 128 {
		return errors.New("password must be between 8 and 128 characters")
	}
	return nil
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (in LoginInput) Validate() error {
	if strings.TrimSpace(in.Email) == "" || strings.TrimSpace(in.Password) == "" {
		return errors.New("email and password are required")
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
		return errors.New("title is required, max 120 characters")
	}
	if strings.TrimSpace(in.PromptText) == "" || len(in.PromptText) > 10000 {
		return errors.New("prompt is required, max 10000 characters")
	}
	if strings.TrimSpace(in.Model) == "" || len(in.Model) > 60 {
		return errors.New("model is required, max 60 characters")
	}
	if strings.TrimSpace(in.Tool) == "" || len(in.Tool) > 60 {
		return errors.New("tool is required, max 60 characters")
	}
	if !IsValidOutcome(in.Outcome) {
		return errors.New("outcome must be success, partial, failed, or in_progress")
	}
	if in.Rating != nil && (*in.Rating < 1 || *in.Rating > 5) {
		return errors.New("rating must be between 1 and 5")
	}
	if len(in.ResultNotes) > 10000 {
		return errors.New("result notes must be at most 10000 characters")
	}
	if _, err := ParseTagsInput(in.Tags); err != nil {
		return err
	}
	return nil
}
