package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

// Sentinel errors — used by the handler for correct HTTP status mapping.
var (
	ErrInvalidSignature = errors.New("invalid HMAC signature")
	ErrInvalidPayload   = errors.New("invalid sync payload")
)

// XP award table per attempt number.
// Base XP is 4*difficulty. Attempt 1 gets 100%, attempts 2-3 get 50%, 4+ gets 0%.
func xpForAttempt(difficulty int32, attemptNumber int64, completed bool) int32 {
	if !completed || difficulty <= 0 {
		return 0
	}
	base := 4 * difficulty
	switch {
	case attemptNumber == 1:
		return base
	case attemptNumber <= 3:
		return base / 2
	default:
		return 0
	}
}

// Service handles the offline-sync business logic.
type Service struct {
	queries    *repository.Queries
	hmacSecret []byte
}

// NewService creates a sync.Service.
func NewService(q *repository.Queries, hmacSecret []byte) *Service {
	if len(hmacSecret) == 0 {
		panic("sync.Service: hmacSecret must not be empty")
	}
	return &Service{queries: q, hmacSecret: hmacSecret}
}

// ProcessSync validates the HMAC, records the sync event, and applies
// the additive merge using server-side recalculation and transactional locks.
// The XP is always recalculated server-side — client-provided xp_awarded is ignored.
//
// sig is the decoded HMAC-SHA256 signature bytes.
func (s *Service) ProcessSync(ctx context.Context, req domain.SyncEventRequest, sig []byte) (domain.SyncEventResponse, error) {
	if err := validateSyncPayload(req); err != nil {
		return domain.SyncEventResponse{}, fmt.Errorf("%w: %s", ErrInvalidPayload, err.Error())
	}

	payloadJSON, err := json.Marshal(req.Payload)
	if err != nil {
		return domain.SyncEventResponse{}, fmt.Errorf("%w: invalid payload JSON", ErrInvalidPayload)
	}
	signingPayload := CanonicalSigningPayload(req)
	isValid := crypto.VerifyHMAC([]byte(signingPayload), sig, s.hmacSecret)

	tx, err := s.queries.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return domain.SyncEventResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.queries.WithTx(tx)

	device, err := qtx.GetActiveDevice(ctx, repository.GetActiveDeviceParams{
		ID:     req.DeviceID,
		UserID: req.UserID,
	})
	if err != nil {
		if repository.IsNoRows(err) {
			return domain.SyncEventResponse{}, ErrInvalidPayload
		}
		return domain.SyncEventResponse{}, err
	}

	// ── Record the sync event regardless of HMAC result (for audit trail) ────
	payloadHash := crypto.GenerateHMAC(payloadJSON, s.hmacSecret)
	err = qtx.InsertSyncEventWithPayload(ctx, repository.InsertSyncEventWithPayloadParams{
		ID:               req.SyncEventID,
		DeviceID:         req.DeviceID,
		UserID:           req.UserID,
		Payload:          payloadJSON,
		PayloadHash:      payloadHash,
		HmacSignature:    sig,
		CryptoKeyVersion: int16(req.CryptoKeyVersion),
		HmacValid:        isValid,
		Status:           "pending",
	})
	if err != nil {
		// Duplicate sync_event_id → idempotent success (already processed).
		if isDuplicateKey(err) {
			return domain.SyncEventResponse{Status: "already_processed"}, nil
		}
		return domain.SyncEventResponse{}, fmt.Errorf("recording sync event: %w", err)
	}

	if !isValid {
		_ = rejectAndCommit(ctx, tx, qtx, req.SyncEventID, "invalid_hmac_signature")
		return domain.SyncEventResponse{}, ErrInvalidSignature
	}

	attempts, streakDates, err := s.validateAndPrepare(ctx, qtx, req)
	if err != nil {
		_ = rejectAndCommit(ctx, tx, qtx, req.SyncEventID, err.Error())
		return domain.SyncEventResponse{}, ErrInvalidPayload
	}

	// ── Process each level attempt with server-side XP calculation ────────────
	for _, attempt := range attempts {
		if err := qtx.LockLevelAttempt(ctx, repository.LockLevelAttemptParams{
			UserID:      req.UserID,
			LevelID:     attempt.LevelID,
			AttemptDate: attempt.Date,
		}); err != nil {
			return domain.SyncEventResponse{}, fmt.Errorf("locking level_attempts: %w", err)
		}

		priorCount, err := qtx.CountLevelAttemptsByDate(ctx, repository.CountLevelAttemptsByDateParams{
			UserID:      req.UserID,
			LevelID:     attempt.LevelID,
			AttemptDate: attempt.Date,
		})
		if err != nil {
			return domain.SyncEventResponse{}, fmt.Errorf("locking level_attempts: %w", err)
		}

		// Server-side XP recalculation — client value is untrusted.
		serverAttemptNumber := priorCount + 1
		serverXP := xpForAttempt(attempt.Difficulty, serverAttemptNumber, attempt.Completed)

		if err := qtx.InsertLevelAttempt(ctx, repository.InsertLevelAttemptParams{
			ID:            uuid.New(),
			UserID:        req.UserID,
			LevelID:       attempt.LevelID,
			AttemptDate:   attempt.Date,
			AttemptNumber: int32(serverAttemptNumber),
			XpAwarded:     serverXP,
			Completed:     attempt.Completed,
		}); err != nil {
			return domain.SyncEventResponse{}, fmt.Errorf("inserting level_attempt: %w", err)
		}

		if err := qtx.UpsertPlayerProgressForAttempt(ctx, repository.UpsertPlayerProgressForAttemptParams{
			UserID:  req.UserID,
			LevelID: attempt.LevelID,
			// best_score is the in-game score (client-reported, validated >= 0),
			// mirroring the online path. XP total stays server-recalculated.
			BestScore:       attempt.Score,
			XpTotalForLevel: serverXP,
			Completed:       attempt.Completed,
		}); err != nil {
			return domain.SyncEventResponse{}, fmt.Errorf("upserting player_progress: %w", err)
		}

		eventType := "sync_failed_attempt"
		if attempt.Completed {
			eventType = "sync_level_completed"
		}
		if err := qtx.InsertExperienceHistory(ctx, repository.InsertExperienceHistoryParams{
			ID:                 uuid.New(),
			UserID:             uuid.NullUUID{UUID: req.UserID, Valid: true},
			LevelID:            attempt.LevelID,
			EventType:          eventType,
			XpGained:           serverXP,
			Source:             "offline_sync",
			VerificationMethod: "hmac_offline",
			SyncEventID:        uuid.NullUUID{UUID: req.SyncEventID, Valid: true},
		}); err != nil {
			return domain.SyncEventResponse{}, fmt.Errorf("inserting experience_history: %w", err)
		}
	}

	// ── Daily streak entries ──────────────────────────────────────────────────
	for _, date := range streakDates {
		if err := qtx.UpsertDailyStreak(ctx, repository.UpsertDailyStreakParams{
			UserID:       req.UserID,
			ActivityDate: date,
		}); err != nil {
			return domain.SyncEventResponse{}, fmt.Errorf("upserting daily streak: %w", err)
		}
	}

	if err := qtx.UpdateSyncEventStatus(ctx, repository.UpdateSyncEventStatusParams{
		ID:     req.SyncEventID,
		Status: "processed",
	}); err != nil {
		return domain.SyncEventResponse{}, fmt.Errorf("updating sync event status: %w", err)
	}
	
	if err := qtx.TouchDevice(ctx, req.DeviceID); err != nil {
		return domain.SyncEventResponse{}, fmt.Errorf("touching device: %w", err)
	}
	totals, err := qtx.GetUserProgressTotals(ctx, req.UserID)
	if err != nil {
		return domain.SyncEventResponse{}, err
	}
	badgesAwarded, err := qtx.AwardEligibleBadges(ctx, repository.AwardEligibleBadgesParams{
		UserID:  req.UserID,
		TotalXP: totals.TotalXP,
	})
	if err != nil {
		return domain.SyncEventResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.SyncEventResponse{}, err
	}

	return domain.SyncEventResponse{
		Status:        "synced",
		WipeLocalData: device.WipeLocalData,
		ServerXPTotal: int(totals.TotalXP),
		BadgesAwarded: badgeIDs(badgesAwarded),
	}, nil
}

