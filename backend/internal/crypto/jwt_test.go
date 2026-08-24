package crypto

import (
	"strings"
	"testing"
	"time"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// crypto/jwt.go had zero tests (audit finding B9) despite being the real
// signing/verification path for every authenticated request. These exercise
// the actual jwt/v5 library end to end — no mocking of the cryptography.

func testConfig() TokenConfig {
	return TokenConfig{Secret: []byte("test-secret-at-least-32-bytes-long!!"), AccessExpiry: time.Hour}
}

func TestGenerateAndValidateTokenRoundTrip(t *testing.T) {
	cfg := testConfig()
	want := domain.JWTClaims{UserID: uuid.New(), Role: domain.RoleAdmin, TokenVersion: 3}

	token, err := GenerateToken(want, cfg)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	got, err := ValidateToken(token, cfg)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if got.UserID != want.UserID || got.Role != want.Role || got.TokenVersion != want.TokenVersion {
		t.Fatalf("ValidateToken() = %+v, want %+v", got, want)
	}
}

func TestValidateTokenRejectsExpiredToken(t *testing.T) {
	cfg := testConfig()
	cfg.AccessExpiry = -time.Minute // already expired at generation time

	token, err := GenerateToken(domain.JWTClaims{UserID: uuid.New(), Role: domain.RolePlayer, TokenVersion: 1}, cfg)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if _, err := ValidateToken(token, cfg); err == nil {
		t.Fatal("ValidateToken() error = nil, want expired-token error")
	}
}

func TestValidateTokenRejectsWrongSecret(t *testing.T) {
	cfg := testConfig()
	token, err := GenerateToken(domain.JWTClaims{UserID: uuid.New(), Role: domain.RolePlayer, TokenVersion: 1}, cfg)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	wrongCfg := cfg
	wrongCfg.Secret = []byte("a-completely-different-secret-value!!!")
	if _, err := ValidateToken(token, wrongCfg); err == nil {
		t.Fatal("ValidateToken() error = nil, want signature verification failure")
	}
}

func TestValidateTokenRejectsTamperedPayload(t *testing.T) {
	cfg := testConfig()
	token, err := GenerateToken(domain.JWTClaims{UserID: uuid.New(), Role: domain.RolePlayer, TokenVersion: 1}, cfg)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Flip a character in the payload segment to simulate a privilege-escalation
	// attempt (e.g. rewriting "player" to "admin") without re-signing.
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected token shape: %d segments", len(parts))
	}
	tampered := parts[0] + "." + parts[1] + "x" + "." + parts[2]

	if _, err := ValidateToken(tampered, cfg); err == nil {
		t.Fatal("ValidateToken() error = nil, want rejection of tampered payload")
	}
}

// TestValidateTokenRejectsAlgorithmConfusion guards the explicit
// SigningMethodHMAC check in ValidateToken against an attacker who crafts a
// token with a different alg (here "none") hoping the verifier skips signature
// checking entirely — a well-known real-world JWT vulnerability class.
func TestValidateTokenRejectsAlgorithmConfusion(t *testing.T) {
	cfg := testConfig()
	claims := jwt.MapClaims{
		"user_id":       uuid.New().String(),
		"role":          "admin",
		"token_version": 1,
		"exp":           jwt.NewNumericDate(time.Now().Add(time.Hour)).Unix(),
	}
	unsigned := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, err := unsigned.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("crafting alg=none token: %v", err)
	}

	if _, err := ValidateToken(tokenStr, cfg); err == nil {
		t.Fatal("ValidateToken() error = nil, want rejection of alg=none token")
	}
}
