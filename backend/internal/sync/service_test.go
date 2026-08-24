package sync

import (
	"testing"

	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/google/uuid"
)

func TestCanonicalSigningPayloadIsOrderIndependent(t *testing.T) {
	syncID := uuid.New()
	userID := uuid.New()
	deviceID := uuid.New()
	levelA := uuid.New()
	levelB := uuid.New()
	badgeA := uuid.New()
	badgeB := uuid.New()

	reqA := domain.SyncEventRequest{
		SyncEventID:      syncID,
		UserID:           userID,
		DeviceID:         deviceID,
		CryptoKeyVersion: 1,
		Payload: domain.SyncPayload{
			LevelAttempts: []domain.LevelAttemptItem{
				{LevelID: levelB, AttemptDate: "2026-07-20", AttemptNumber: 1, XPAwarded: 999, Completed: true},
				{LevelID: levelA, AttemptDate: "2026-07-19", AttemptNumber: 1, XPAwarded: 0, Completed: false},
			},
			DailyStreakDates: []string{"2026-07-20", "2026-07-19"},
			BadgeIDsEarned:   []uuid.UUID{badgeB, badgeA},
		},
	}
	reqB := domain.SyncEventRequest{
		SyncEventID:      syncID,
		UserID:           userID,
		DeviceID:         deviceID,
		CryptoKeyVersion: 1,
		Payload: domain.SyncPayload{
			LevelAttempts: []domain.LevelAttemptItem{
				reqA.Payload.LevelAttempts[1],
				reqA.Payload.LevelAttempts[0],
			},
			DailyStreakDates: []string{"2026-07-19", "2026-07-20"},
			BadgeIDsEarned:   []uuid.UUID{badgeA, badgeB},
		},
	}

	if got, want := CanonicalSigningPayload(reqA), CanonicalSigningPayload(reqB); got != want {
		t.Fatalf("canonical payload should be order-independent\ngot:  %s\nwant: %s", got, want)
	}
}

func TestCanonicalSigningPayloadHMAC(t *testing.T) {
	req := domain.SyncEventRequest{
		SyncEventID:      uuid.New(),
		UserID:           uuid.New(),
		DeviceID:         uuid.New(),
		CryptoKeyVersion: 1,
		Payload: domain.SyncPayload{
			LevelAttempts: []domain.LevelAttemptItem{
				{LevelID: uuid.New(), AttemptDate: "2026-07-20", AttemptNumber: 1, XPAwarded: 200, Completed: true},
			},
		},
	}
	secret := []byte("test-secret")
	payload := []byte(CanonicalSigningPayload(req))
	signature := crypto.GenerateHMAC(payload, secret)
	if !crypto.VerifyHMAC(payload, signature, secret) {
		t.Fatal("expected generated HMAC to verify")
	}
}

func TestXPForAttemptUsesDifficulty(t *testing.T) {
	tests := []struct {
		name       string
		difficulty int32
		attempt    int64
		completed  bool
		want       int32
	}{
		{name: "first attempt full XP", difficulty: 7, attempt: 1, completed: true, want: 28},
		{name: "second attempt half XP", difficulty: 7, attempt: 2, completed: true, want: 14},
		{name: "fourth attempt zero XP", difficulty: 7, attempt: 4, completed: true, want: 0},
		{name: "failed attempt zero XP", difficulty: 7, attempt: 1, completed: false, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := xpForAttempt(tt.difficulty, tt.attempt, tt.completed); got != tt.want {
				t.Fatalf("xpForAttempt() = %d, want %d", got, tt.want)
			}
		})
	}
}
