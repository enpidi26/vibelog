package domain

import (
	"errors"
	"net/mail"
	"strings"
)

// NormalizeEmail merapikan dan memvalidasi alamat email. Display name
// ("Nama <a@b>") ditolak karena API memakai alamat telanjang.
func NormalizeEmail(raw string) (string, bool) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" || len(email) > 120 {
		return "", false
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || !strings.Contains(email, "@") {
		return "", false
	}
	return email, true
}

// RegisterInput adalah payload pendaftaran akun.
type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (in RegisterInput) Validate() error {
	if strings.TrimSpace(in.Name) == "" || len(in.Name) > 80 {
		return errors.New("name is required, max 80 characters")
	}
	if _, ok := NormalizeEmail(in.Email); !ok {
		return errors.New("invalid email")
	}
	if len(in.Password) < 8 || len(in.Password) > 128 {
		return errors.New("password must be between 8 and 128 characters")
	}
	return nil
}

// LoginInput adalah payload login password.
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
