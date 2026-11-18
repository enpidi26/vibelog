package database

import (
	"errors"
	"testing"
	"time"
)

func TestMagicTokenLifecycle(t *testing.T) {
	db, _ := testDB(t)
	u := testUser(t, db, "magic@test.id")

	if err := IssueMagicToken(db, u.ID, "token-one", 15*time.Minute); err != nil {
		t.Fatal(err)
	}
	userID, err := ConsumeMagicToken(db, "token-one")
	if err != nil {
		t.Fatal(err)
	}
	if userID != u.ID {
		t.Fatalf("userID = %d, want %d", userID, u.ID)
	}
	got, err := GetUserByID(db, u.ID)
	if err != nil || !got.EmailVerified {
		t.Fatalf("email harus terverifikasi setelah konsumsi token: %+v err=%v", got, err)
	}
	if _, err := ConsumeMagicToken(db, "token-one"); !errors.Is(err, ErrTokenUsed) {
		t.Fatalf("replay harus ErrTokenUsed, got %v", err)
	}
	if _, err := ConsumeMagicToken(db, "tidak-ada"); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("token asing harus ErrTokenInvalid, got %v", err)
	}
}

func TestNewMagicTokenInvalidatesOld(t *testing.T) {
	db, _ := testDB(t)
	u := testUser(t, db, "magic@test.id")

	if err := IssueMagicToken(db, u.ID, "token-one", 15*time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := IssueMagicToken(db, u.ID, "token-two", 15*time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := ConsumeMagicToken(db, "token-one"); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("token lama harus invalid, got %v", err)
	}
	if _, err := ConsumeMagicToken(db, "token-two"); err != nil {
		t.Fatalf("token baru harus valid: %v", err)
	}
}

func TestExpiredMagicToken(t *testing.T) {
	db, _ := testDB(t)
	u := testUser(t, db, "magic@test.id")
	if err := IssueMagicToken(db, u.ID, "token-old", -time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := ConsumeMagicToken(db, "token-old"); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("token kadaluarsa harus ErrTokenExpired, got %v", err)
	}
}

func TestCleanupExpiredTokens(t *testing.T) {
	db, _ := testDB(t)
	u := testUser(t, db, "magic@test.id")
	if err := IssueMagicToken(db, u.ID, "expired", -48*time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := IssueMagicToken(db, u.ID, "fresh", 10*time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := CleanupExpiredTokens(db); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM otp_codes WHERE code = 'expired'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("expired token should be deleted")
	}
	if _, err := ConsumeMagicToken(db, "fresh"); err != nil {
		t.Fatalf("fresh token should survive cleanup: %v", err)
	}
}
