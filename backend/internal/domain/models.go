package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserRole defines valid roles for users in the USBI system.
type UserRole string

const (
	RolePlayer   UserRole = "player"
	RoleAdmin    UserRole = "admin"
	RoleOperator UserRole = "operator"
	RoleDirector UserRole = "director"
)

// UserStatus defines valid account lifecycle statuses.
type UserStatus string

const (
	StatusActive              UserStatus = "active"
	StatusSuspended           UserStatus = "suspended"
	StatusPendingTutorConsent UserStatus = "pending_tutor_consent"
	StatusDeleted             UserStatus = "deleted"
)

// User is the decoupled public DTO. It NEVER includes password_hash,
// email/phone ciphertext, HMAC keys, or any cryptographic material.
type User struct {
	ID        uuid.UUID  `json:"id"`
	FullName  string     `json:"full_name"`
	IsAdult   bool       `json:"is_adult"`
	Role      UserRole   `json:"role"`
	Status    UserStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

// JWTClaims carries the standard claims embedded in every USBI JWT.
// TokenVersion is validated against DB on every authenticated request.
type JWTClaims struct {
	UserID       uuid.UUID `json:"user_id"`
	Role         UserRole  `json:"role"`
	TokenVersion int       `json:"token_version"`
}

// SyncSource defines valid origins for experience events.
type SyncSource string

const (
	SyncSourceOnline      SyncSource = "online"
	SyncSourceOfflineSync SyncSource = "offline_sync"
)

// VerificationMethod defines how XP was verified.
type VerificationMethod string

const (
	VerificationOnlineDirect VerificationMethod = "online_direct"
	VerificationHMACOffline  VerificationMethod = "hmac_offline"
)

// LevelAttemptItem represents a single level attempt from an offline sync payload.
// NOTE: xp_awarded is provided by the client but MUST be recalculated by Go
// backend using official level difficulty and transactional locking. Client value is untrusted.
type LevelAttemptItem struct {
	LevelID       uuid.UUID `json:"level_id"`
	AttemptDate   string    `json:"attempt_date"` // ISO 8601 date: YYYY-MM-DD
	AttemptNumber int       `json:"attempt_number"`
	XPAwarded     int       `json:"xp_awarded"` // Untrusted. Backend recalculates.
	// Score is the in-game score for this attempt (NOT XP). It feeds
	// player_progress.best_score exactly like the online path; XP is always
	// recalculated server-side and never derived from Score. Must be >= 0.
	Score     int  `json:"score"`
	Completed bool `json:"completed"`
}

// SyncPayload is the strictly typed offline progress packet.
// It MUST NOT contain name, email, phone, tutor data, or any PII.
type SyncPayload struct {
	LevelAttempts    []LevelAttemptItem `json:"level_attempts"`
	DailyStreakDates []string           `json:"daily_streak_dates,omitempty"` // YYYY-MM-DD
	BadgeIDsEarned   []uuid.UUID        `json:"badge_ids_earned,omitempty"`
}

// SyncEventRequest is the incoming body for POST /api/v1/sync.
type SyncEventRequest struct {
	SyncEventID      uuid.UUID   `json:"sync_event_id"` // Idempotency key
	UserID           uuid.UUID   `json:"user_id"`
	DeviceID         uuid.UUID   `json:"device_id"`
	CryptoKeyVersion int         `json:"crypto_key_version"`
	Payload          SyncPayload `json:"payload"`
	HMACSignature    []byte      `json:"hmac_signature"` // HMAC-SHA256, Base64
}

// SyncEventResponse is returned by POST /api/v1/sync.
type SyncEventResponse struct {
	Status        string      `json:"status"`          // "synced" | "already_processed"
	WipeLocalData bool        `json:"wipe_local_data"` // ARCO cancellation flag
	ServerXPTotal int         `json:"server_xp_total"` // Post-merge total for client validation
	BadgesAwarded []uuid.UUID `json:"badges_awarded,omitempty"`
}

// ProblemDetails implements RFC 7807 for all error responses.
type ProblemDetails struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// ArcoRequestType defines the valid ARCO right types.
type ArcoRequestType string

const (
	ArcoAcceso        ArcoRequestType = "acceso"
	ArcoRectificacion ArcoRequestType = "rectificacion"
	ArcoCancelacion   ArcoRequestType = "cancelacion"
	ArcoOposicion     ArcoRequestType = "oposicion"
)

// ArcoRequest is the body for POST /api/v1/arco.
type ArcoRequest struct {
	RequestType ArcoRequestType `json:"request_type"`
	Details     string          `json:"details,omitempty"`
}

// ContextKey is an unexported type for context keys to avoid collisions
// with other packages that use context.WithValue.
type ContextKey string

// ClaimsKey is the context key for storing JWT claims in the request context.
const ClaimsKey ContextKey = "jwt_claims"
