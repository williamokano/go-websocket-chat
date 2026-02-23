package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken_CreatesValidToken(t *testing.T) {
	svc := NewJWTService("test-secret-key")

	token, err := svc.GenerateToken("user-123", "alice")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken returned empty token")
	}
}

func TestValidateToken_SucceedsWithValidToken(t *testing.T) {
	svc := NewJWTService("test-secret-key")

	token, err := svc.GenerateToken("user-123", "alice")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
	if claims == nil {
		t.Fatal("ValidateToken returned nil claims")
	}
}

func TestValidateToken_ClaimsContainCorrectUserIDAndUsername(t *testing.T) {
	svc := NewJWTService("test-secret-key")

	token, err := svc.GenerateToken("user-456", "bob")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}

	if claims.UserID != "user-456" {
		t.Errorf("expected UserID %q, got %q", "user-456", claims.UserID)
	}
	if claims.Username != "bob" {
		t.Errorf("expected Username %q, got %q", "bob", claims.Username)
	}
}

func TestValidateToken_FailsWithExpiredToken(t *testing.T) {
	secret := []byte("test-secret-key")

	// Manually create an expired token
	claims := Claims{
		UserID:   "user-123",
		Username: "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	svc := NewJWTService("test-secret-key")
	_, err = svc.ValidateToken(tokenString)
	if err == nil {
		t.Fatal("ValidateToken should fail with expired token")
	}
}

func TestValidateToken_FailsWithWrongSecret(t *testing.T) {
	svc1 := NewJWTService("secret-one")
	svc2 := NewJWTService("secret-two")

	token, err := svc1.GenerateToken("user-123", "alice")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	_, err = svc2.ValidateToken(token)
	if err == nil {
		t.Fatal("ValidateToken should fail with wrong secret")
	}
}
