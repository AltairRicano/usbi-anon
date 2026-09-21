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

// Parámetros de Argon2id — alineados con el mínimo de OWASP 2023 para inicios de sesión interactivos.
// memory=64MB, time=3, threads=1. El servidor de producción tiene 1 vCPU, así que usar
// un solo hilo de Argon2 evita sobrecargar la CPU durante ráfagas de autenticación.
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

// HashPassword genera un hash Argon2id en formato PHC de la contraseña.
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

// VerifyPassword compara una contraseña en texto plano contra un hash Argon2id en formato PHC.
func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	// Formato PHC: ["", "argon2id", "v=19", "m=...", "b64salt", "b64hash"]
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

// GenerateHMAC calcula un HMAC-SHA256 sobre el payload usando secret.
func GenerateHMAC(payload, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(payload)
	return h.Sum(nil)
}

// VerifyHMAC realiza una comparación en tiempo constante del HMAC esperado frente al provisto.
func VerifyHMAC(payload, signature, secret []byte) bool {
	expected := GenerateHMAC(payload, secret)
	return hmac.Equal(expected, signature)
}

// BlindIndexHMAC genera un HMAC-SHA256 con clave para búsquedas deterministas
// de coincidencia exacta (correo, teléfono). Usar HMAC en lugar de SHA256 simple previene
// ataques de extensión de longitud y ataques de diccionario entre sistemas.
// secret DEBE derivarse de la variable de entorno BLIND_INDEX_SECRET.
func BlindIndexHMAC(data string, secret []byte) []byte {
	return GenerateHMAC([]byte(data), secret)
}
