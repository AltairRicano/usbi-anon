package crypto

import (
	"strings"
	"testing"
	"time"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// crypto/jwt.go no tenía pruebas (hallazgo de auditoría B9) a pesar de ser la ruta
// real de firma/verificación para cada petición autenticada. Estas prueban
// la biblioteca jwt/v5 real de extremo a extremo — sin mockear la criptografía. (Útil)

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

	// Voltea un carácter en el segmento del payload para simular un intento de
	// escalamiento de privilegios (ej. reescribiendo "player" a "admin") sin volver a firmar. (Útil)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected token shape: %d segments", len(parts))
	}
	tampered := parts[0] + "." + parts[1] + "x" + "." + parts[2]

	if _, err := ValidateToken(tampered, cfg); err == nil {
		t.Fatal("ValidateToken() error = nil, want rejection of tampered payload")
	}
}

// TestValidateTokenRejectsAlgorithmConfusion protege la comprobación explícita de
// SigningMethodHMAC en ValidateToken contra un atacante que crea un
// token con un alg diferente (aquí "none") esperando que el verificador salte por
// completo la comprobación de la firma — una vulnerabilidad de JWT del mundo real bien conocida. (Útil)
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
