package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters — aligned with OWASP 2023 minimum for interactive logins.
// memory=64MB, time=3, threads=1. The production server has 1 vCPU, so using
// a single Argon2 lane avoids oversubscribing CPU during auth bursts.
const (
	argon2Memory      uint32 = 64 * 1024
	argon2Iterations  uint32 = 3
	argon2Parallelism uint8  = 1
	argon2SaltLength  int    = 16
	argon2KeyLength   uint32 = 32
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")
)

// HashPassword generates a PHC-format Argon2id hash of the password.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password), salt,
		argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Iterations, argon2Parallelism,
		b64Salt, b64Hash,
	)

	return encodedHash, nil
}

// VerifyPassword compares a plaintext password against a PHC-format Argon2id hash.
func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	// PHC format: ["", "argon2id", "v=19", "m=...", "b64salt", "b64hash"]
	if len(parts) != 6 {
		return false, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("parsing argon2 version: %w", err)
	}
	if version != argon2.Version {
		return false, ErrIncompatibleVersion
	}

	var mem, iters uint32
	var par uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &iters, &par); err != nil {
		return false, fmt.Errorf("parsing argon2 params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decoding salt: %w", err)
	}

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decoding hash: %w", err)
	}

	computedHash := argon2.IDKey([]byte(password), salt, iters, mem, par, uint32(len(storedHash)))
	return subtle.ConstantTimeCompare(storedHash, computedHash) == 1, nil
}

// GenerateHMAC computes an HMAC-SHA256 over the payload using secret.
func GenerateHMAC(payload, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(payload)
	return h.Sum(nil)
}

// VerifyHMAC performs a constant-time comparison of the expected vs provided HMAC.
func VerifyHMAC(payload, signature, secret []byte) bool {
	expected := GenerateHMAC(payload, secret)
	return hmac.Equal(expected, signature)
}

// BlindIndexHMAC generates a keyed HMAC-SHA256 for deterministic exact-match
// lookups (email, phone). Using HMAC instead of plain SHA256 prevents
// length-extension attacks and cross-system dictionary attacks.
// secret MUST be derived from BLIND_INDEX_SECRET environment variable.
func BlindIndexHMAC(data string, secret []byte) []byte {
	return GenerateHMAC([]byte(data), secret)
}
