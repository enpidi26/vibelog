package domain

import (
	"errors"
	"strings"
	"time"
)

// Outcome yang diizinkan untuk satu entri jurnal.
const (
	OutcomeSuccess    = "success"
	OutcomePartial    = "partial"
	OutcomeFailed     = "failed"
	OutcomeInProgress = "in_progress"
)

var validOutcomes = map[string]bool{
	OutcomeSuccess:    true,
	OutcomePartial:    true,
	OutcomeFailed:     true,
	OutcomeInProgress: true,
}

func IsValidOutcome(v string) bool { return validOutcomes[v] }

// Entry adalah satu catatan prompt beserta hasil dan iterasinya.
type Entry struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"-"`
	ProjectID   *int64    `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Title       string    `json:"title"`
	PromptText  string    `json:"prompt_text"`
	Model       string    `json:"model"`
	Tool        string    `json:"tool"`
	Outcome     string    `json:"outcome"`
	Rating      *int      `json:"rating"`
	ResultNotes string    `json:"result_notes"`
	Tags        []string  `json:"tags"`
	IterationNo int       `json:"iteration_no"`
	ParentID    *int64    `json:"parent_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EntryFilter membatasi daftar entri. All=true dipakai ekspor supaya seluruh
// riwayat ikut, bukan hanya halaman pertama.
type EntryFilter struct {
	Query     string
	ProjectID *int64
	Model     string
	Tool      string
	Outcome   string
	Tag       string
	Limit     int
	All       bool
}

// EntryInput adalah payload pembuatan/pengubahan entri.
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
		return nil, errors.New("maximum of 10 tags")
	}
	for _, t := range out {
		if len(t) > 30 {
			return nil, errors.New("each tag must be at most 30 characters")
		}
	}
	return out, nil
}

// ParseTagsInput menerima tag dari string "a, b" atau array. API resminya
// array; string diterima sebagai kemudahan dan dipecah koma.
func ParseTagsInput(raw []string) ([]string, error) {
	expanded := []string{}
	for _, t := range raw {
		for _, part := range strings.Split(t, ",") {
			expanded = append(expanded, part)
		}
	}
	return NormalizeTags(expanded)
}
