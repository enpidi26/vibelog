package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword memakai bcrypt dengan cost default (10).
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

type claims struct {
	UserID int64  `json:"sub"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

// IssueToken membuat JWT HS256 untuk satu user.
func IssueToken(secret string, userID int64, email, name string, expiryHours int) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID: userID,
		Email:  email,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expiryHours) * time.Hour)),
			Issuer:    "vibelog-api",
		},
	})
	return token.SignedString([]byte(secret))
}

// ParseToken memvalidasi token dan mengembalikan identitas user.
func ParseToken(secret, tokenString string) (userID int64, email, name string, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritma token tidak didukung: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer("vibelog-api"))
	if err != nil {
		return 0, "", "", err
	}
	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid || c.UserID <= 0 {
		return 0, "", "", errors.New("invalid token")
	}
	return c.UserID, c.Email, c.Name, nil
}
