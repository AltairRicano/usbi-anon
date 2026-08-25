package auth

import (
	"time"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/google/uuid"
)

// RegisterRequest is the body for POST /api/v1/auth/register.
// FullName and Phone no existen: no se registra PII más allá del correo,
// necesario únicamente para autenticar (ver plan/01_Base_de_datos.md §2.1).
type RegisterRequest struct {
	Email                string `json:"email"`
	Password             string `json:"password"`
	IsAdult              bool   `json:"is_adult"`
	PrivacyNoticeVersion string `json:"privacy_notice_version"`
}

// RegisterResponse is returned on successful registration.
type RegisterResponse struct {
	UserID  uuid.UUID `json:"user_id"`
	Status  string    `json:"status"`
	Message string    `json:"message"`
	// RegistrationToken is only set when Status is pending_tutor_consent.
	// The client must echo it back in TutorConsentRequest — it proves the
	// caller was present at registration, preventing third parties who
	// merely learn the user_id from submitting themselves as the tutor.
	RegistrationToken string `json:"registration_token,omitempty"`
}

// LoginRequest is the body for POST /api/v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is returned on successful login.
// User is typed as domain.User — never contains PII or crypto material.
type LoginResponse struct {
	AccessToken           string      `json:"access_token"`
	RefreshToken          string      `json:"refresh_token"`
	TokenType             string      `json:"token_type"`
	AccessTokenExpiresIn  int64       `json:"access_token_expires_in"`
	RefreshTokenExpiresAt time.Time   `json:"refresh_token_expires_at"`
	User                  domain.User `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// TutorConsentRequest is the body for POST /api/v1/auth/tutor-consent.
type TutorConsentRequest struct {
	UserID               uuid.UUID `json:"user_id"`
	TutorName            string    `json:"tutor_name"`
	TutorEmail           string    `json:"tutor_email"`
	PrivacyNoticeVersion string    `json:"privacy_notice_version"`
	// RegistrationToken must match the value returned by /auth/register for
	// this UserID (see RegisterResponse.RegistrationToken).
	RegistrationToken string `json:"registration_token"`
}

// ArcoRequestDTO is the body for POST /api/v1/arco.
// Uses domain type for compile-time validation.
type ArcoRequestDTO struct {
	RequestType domain.ArcoRequestType `json:"request_type"`
	Details     string                 `json:"details,omitempty"`
}

type ArcoResponseDTO struct {
	RequestID uuid.UUID `json:"request_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
}

type ArcoPendingItemDTO struct {
	ID            uuid.UUID `json:"id"`
	RequesterType string    `json:"requester_type"`
	RequestType   string    `json:"request_type"`
	Status        string    `json:"status"`
	ReceivedAt    time.Time `json:"received_at"`
}

type ArcoPendingListDTO struct {
	Items []ArcoPendingItemDTO `json:"items"`
}

type ResolveArcoRequestDTO struct {
	Approved        bool   `json:"approved"`
	ResponseSummary string `json:"response_summary"`
}