type preparedAttempt struct {
	LevelID    uuid.UUID
	Date       time.Time
	Difficulty int32
	Score      int32
	Completed  bool
}

func (s *Service) validateAndPrepare(ctx context.Context, q *repository.Queries, req domain.SyncEventRequest) ([]preparedAttempt, []time.Time, error) {
	today := dateOnly(time.Now().UTC())
	attempts := make([]preparedAttempt, 0, len(req.Payload.LevelAttempts))
	for _, attempt := range req.Payload.LevelAttempts {
		date, err := time.Parse("2006-01-02", attempt.AttemptDate)
		if err != nil {
			return nil, nil, errors.New("invalid_attempt_date")
		}
		date = dateOnly(date)
		if date.After(today) {
			return nil, nil, errors.New("future_attempt_date")
		}
		if attempt.Score < 0 {
			return nil, nil, errors.New("invalid_score")
		}
		level, err := q.GetLevelByID(ctx, attempt.LevelID)
		if err != nil || !level.IsPublished {
			return nil, nil, errors.New("level_not_available")
		}
		attempts = append(attempts, preparedAttempt{
			LevelID:    attempt.LevelID,
			Date:       date,
			Difficulty: level.Difficulty,
			Score:      int32(attempt.Score),
			Completed:  attempt.Completed,
		})
	}

	streakDates := make([]time.Time, 0, len(req.Payload.DailyStreakDates))
	for _, rawDate := range req.Payload.DailyStreakDates {
		date, err := time.Parse("2006-01-02", rawDate)
		if err != nil {
			return nil, nil, errors.New("invalid_streak_date")
		}
		date = dateOnly(date)
		if date.After(today) {
			return nil, nil, errors.New("future_streak_date")
		}
		streakDates = append(streakDates, date)
	}
	return attempts, streakDates, nil
}

