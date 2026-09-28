// Package config centraliza la carga y análisis de variables de entorno para el binario del servidor.
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

// LoadEnvironment carga .env (o el archivo nombrado por USBI_BACKEND_ENV_FILE) en
// el entorno del proceso. Los archivos faltantes no son fatales — en producción, las variables
// normalmente provienen de systemd/Docker/o el entorno de alojamiento.
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

// DatabaseURL construye el DSN de Postgres para el pool de jugador (rol usbi_app)
// a partir de DATABASE_URL si está configurado, o desde las variables
// DB_* individuales (DB_USER/DB_PASSWORD/DB_HOST son obligatorias).
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

// ModeratorDatabaseURL construye el DSN del pool de moderador (rol usbi_moderador)
// para operaciones administrativas (gestión de contenido, banco de preguntas, incidentes de
// seguridad). Usa variables independientes para no mezclar credenciales entre pools.
func ModeratorDatabaseURL() string {
	return databaseURLFromEnv(dsnEnvNames{
		urlVar:      "DATABASE_MODERATOR_URL",
		userVar:     "DB_MODERATOR_USER",
		passwordVar: "DB_MODERATOR_PASSWORD",
		hostVar:     "DB_MODERATOR_HOST",
		portVar:     "DB_MODERATOR_PORT",
		nameVar:     "DB_MODERATOR_NAME",
		sslModeVar:  "DB_MODERATOR_SSLMODE",
	})
}

// DBMaintDatabaseURL construye el DSN del pool de mantenimiento de particiones
// (rol usbi_dbmaint): sin privilegios de tabla, únicamente EXECUTE sobre ensure_yearly_partition.
func DBMaintDatabaseURL() string {
	return databaseURLFromEnv(dsnEnvNames{
		urlVar:      "DATABASE_DBMAINT_URL",
		userVar:     "DB_DBMAINT_USER",
		passwordVar: "DB_DBMAINT_PASSWORD",
		hostVar:     "DB_DBMAINT_HOST",
		portVar:     "DB_DBMAINT_PORT",
		nameVar:     "DB_DBMAINT_NAME",
		sslModeVar:  "DB_DBMAINT_SSLMODE",
	})
}

// MigrateDatabaseURL construye el DSN del rol de migración (usbi_migrate,
// dueño del esquema — ver backend/sql/00_roles_unificado.sql). Lo usa
// exclusivamente cmd/usbictl migrate (M3.3): el servidor HTTP nunca aplica
// migraciones ni conoce este rol.
func MigrateDatabaseURL() string {
	return databaseURLFromEnv(dsnEnvNames{
		urlVar:      "DATABASE_MIGRATE_URL",
		userVar:     "DB_MIGRATE_USER",
		passwordVar: "DB_MIGRATE_PASSWORD",
		hostVar:     "DB_MIGRATE_HOST",
		portVar:     "DB_MIGRATE_PORT",
		nameVar:     "DB_MIGRATE_NAME",
		sslModeVar:  "DB_MIGRATE_SSLMODE",
	})
}

// dsnEnvNames nombra las variables de entorno que describen una conexión de Postgres
// (una URL completa opcional, o el juego de variables sueltas).
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

	// Una conexión de base de datos en texto plano solo es aceptable cuando nunca sale del host
	// (loopback). Advierte si no es así, ya que las credenciales y datos personales viajarían sin cifrar.
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

// RequireEnv termina de forma fatal si la variable de entorno falta o está vacía.
// Esto revela malas configuraciones inmediatamente en lugar de fallar silenciosamente más tarde.
func RequireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("[FATAL] Required environment variable %q is not set", key)
	}
	return val
}

// MinSecretLength es la longitud mínima aceptada (en bytes) para secretos criptográficos.
// 32 bytes coinciden con la salida SHA-256 usada en firmas JWT/HMAC.
const MinSecretLength = 32

// RequireSecret es RequireEnv más una comprobación de longitud mínima, para que un binario se rehúse
// a iniciar con una clave JWT/HMAC/blind-index/encriptación trivialmente adivinable.
// Está separada intencionalmente de RequireEnv, que también protege valores no secretos
// (DB_USER, DB_HOST, …) donde un mínimo de longitud sería incorrecto.
func RequireSecret(key string) string {
	val := RequireEnv(key)
	if len(val) < MinSecretLength {
		log.Fatalf("[FATAL] %s must be at least %d characters for adequate strength (got %d)",
			key, MinSecretLength, len(val))
	}
	return val
}

// isLoopbackHost reporta si el host es una dirección local de loopback, donde una
// conexión no cifrada DB_SSLMODE=disable nunca sale de la máquina.
func isLoopbackHost(host string) bool {
	switch strings.ToLower(strings.TrimSpace(host)) {
	case "127.0.0.1", "localhost", "::1", "[::1]":
		return true
	default:
		return false
	}
}

// GetEnv devuelve el valor de la variable de entorno o un valor por defecto.
func GetEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// GetBoolEnv analiza una variable de entorno booleana, cerrando de forma fatal si hay un valor inválido.
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

// GetDurationEnv analiza una variable de entorno time.Duration, cerrando de forma fatal si hay un valor inválido.
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

// GetInt32Env analiza una variable de entorno int32 positiva, cerrando de forma fatal si hay un valor inválido.
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

// CheckConnPoolBounds cierra de forma fatal si las conexiones inactivas exceden las abiertas
// — la misma comprobación que main.go aplica sobre el único pool del sistema.
func CheckConnPoolBounds(maxIdle, maxOpen int32) {
	if maxIdle > maxOpen {
		log.Fatalf("[FATAL] DB_MAX_IDLE_CONNS (%d) cannot exceed DB_MAX_OPEN_CONNS (%d)", maxIdle, maxOpen)
	}
}
