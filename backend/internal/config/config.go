// Package config centralises environment loading and parsing para el binario
// del servidor. Copiado de ../usbi/backend/internal/config; el rediseño de
// identidad de plan/04_Rediseno_identidad_gustos.md volvió a dejar una sola
// base de datos, así que la construcción del DSN vuelve a ser la de
// ../usbi (un único pool).
package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// LoadEnvironment loads .env (or the file named by USBI_BACKEND_ENV_FILE) into
// the process environment. Missing files are not fatal — in production, vars
// normally come from systemd/Docker/the hosting environment instead.
func LoadEnvironment() {
	if envFile := os.Getenv("USBI_BACKEND_ENV_FILE"); envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			log.Printf("[WARN] %s not found or unreadable — reading from system environment", envFile)
		}
		return
	}
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] .env not found — reading from system environment")
	}
}

// DatabaseURL builds the Postgres DSN for la única base del sistema
// (identidad + progreso + contenido, ver plan/04_Rediseno_identidad_gustos.md)
// from DATABASE_URL if set, else from the individual DB_* variables
// (DB_USER/DB_PASSWORD/DB_HOST required).
func DatabaseURL() string {
	return databaseURLFromEnv(dsnEnvNames{
		urlVar:      "DATABASE_URL",
		userVar:     "DB_USER",
		passwordVar: "DB_PASSWORD",
		hostVar:     "DB_HOST",
		portVar:     "DB_PORT",
		nameVar:     "DB_NAME",
		sslModeVar:  "DB_SSLMODE",
	})
}

// dsnEnvNames names the environment variables that describe one Postgres
// connection (una URL completa opcional, o el juego de variables sueltas).
type dsnEnvNames struct {
	urlVar      string
	userVar     string
	passwordVar string
	hostVar     string
	portVar     string
	nameVar     string
	sslModeVar  string
}

func databaseURLFromEnv(names dsnEnvNames) string {
	if val := os.Getenv(names.urlVar); val != "" {
		return val
	}

	dbUser := RequireEnv(names.userVar)
	dbPassword := RequireEnv(names.passwordVar)
	dbHost := RequireEnv(names.hostVar)
	dbPort := GetEnv(names.portVar, "5432")
	dbName := RequireEnv(names.nameVar)
	sslMode := GetEnv(names.sslModeVar, "disable")

	// A cleartext DB connection is only acceptable when it never leaves the host
	// (loopback). Warn loudly otherwise: credentials and personal data would
	// travel unencrypted (CN-008).
	if sslMode == "disable" && !isLoopbackHost(dbHost) {
		log.Printf("[WARN] %s=disable with non-loopback %s=%q: database traffic "+
			"(credentials and personal data) is unencrypted. Use %s=require or "+
			"verify-full when the database is not on localhost.",
			names.sslModeVar, names.hostVar, dbHost, names.sslModeVar)
	}

	dsn := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(dbUser, dbPassword),
		Host:   fmt.Sprintf("%s:%s", dbHost, dbPort),
		Path:   dbName,
	}
	query := dsn.Query()
	query.Set("sslmode", sslMode)
	dsn.RawQuery = query.Encode()
	return dsn.String()
}

// RequireEnv fatally exits if the environment variable is missing or empty.
// This surfaces misconfigurations immediately instead of silently failing later.
func RequireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("[FATAL] Required environment variable %q is not set", key)
	}
	return val
}

// MinSecretLength is the minimum accepted length (in bytes) for cryptographic
// secrets. 32 bytes matches the SHA-256 output used across JWT/HMAC signing.
const MinSecretLength = 32

// RequireSecret is RequireEnv plus a minimum-length check, so a binary refuses
// to start with a trivially guessable JWT/HMAC/blind-index/encryption key.
// Intentionally separate from RequireEnv, which also guards non-secret values
// (DB_USER, DB_HOST, …) where a length floor would be wrong.
func RequireSecret(key string) string {
	val := RequireEnv(key)
	if len(val) < MinSecretLength {
		log.Fatalf("[FATAL] %s must be at least %d characters for adequate strength (got %d)",
			key, MinSecretLength, len(val))
	}
	return val
}

// isLoopbackHost reports whether host is a local loopback address, where an
// unencrypted DB_SSLMODE=disable connection never leaves the machine.
func isLoopbackHost(host string) bool {
	switch strings.ToLower(strings.TrimSpace(host)) {
	case "127.0.0.1", "localhost", "::1", "[::1]":
		return true
	default:
		return false
	}
}

// GetEnv returns the environment variable value or a fallback default.
func GetEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// GetBoolEnv parses a boolean env var, fatally exiting on an invalid value.
func GetBoolEnv(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		log.Fatalf("[FATAL] Invalid %s: %v", key, err)
	}
	return parsed
}

// GetDurationEnv parses a time.Duration env var, fatally exiting on an invalid value.
func GetDurationEnv(key string, fallback time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(val)
	if err != nil {
		log.Fatalf("[FATAL] Invalid %s: %v", key, err)
	}
	return parsed
}

// GetInt32Env parses a positive int32 env var, fatally exiting on an invalid value.
func GetInt32Env(key string, fallback int32) int32 {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(val, 10, 32)
	if err != nil || parsed <= 0 {
		log.Fatalf("[FATAL] Invalid %s: %v", key, err)
	}
	return int32(parsed)
}

// CheckConnPoolBounds fatally exits if idle connections exceed open connections
// — la misma comprobación que main.go aplica sobre el único pool del sistema.
func CheckConnPoolBounds(maxIdle, maxOpen int32) {
	if maxIdle > maxOpen {
		log.Fatalf("[FATAL] DB_MAX_IDLE_CONNS (%d) cannot exceed DB_MAX_OPEN_CONNS (%d)", maxIdle, maxOpen)
	}
}
