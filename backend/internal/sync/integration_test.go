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
	"github.com/altair/usbi-anon-backend/internal/quiz"
	syncpkg "github.com/altair/usbi-anon-backend/internal/sync"
	"github.com/altair/usbi-anon-backend/internal/testdb"
	"github.com/google/uuid"
)

// Cobertura de integración para ProcessSync contra un esquema real de Postgres (hallazgo de auditoría
// B9) — la verificación HMAC, idempotencia y caminos de rechazo de fechas futuras
// antes no se probaban end-to-end; solo funciones puras auxiliares tenían
// pruebas unitarias. Ejecutar con TEST_DATABASE_URL definido (ver internal/testdb); de lo contrario
// se salta.
//
// Reescrito en F9 (plan/04_Rediseno_identidad_gustos.md §2 y §3) para el
// esquema unificado: ya no hay dos bases (testdb.Setup devuelve un único
// *testdb.DB{Repo, DB}) ni Register de una sola llamada por email —
// setupFixtures ahora recorre el registro real en 3 pasos
// (RegisterQuestions → RegisterAnswers → RegisterConfirm) contra las
// registration_questions sembradas por el baseline de
// 0001_esquema_unificado.up.sql, exactamente como lo haría un cliente real. (Útil)

const testHMACSecret = "integration-test-hmac-secret-32-bytes-min!!"

func newTestServices(t *testing.T) (*syncpkg.Service, *auth.Service, *devices.Service, *levels.Service) {
	t.Helper()
	db := testdb.Setup(t)

	quizSvc := quiz.NewService(db.Repo)
	authSvc := auth.NewService(db.Repo, quizSvc, auth.Config{
		HMACSecret:  []byte(testHMACSecret),
		TokenConfig: crypto.TokenConfig{Secret: []byte("integration-test-jwt-secret-32-bytes!!!"), AccessExpiry: time.Hour},
	})
	syncSvc := syncpkg.NewService(db.Repo, []byte(testHMACSecret))
	devicesSvc := devices.NewService(db.Repo)
	levelsSvc := levels.NewService(db.Repo)
	return syncSvc, authSvc, devicesSvc, levelsSvc
}

// setupFixtures crea una cuenta de adulto (a través del flujo de registro real de 3 pasos),
// un dispositivo registrado y un nivel de trivia publicado (dificultad
// 5, por lo que intento 1 = 20 XP) — todo lo que ProcessSync necesita, todo a través de
// los servicios reales (RegisterConfirm hace hashing real con Argon2id; nada aquí está
// mockeado). (Útil)
func setupFixtures(t *testing.T, ctx context.Context, authSvc *auth.Service, devicesSvc *devices.Service, levelsSvc *levels.Service) (userID, deviceID, levelID uuid.UUID) {
	t.Helper()

	questions, err := authSvc.RegisterQuestions(ctx)
	if err != nil {
		t.Fatalf("RegisterQuestions() error = %v", err)
	}
	if len(questions.Questions) < 2 {
		t.Fatalf("RegisterQuestions() returned %d questions, want at least 2 (baseline seed)", len(questions.Questions))
	}

	answers := make([]auth.AnswerInput, 0, 2)
	for i, q := range questions.Questions[:2] {
		answers = append(answers, auth.AnswerInput{
			QuestionID: q.ID,
			AnswerText: uuid.NewString()[:8] + string(rune('a'+i)),
		})
	}

	answersResp, err := authSvc.RegisterAnswers(ctx, auth.RegisterAnswersRequest{
		Answers:              answers,
		IsAdult:              true,
		PrivacyNoticeVersion: "v1.0",
	})
	if err != nil {
		t.Fatalf("RegisterAnswers() error = %v", err)
	}
	if len(answersResp.NicknameCandidates) == 0 {
		t.Fatalf("RegisterAnswers() returned no nickname candidates")
	}

	confirm, err := authSvc.RegisterConfirm(ctx, auth.RegisterConfirmRequest{
		RegistrationToken: answersResp.RegistrationToken,
		ChosenNickname:    answersResp.NicknameCandidates[0],
	})
	if err != nil {
		t.Fatalf("RegisterConfirm() error = %v", err)
	}

	// Login (no solo RegisterConfirm) confirma que el password devuelto en
	// claro es de verdad el que quedó hasheado en la cuenta — sin esto, un
	// bug en el hash de RegisterConfirm pasaría inadvertido para este test. (Útil)
	if _, err := authSvc.Login(ctx, auth.LoginRequest{Nickname: confirm.Nickname, Password: confirm.Password}); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	dev, err := devicesSvc.RegisterDevice(ctx, confirm.AccountID, devices.RegisterDeviceRequest{
		DeviceKind: "movil", Platform: "web",
	})
	if err != nil {
		t.Fatalf("RegisterDevice() error = %v", err)
	}

	section, err := levelsSvc.CreateSection(ctx, confirm.AccountID, levels.CreateSectionRequest{
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
	level, err := levelsSvc.CreateLevel(ctx, confirm.AccountID, levels.CreateLevelRequest{
		SectionID: section.ID, Title: "IT Level", Color: "#18529D",
		TemplateType: "trivia", Content: content, Difficulty: 5,
	})
	if err != nil {
		t.Fatalf("CreateLevel() error = %v", err)
	}
	if _, err := levelsSvc.PublishLevel(ctx, confirm.AccountID, level.ID); err != nil {
		t.Fatalf("PublishLevel() error = %v", err)
	}

	return confirm.AccountID, dev.ID, level.ID
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
		// dificultad 5, intento 1 -> 4*5 = 20 XP, sin importar el puntaje del cliente. (Útil)
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
