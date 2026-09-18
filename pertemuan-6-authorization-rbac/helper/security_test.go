package helper

import (
	"testing"
)

func TestHashPasswordAndVerify(t *testing.T) {
	plain := "correct-password1"
	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatal(err)
	}
	if hash == plain {
		t.Fatal("password must not equal hash")
	}
	if !VerifyPassword(hash, plain) {
		t.Fatal("expected password to verify")
	}
	if VerifyPassword(hash, "wrong-password1") {
		t.Fatal("wrong password must fail")
	}
}

func TestHashPasswordUsesDifferentSalt(t *testing.T) {
	first, err := HashPassword("correct-password1")
	if err != nil {
		t.Fatal(err)
	}
	second, err := HashPassword("correct-password1")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("hashes should use different salts")
	}
}

func TestDummyPasswordVerification(t *testing.T) {
	VerifyDummyPassword("wrong-password1")
}

func TestRandomTokenAndSHA256Hex(t *testing.T) {
	first, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	second, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	if first == "" || second == "" || first == second {
		t.Fatal("tokens must be non-empty and unique")
	}
	if SHA256Hex(first) != SHA256Hex(first) {
		t.Fatal("hash must be deterministic")
	}
	if SHA256Hex(first) == first {
		t.Fatal("hash must not equal raw token")
	}
}
