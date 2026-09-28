// registration_token.go implementa el estado intermedio del registro: un
// token HMAC firmado con TTL de 10 minutos, sin tabla intermedia en la base.
// El payload completo viaja firmado en el propio token.
package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/google/uuid"
)

const registrationTokenTTL = 10 * time.Minute

type answerPayload struct {
	QuestionID           uuid.UUID `json:"question_id"`
	QuestionTextSnapshot string    `json:"question_text_snapshot"`
	AnswerText           string    `json:"answer_text"`
}

type registrationTokenPayload struct {
	Answers              []answerPayload `json:"answers"`
	IsAdult              bool            `json:"is_adult"`
	PrivacyNoticeVersion string          `json:"privacy_notice_version"`
	NicknameCandidates   []string        `json:"nickname_candidates"`
	ExpiresAt            time.Time       `json:"expires_at"`
}

// signRegistrationToken serializa el payload y lo firma.
// Principio: firmar, no cifrar — el cliente puede leer el contenido, no falsificarlo.
func (s *Service) signRegistrationToken(payload registrationTokenPayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sig := crypto.GenerateHMAC(raw, s.cfg.HMACSecret)
	return base64.RawURLEncoding.EncodeToString(raw) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

// verifyRegistrationToken valida la firma ANTES de confiar en cualquier
// campo del payload (incluido ExpiresAt) para evitar ataques de tiempo.
func (s *Service) verifyRegistrationToken(token string) (registrationTokenPayload, error) {
	var payload registrationTokenPayload

	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return payload, ErrRegistrationTokenInvalid
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return payload, ErrRegistrationTokenInvalid
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return payload, ErrRegistrationTokenInvalid
	}
	if !crypto.VerifyHMAC(raw, sig, s.cfg.HMACSecret) {
		return payload, ErrRegistrationTokenInvalid
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return payload, ErrRegistrationTokenInvalid
	}
	if time.Now().UTC().After(payload.ExpiresAt) {
		return payload, ErrRegistrationTokenExpired
	}
	return payload, nil
}

var (
	ErrRegistrationTokenInvalid = errors.New("registration token invalid")
	ErrRegistrationTokenExpired = errors.New("registration token expired")
)
