package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/altair/usbi-anon-backend/internal/crypto"
	legaltext "github.com/altair/usbi-anon-backend/legal"
)

func newTestService() *Service {
	return NewService(nil, nil, Config{
		HMACSecret: []byte("test-hmac-secret"),
		TokenConfig: crypto.TokenConfig{
			Secret:       []byte("test-jwt-secret"),
			AccessExpiry: 15 * time.Minute,
		},
	})
}

func TestRegisterAnswers_RejectsOutdatedPrivacyVersion(t *testing.T) {
	svc := newTestService()

	_, err := svc.RegisterAnswers(context.Background(), RegisterAnswersRequest{
		Answers:              []AnswerInput{},
		PrivacyNoticeVersion: "v1.0-preliminar", // versión vieja, ya no vigente
	})
	if !errors.Is(err, ErrPrivacyVersionOutdated) {
		t.Fatalf("err = %v, want ErrPrivacyVersionOutdated", err)
	}
}

func TestRegisterAnswers_RejectsMissingPrivacyVersion(t *testing.T) {
	svc := newTestService()

	_, err := svc.RegisterAnswers(context.Background(), RegisterAnswersRequest{
		Answers:              []AnswerInput{},
		PrivacyNoticeVersion: "",
	})
	if !errors.Is(err, ErrPrivacyVersionOutdated) {
		t.Fatalf("err = %v, want ErrPrivacyVersionOutdated", err)
	}
}

func TestRegisterAnswers_AcceptsCurrentPrivacyVersion(t *testing.T) {
	svc := newTestService()

	// Con la versión vigente, la validación de privacidad pasa y el error
	// que sigue es el siguiente check de la función (answers vacías) — no
	// ErrPrivacyVersionOutdated. Confirma que el check de versión no rechaza
	// de más.
	_, err := svc.RegisterAnswers(context.Background(), RegisterAnswersRequest{
		Answers:              []AnswerInput{},
		PrivacyNoticeVersion: legaltext.CurrentVersion,
	})
	if errors.Is(err, ErrPrivacyVersionOutdated) {
		t.Fatalf("did not expect ErrPrivacyVersionOutdated for the current version, got %v", err)
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation (empty answers)", err)
	}
}
