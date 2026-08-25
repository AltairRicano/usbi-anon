package sync_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/altair/usbi-anon-backend/internal/auth"
	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/altair/usbi-anon-backend/internal/devices"
	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/levels"
	syncpkg "github.com/altair/usbi-anon-backend/internal/sync"
	"github.com/altair/usbi-anon-backend/internal/testdb"
	"github.com/google/uuid"
)

// Integration coverage for ProcessSync against a real Postgres schema (audit
// finding B9) — the HMAC verification, idempotency, and future-date rejection
// paths were previously untested end-to-end; only pure helper functions had
// unit tests. Run with TEST_DATABASE_URL set (see internal/testdb); skipped
// otherwise.
//
// Diferido de F3 a F4 (dependía de auth.NewService/testdb.Setup, reescritos
// aquí — ver estado_proyecto.md). Frente a ../usbi, este archivo YA NO puede
// ser diff-cero: testdb.Setup ahora devuelve *testdb.DB (dos bases, no una),
// auth.NewService recibe (ident, main, cfg), Register perdió FullName, y
// devices.RegisterDeviceRequest usa DeviceKind en vez de DeviceLabel. Es una
// segunda excepción al criterio 4 de plan/02_Backend.md §8, de naturaleza
// distinta a la de internal/devices: ahí el motivo fue un bug de esquema
// heredado de F3; aquí es que este archivo depende directamente de firmas
// que F4 reescribe a propósito.

const testHMACSecret = "integration-test-hmac-secret-32-bytes-min!!"

func newTestServices(t *testing.T) (*syncpkg.Service, *auth.Service, *devices.Service, *levels.Service) {
	t.Helper()
	db := testdb.Setup(t)

	authSvc := auth.NewService(db.Ident, db.Main, auth.Config{
		EncryptionKey:    "integration-test-pgp-key-32-bytes-min!!",
		BlindIndexSecret: []byte("integration-test-blind-index-32-bytes!!"),
		HMACSecret:       []byte(testHMACSecret),
		TokenConfig:      crypto.TokenConfig{Secret: []byte("integration-test-jwt-secret-32-bytes!!!"), AccessExpiry: time.Hour},
	})
	syncSvc := syncpkg.NewService(db.Main, []byte(testHMACSecret))
	devicesSvc := devices.NewService(db.Main)
	levelsSvc := levels.NewService(db.Main)
	return syncSvc, authSvc, devicesSvc, levelsSvc
}

