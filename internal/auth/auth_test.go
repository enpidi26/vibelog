package auth

import "testing"

func TestPasswordAndTokenRoundTrip(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword("password123", hash) || CheckPassword("salah", hash) {
		t.Fatal("verifikasi password salah")
	}
	token, err := IssueToken("secret-yang-cukup-panjang-untuk-test", 7, "a@b.test", "Ana", 24)
	if err != nil {
		t.Fatal(err)
	}
	uid, email, _, err := ParseToken("secret-yang-cukup-panjang-untuk-test", token)
	if err != nil || uid != 7 || email != "a@b.test" {
		t.Fatalf("parse token: uid=%d email=%s err=%v", uid, email, err)
	}
	if _, _, _, err := ParseToken("secret-lain", token); err == nil {
		t.Fatal("token dengan secret lain harus ditolak")
	}
}

func TestRandomToken(t *testing.T) {
	a, err := RandomToken(16)
	if err != nil {
		t.Fatal(err)
	}
	b, err := RandomToken(16)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 32 || a == b {
		t.Fatalf("token must be 32 hex chars and unique: %q %q", a, b)
	}
}
