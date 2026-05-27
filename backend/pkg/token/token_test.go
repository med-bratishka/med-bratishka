package token

import (
	"testing"
	"time"

	"medbratishka/internal/domain"
)

func TestJWTTokenManagerGenerateParseRoundTrip(t *testing.T) {
	manager := NewJWTTokenManager("jwt-secret")
	expiresAt := time.Now().Add(time.Hour).Truncate(time.Second)

	tokenString, err := manager.GenerateToken(&domain.TokenData{
		ID:        42,
		Role:      domain.RoleDoctor,
		Login:     "doctor@example.com",
		Purpose:   domain.PurposeAccess,
		Number:    3,
		Secret:    "session-secret",
		ExpiresAt: expiresAt,
	})
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	parsed, err := manager.ParseToken(tokenString)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}

	if parsed.ID != 42 || parsed.Role != domain.RoleDoctor || parsed.Login != "doctor@example.com" {
		t.Fatalf("unexpected parsed identity: %#v", parsed)
	}
	if parsed.Purpose != domain.PurposeAccess || parsed.Number != 3 || parsed.Secret != "session-secret" {
		t.Fatalf("unexpected parsed session data: %#v", parsed)
	}
	if !parsed.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("expected expires_at %s, got %s", expiresAt, parsed.ExpiresAt)
	}
}

func TestJWTTokenManagerRejectsWrongSecret(t *testing.T) {
	tokenString, err := NewJWTTokenManager("jwt-secret").GenerateToken(&domain.TokenData{
		ID:        42,
		Role:      domain.RoleDoctor,
		Login:     "doctor@example.com",
		Purpose:   domain.PurposeAccess,
		Number:    1,
		Secret:    "session-secret",
		ExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	if _, err := NewJWTTokenManager("other-secret").ParseToken(tokenString); err == nil {
		t.Fatal("expected parse with wrong secret to fail")
	}
}

func TestJWTTokenManagerRejectsMalformedToken(t *testing.T) {
	if _, err := NewJWTTokenManager("jwt-secret").ParseToken("not-a-token"); err == nil {
		t.Fatal("expected malformed token to fail")
	}
}