// setupFixtures creates one adult user, one registered device, and one
// published trivia level (difficulty 5, so attempt 1 = 20 XP) — everything
// ProcessSync needs, all through the real services (Register does real
// Argon2id hashing; nothing here is mocked).
func setupFixtures(t *testing.T, ctx context.Context, authSvc *auth.Service, devicesSvc *devices.Service, levelsSvc *levels.Service) (userID, deviceID, levelID uuid.UUID) {
	t.Helper()

	email := "sync-it-" + uuid.NewString() + "@example.test"
	reg, err := authSvc.Register(ctx, auth.RegisterRequest{
		Email: email, Password: "correct horse battery staple",
		IsAdult: true, PrivacyNoticeVersion: "v1.0",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Login (no solo Register) es imprescindible aquí: la réplica accounts
	// [P] se crea recién en el primer Login (plan/02_Backend.md §3.5), y
	// devices.user_id tiene FK a accounts(id) — RegisterDevice fallaría con
	// una violación de FK si se llamara justo después de Register.
	if _, err := authSvc.Login(ctx, auth.LoginRequest{Email: email, Password: "correct horse battery staple"}); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	dev, err := devicesSvc.RegisterDevice(ctx, reg.UserID, devices.RegisterDeviceRequest{
		DeviceKind: "movil", Platform: "web",
	})
	if err != nil {
		t.Fatalf("RegisterDevice() error = %v", err)
	}

	section, err := levelsSvc.CreateSection(ctx, reg.UserID, levels.CreateSectionRequest{
		Title: "IT Section", Description: "integration test", Color: "#18529D",
	})
	if err != nil {
		t.Fatalf("CreateSection() error = %v", err)
	}

	content, _ := json.Marshal(map[string]any{
		"questions": []map[string]any{
			{"question": "2+2?", "options": []string{"3", "4"}, "correct_index": 1},
		},
	})
	level, err := levelsSvc.CreateLevel(ctx, reg.UserID, levels.CreateLevelRequest{
		SectionID: section.ID, Title: "IT Level", Color: "#18529D",
		TemplateType: "trivia", Content: content, Difficulty: 5,
	})
	if err != nil {
		t.Fatalf("CreateLevel() error = %v", err)
	}
	if _, err := levelsSvc.PublishLevel(ctx, reg.UserID, level.ID); err != nil {
		t.Fatalf("PublishLevel() error = %v", err)
	}

	return reg.UserID, dev.ID, level.ID
}

func signedRequest(t *testing.T, userID, deviceID, levelID uuid.UUID, score int, completed bool, attemptDate string) (domain.SyncEventRequest, []byte) {
	t.Helper()
	req := domain.SyncEventRequest{
		SyncEventID:      uuid.New(),
		UserID:           userID,
		DeviceID:         deviceID,
		CryptoKeyVersion: 1,
		Payload: domain.SyncPayload{
			LevelAttempts: []domain.LevelAttemptItem{
				{LevelID: levelID, AttemptDate: attemptDate, AttemptNumber: 1, XPAwarded: 0, Score: score, Completed: completed},
			},
		},
	}
	sig := crypto.GenerateHMAC([]byte(syncpkg.CanonicalSigningPayload(req)), []byte(testHMACSecret))
	return req, sig
}

func TestProcessSyncEndToEnd(t *testing.T) {
	ctx := context.Background()
	syncSvc, authSvc, devicesSvc, levelsSvc := newTestServices(t)
	userID, deviceID, levelID := setupFixtures(t, ctx, authSvc, devicesSvc, levelsSvc)
	today := time.Now().UTC().Format("2006-01-02")

	t.Run("valid sync recalculates XP server-side and stores the reported score", func(t *testing.T) {
		req, sig := signedRequest(t, userID, deviceID, levelID, 77, true, today)

		resp, err := syncSvc.ProcessSync(ctx, req, sig)
		if err != nil {
			t.Fatalf("ProcessSync() error = %v", err)
		}
		if resp.Status != "synced" {
			t.Fatalf("ProcessSync() status = %q, want %q", resp.Status, "synced")
		}
		// difficulty 5, attempt 1 -> 4*5 = 20 XP, regardless of the client's score.
		if resp.ServerXPTotal != 20 {
			t.Fatalf("ProcessSync() server_xp_total = %d, want 20 (server-recalculated, client XP ignored)", resp.ServerXPTotal)
		}
	})

	t.Run("replaying the same sync_event_id is idempotent, not a duplicate award", func(t *testing.T) {
		req, sig := signedRequest(t, userID, deviceID, levelID, 5, true, today)

		first, err := syncSvc.ProcessSync(ctx, req, sig)
		if err != nil {
			t.Fatalf("first ProcessSync() error = %v", err)
		}
		if first.Status != "synced" {
			t.Fatalf("first ProcessSync() status = %q, want synced", first.Status)
		}

		second, err := syncSvc.ProcessSync(ctx, req, sig)
		if err != nil {
			t.Fatalf("replayed ProcessSync() error = %v", err)
		}
		if second.Status != "already_processed" {
			t.Fatalf("replayed ProcessSync() status = %q, want already_processed", second.Status)
		}
	})

	t.Run("tampered signature is rejected", func(t *testing.T) {
		req, sig := signedRequest(t, userID, deviceID, levelID, 5, true, today)
		tamperedSig := append([]byte(nil), sig...)
		tamperedSig[0] ^= 0xFF // flip a bit: same payload, wrong signature

		_, err := syncSvc.ProcessSync(ctx, req, tamperedSig)
		if err == nil {
			t.Fatal("ProcessSync() error = nil, want ErrInvalidSignature")
		}
	})

	t.Run("a future attempt_date is rejected", func(t *testing.T) {
		future := time.Now().UTC().AddDate(0, 0, 3).Format("2006-01-02")
		req, sig := signedRequest(t, userID, deviceID, levelID, 5, true, future)

		_, err := syncSvc.ProcessSync(ctx, req, sig)
		if err == nil {
			t.Fatal("ProcessSync() error = nil, want ErrInvalidPayload (future attempt_date)")
		}
	})
}
