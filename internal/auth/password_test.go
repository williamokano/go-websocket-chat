package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword_ReturnsValidBcryptHash(t *testing.T) {
	hash, err := HashPassword("mysecretpassword")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}

	// Verify it's a valid bcrypt hash by checking the prefix
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("mysecretpassword")); err != nil {
		t.Fatalf("hash is not a valid bcrypt hash: %v", err)
	}
}

func TestCheckPassword_CorrectPassword(t *testing.T) {
	hash, err := HashPassword("correctpassword")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if !CheckPassword("correctpassword", hash) {
		t.Fatal("CheckPassword returned false for correct password")
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	hash, err := HashPassword("correctpassword")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if CheckPassword("wrongpassword", hash) {
		t.Fatal("CheckPassword returned true for wrong password")
	}
}

func TestHashPassword_DifferentPasswordsProduceDifferentHashes(t *testing.T) {
	hash1, err := HashPassword("password1")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	hash2, err := HashPassword("password2")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash1 == hash2 {
		t.Fatal("different passwords produced the same hash")
	}
}

func TestHashPassword_SamePasswordProducesDifferentHashes(t *testing.T) {
	// bcrypt uses random salt, so same password should produce different hashes
	hash1, err := HashPassword("samepassword")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	hash2, err := HashPassword("samepassword")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash1 == hash2 {
		t.Fatal("same password produced identical hashes (expected different due to salt)")
	}
}
