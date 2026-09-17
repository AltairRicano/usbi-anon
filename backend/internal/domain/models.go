package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RolePlayer UserRole = "player"
	RoleAdmin  UserRole = "admin"
)

type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusSuspended UserStatus = "suspended"
	StatusDeleted   UserStatus = "deleted"
)

// User es el DTO público desacoplado. NUNCA incluye password_hash,
// texto cifrado de email/teléfono, claves HMAC ni ningún material criptográfico.
//
// FullName no existe: USBI-Anon no guarda nombre real en ningún lado del
// sistema principal. Tampoco existe ningún dato de tutor ni de correo — el
// registro se hace por cuestionario de gustos, y Nickname es la credencial de login
// derivada de esas respuestas. DisplayAlias es un dato distinto y no
// relacionado: un alias 100% aleatorio ("Jaguar Azul 42"), resuelto contra la
// vista account_aliases. Ninguno de los dos debe usarse como clave de
// búsqueda salvo Nickname, que sí lo es por diseño (UNIQUE, `WHERE nickname =
// $1`).
type User struct {
	ID           uuid.UUID  `json:"id"`
	Nickname     string     `json:"nickname"`
	DisplayAlias string     `json:"display_alias"`
	IsAdult      bool       `json:"is_adult"`
	Role         UserRole   `json:"role"`
	Status       UserStatus `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}

// JWTClaims transporta los claims estándar integrados en cada JWT de USBI.
// TokenVersion se valida contra la base de datos en cada petición autenticada.
type JWTClaims struct {
	UserID       uuid.UUID `json:"user_id"`
	Role         UserRole  `json:"role"`
	TokenVersion int       `json:"token_version"`
}

// SyncSource define los orígenes válidos para los eventos de experiencia.
type SyncSource string

const (
	SyncSourceOnline      SyncSource = "online"
	SyncSourceOfflineSync SyncSource = "offline_sync"
)

// VerificationMethod define cómo se verificó la XP.
type VerificationMethod string

const (
	// VerificationOnlineVerified: el servidor recalculó completed/score desde
	// answers contra level.content (M1 Fase B, B-1/B-2) — inmune a manipulación.
	VerificationOnlineVerified VerificationMethod = "online_verified"
	// VerificationOnlineReported: no hay verificación posible para la
	// plantilla (memory, snakes_ladders — M1.4-B3) y el servidor acepta el
	// completed/score que reportó el cliente.
	VerificationOnlineReported VerificationMethod = "online_reported"
	VerificationHMACOffline    VerificationMethod = "hmac_offline"
)

// LevelAttemptItem representa un intento de nivel individual desde un payload de sincronización offline.
// NOTA: xp_awarded es provisto por el cliente pero DEBE ser recalculado por el
// backend en Go usando la dificultad oficial del nivel y bloqueo transaccional. El valor del cliente no es confiable.
type LevelAttemptItem struct {
	LevelID       uuid.UUID `json:"level_id"`
	AttemptDate   string    `json:"attempt_date"` // fecha ISO 8601: YYYY-MM-DD
	AttemptNumber int       `json:"attempt_number"`
	XPAwarded     int       `json:"xp_awarded"` // No confiable. El backend recalcula.
	// Score es la puntuación del juego para este intento (NO la XP). Alimenta a
	// player_progress.best_score exactamente igual que la ruta online; la XP siempre
	// se recalcula del lado del servidor y nunca se deriva del Score. Debe ser >= 0.
	Score     int  `json:"score"`
	Completed bool `json:"completed"`
}

// SyncPayload es el paquete de progreso offline estrictamente tipado.
// NO DEBE contener nombre, correo electrónico, teléfono, datos de tutor, ni ningún dato personal identificable (PII).
type SyncPayload struct {
	LevelAttempts    []LevelAttemptItem `json:"level_attempts"`
	DailyStreakDates []string           `json:"daily_streak_dates,omitempty"` // YYYY-MM-DD
	BadgeIDsEarned   []uuid.UUID        `json:"badge_ids_earned,omitempty"`
}

// SyncEventRequest es el cuerpo de entrada para POST /api/v1/sync.
type SyncEventRequest struct {
	SyncEventID      uuid.UUID   `json:"sync_event_id"` // Clave de idempotencia
	UserID           uuid.UUID   `json:"user_id"`
	DeviceID         uuid.UUID   `json:"device_id"`
	CryptoKeyVersion int         `json:"crypto_key_version"`
	Payload          SyncPayload `json:"payload"`
	HMACSignature    []byte      `json:"hmac_signature"` // HMAC-SHA256, Base64
}

// SyncEventResponse es devuelto por POST /api/v1/sync.
type SyncEventResponse struct {
	Status        string      `json:"status"`          // "synced" | "already_processed"
	WipeLocalData bool        `json:"wipe_local_data"` // establecido en la cancelación de cuenta (autoservicio)
	ServerXPTotal int         `json:"server_xp_total"` // Total post-fusión para la validación del cliente
	BadgesAwarded []uuid.UUID `json:"badges_awarded,omitempty"`
}

// ProblemDetails implementa RFC 7807 para todas las respuestas de error.
type ProblemDetails struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// ContextKey es un tipo no exportado para las claves de contexto para evitar colisiones
// con otros paquetes que usan context.WithValue.
type ContextKey string

// ClaimsKey es la clave de contexto para almacenar los claims JWT en el contexto de la petición.
const ClaimsKey ContextKey = "jwt_claims"