func rejectAndCommit(ctx context.Context, tx *sql.Tx, q *repository.Queries, syncEventID uuid.UUID, reason string) error {
	if err := q.UpdateSyncEventRejected(ctx, repository.UpdateSyncEventRejectedParams{
		ID:     syncEventID,
		Reason: reason,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func CanonicalSigningPayload(req domain.SyncEventRequest) string {
	attempts := make([]string, 0, len(req.Payload.LevelAttempts))
	for _, attempt := range req.Payload.LevelAttempts {
		attempts = append(attempts, strings.Join([]string{
			attempt.LevelID.String(),
			attempt.AttemptDate,
			strconv.Itoa(attempt.AttemptNumber),
			strconv.Itoa(attempt.XPAwarded),
			strconv.Itoa(attempt.Score),
			strconv.FormatBool(attempt.Completed),
		}, ","))
	}
	sort.Strings(attempts)

	streakDates := append([]string(nil), req.Payload.DailyStreakDates...)
	sort.Strings(streakDates)

	badges := make([]string, 0, len(req.Payload.BadgeIDsEarned))
	for _, badgeID := range req.Payload.BadgeIDsEarned {
		badges = append(badges, badgeID.String())
	}
	sort.Strings(badges)

	return strings.Join([]string{
		req.SyncEventID.String(),
		req.UserID.String(),
		req.DeviceID.String(),
		strconv.Itoa(req.CryptoKeyVersion),
		strings.Join(attempts, ";"),
		strings.Join(streakDates, ";"),
		strings.Join(badges, ";"),
	}, "|")
}

func dateOnly(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func badgeIDs(rows []repository.BadgeWithEarnedAt) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids
}

// isDuplicateKey checks for PostgreSQL unique constraint violation.
func isDuplicateKey(err error) bool {
	return err != nil && (containsAny(err.Error(), "duplicate key", "23505"))
}

func containsAny(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func validateSyncPayload(req domain.SyncEventRequest) error {
	if req.SyncEventID == uuid.Nil {
		return errors.New("sync_event_id is required")
	}
	if req.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if req.DeviceID == uuid.Nil {
		return errors.New("device_id is required")
	}
	return nil
}
