package domain

import (
	"testing"
)

func TestEntryValidation(t *testing.T) {
	bad := EntryInput{Title: "x", PromptText: "p", Model: "Claude", Tool: "Cursor", Outcome: "aneh"}
	if err := bad.Validate(); err == nil {
		t.Fatal("outcome tidak dikenal harus ditolak")
	}
	rating := 6
	bad2 := EntryInput{Title: "x", PromptText: "p", Model: "Claude", Tool: "Cursor", Outcome: OutcomeSuccess, Rating: &rating}
	if err := bad2.Validate(); err == nil {
		t.Fatal("rating 6 harus ditolak")
	}
}

func TestNormalizeTags(t *testing.T) {
	tags, err := NormalizeTags([]string{"AI Agent", "ai-agent", "  ", "a/b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 3 || tags[0] != "ai agent" || tags[1] != "ai-agent" || tags[2] != "ab" {
		t.Fatalf("unexpected tags: %v", tags)
	}
	if _, err := NormalizeTags([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}); err == nil {
		t.Fatal("lebih dari 10 tag harus ditolak")
	}
}

func TestNormalizeEmail(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{" Ana@Example.com ", "ana@example.com", true},
		{"ana@example.com", "ana@example.com", true},
		{"Name <ana@example.com>", "", false},
		{"bukan-email", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, ok := NormalizeEmail(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("NormalizeEmail(%q) = %q,%v; want %q,%v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestRegisterValidation(t *testing.T) {
	ok := RegisterInput{Name: "Ana", Email: "ana@example.com", Password: "password123"}
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid register rejected: %v", err)
	}
	if err := (RegisterInput{Name: "Ana", Email: "nope", Password: "password123"}).Validate(); err == nil {
		t.Fatal("invalid email harus ditolak")
	}
	if err := (RegisterInput{Name: "Ana", Email: "ana@example.com", Password: "short"}).Validate(); err == nil {
		t.Fatal("password pendek harus ditolak")
	}
}

func TestProjectValidation(t *testing.T) {
	if err := (ProjectInput{Name: "VibeLog", Color: "teal"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (ProjectInput{Name: "VibeLog", Color: "pink"}).Validate(); err == nil {
		t.Fatal("warna tidak dikenal harus ditolak")
	}
}
