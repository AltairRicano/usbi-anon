// PlayerService agrupa las operaciones para jugar y consultar el progreso. Nunca ve contenido no publicado.
package levels

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

type PlayerService struct {
	repo *repository.Queries
}

func NewPlayerService(repo *repository.Queries) *PlayerService {
	return &PlayerService{repo: repo}
}

// GetLevel nunca devuelve un nivel no publicado.
func (s *PlayerService) GetLevel(ctx context.Context, levelID uuid.UUID) (LevelResponse, error) {
	if levelID == uuid.Nil {
		return LevelResponse{}, ErrValidation
	}
	level, err := s.repo.GetLevelByID(ctx, levelID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LevelResponse{}, ErrNotFound
		}
		return LevelResponse{}, err
	}
	if !level.IsPublished {
		return LevelResponse{}, ErrNotFound
	}
	return levelToResponse(level), nil
}

// ListLevels siempre filtra a solo publicados.
func (s *PlayerService) ListLevels(ctx context.Context, cursor uuid.UUID, sectionID uuid.UUID, pageSize int32) (LevelsPage, error) {
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	rows, err := s.repo.ListLevels(ctx, repository.ListLevelsParams{
		IncludeUnpublished: false,
		HasSectionID:       sectionID != uuid.Nil,
		SectionID:          sectionID,
		Cursor:             cursor,
		PageSize:           pageSize + 1,
	})
	if err != nil {
		return LevelsPage{}, err
	}

	items := make([]LevelSummary, 0, len(rows))
	for _, r := range rows {
		items = append(items, LevelSummary{
			ID:           r.ID,
			SectionID:    r.SectionID,
			Title:        r.Title,
			Color:        r.Color,
			TemplateType: r.TemplateType,
			Difficulty:   r.Difficulty,
			IsPublished:  r.IsPublished,
			CreatedAt:    r.CreatedAt,
		})
	}

	page := LevelsPage{}
	if len(items) > int(pageSize) {
		items = items[:pageSize]
		page.NextCursor = items[pageSize-1].ID.String()
	}
	page.Items = items

	return page, nil
}

// ListSections siempre filtra a solo publicadas.
func (s *PlayerService) ListSections(ctx context.Context) (SectionsResponse, error) {
	sections, err := s.repo.ListSections(ctx, repository.ListSectionsParams{IncludeUnpublished: false})
	if err != nil {
		return SectionsResponse{}, err
	}
	items := make([]SectionResponse, 0, len(sections))
	for _, section := range sections {
		items = append(items, sectionToResponse(section))
	}
	return SectionsResponse{Items: items}, nil
}

func (s *PlayerService) CompleteLevel(ctx context.Context, userID, levelID uuid.UUID, req CompleteLevelRequest) (CompleteLevelResponse, error) {
	if userID == uuid.Nil || levelID == uuid.Nil || req.Score < 0 {
		return CompleteLevelResponse{}, ErrValidation
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return CompleteLevelResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := s.repo.WithTx(tx)
	level, err := qtx.GetLevelByID(ctx, levelID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CompleteLevelResponse{}, ErrNotFound
		}
		return CompleteLevelResponse{}, err
	}
	if !level.IsPublished {
		return CompleteLevelResponse{}, ErrNotFound
	}

	// Recalcula completed/score en el servidor cuando es posible (verify.go), si no confía en el cliente.
	completed := req.Completed
	score := req.Score
	verificationMethod := domain.VerificationOnlineReported
	if verified, ok := verifyAnswers(level.TemplateType, level.Content, req.Answers); ok {
		completed = verified.completed
		score = verified.score
		verificationMethod = domain.VerificationOnlineVerified
	}

	attemptDate := completionDate(req.ClientFinishedAt)
	if err := qtx.LockLevelAttempt(ctx, repository.LockLevelAttemptParams{
		UserID:      userID,
		LevelID:     levelID,
		AttemptDate: attemptDate,
	}); err != nil {
		return CompleteLevelResponse{}, fmt.Errorf("locking attempt: %w", err)
	}

	priorAttempts, err := qtx.CountLevelAttemptsByDate(ctx, repository.CountLevelAttemptsByDateParams{
		UserID:      userID,
		LevelID:     levelID,
		AttemptDate: attemptDate,
	})
	if err != nil {
		return CompleteLevelResponse{}, err
	}

	attemptNumber := int32(priorAttempts + 1)
	xpAwarded := CalculateXP(level.Difficulty, attemptNumber, completed)

	if err := qtx.InsertLevelAttempt(ctx, repository.InsertLevelAttemptParams{
		ID:            newID(),
		UserID:        userID,
		LevelID:       levelID,
		AttemptDate:   attemptDate,
		AttemptNumber: attemptNumber,
		XpAwarded:     xpAwarded,
		Completed:     completed,
	}); err != nil {
		return CompleteLevelResponse{}, err
	}

	if err := qtx.UpsertPlayerProgressForAttempt(ctx, repository.UpsertPlayerProgressForAttemptParams{
		UserID:          userID,
		LevelID:         levelID,
		BestScore:       score,
		XpTotalForLevel: xpAwarded,
		Completed:       completed,
	}); err != nil {
		return CompleteLevelResponse{}, err
	}

	if completed {
		if err := qtx.UpsertDailyStreak(ctx, repository.UpsertDailyStreakParams{
			UserID:       userID,
			ActivityDate: attemptDate,
		}); err != nil {
			return CompleteLevelResponse{}, err
		}
	}

	eventType := "level_failed"
	if completed {
		eventType = "level_completed"
	}
	if err := qtx.InsertExperienceHistory(ctx, repository.InsertExperienceHistoryParams{
		ID:                 newID(),
		UserID:             uuid.NullUUID{UUID: userID, Valid: true},
		LevelID:            levelID,
		EventType:          eventType,
		XpGained:           xpAwarded,
		Source:             "online",
		VerificationMethod: string(verificationMethod),
		SyncEventID:        uuid.NullUUID{},
	}); err != nil {
		return CompleteLevelResponse{}, err
	}

	totals, err := qtx.GetUserProgressTotals(ctx, userID)
	if err != nil {
		return CompleteLevelResponse{}, err
	}
	badgesAwarded, err := qtx.AwardEligibleBadges(ctx, repository.AwardEligibleBadgesParams{
		UserID:  userID,
		TotalXP: totals.TotalXP,
	})
	if err != nil {
		return CompleteLevelResponse{}, err
	}
	streakDates, err := qtx.ListDailyStreakDates(ctx, userID, 370)
	if err != nil {
		return CompleteLevelResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return CompleteLevelResponse{}, err
	}

	return CompleteLevelResponse{
		LevelID:       levelID,
		Completed:     completed,
		AttemptNumber: attemptNumber,
		XPAwarded:     xpAwarded,
		TotalXP:       totals.TotalXP,
		CurrentStreak: calculateCurrentStreak(streakDates),
		BadgesAwarded: badgesToResponse(badgesAwarded),
	}, nil
}

func (s *PlayerService) GetProfileProgress(ctx context.Context, userID uuid.UUID) (ProfileProgressResponse, error) {
	if userID == uuid.Nil {
		return ProfileProgressResponse{}, ErrValidation
	}
	totals, err := s.repo.GetUserProgressTotals(ctx, userID)
	if err != nil {
		return ProfileProgressResponse{}, err
	}
	streakDates, err := s.repo.ListDailyStreakDates(ctx, userID, 370)
	if err != nil {
		return ProfileProgressResponse{}, err
	}
	badgeRows, err := s.repo.ListUserBadges(ctx, userID)
	if err != nil {
		return ProfileProgressResponse{}, err
	}
	progressRows, err := s.repo.ListUserProgressLevels(ctx, userID)
	if err != nil {
		return ProfileProgressResponse{}, err
	}

	levels := make([]ProgressLevelResponse, 0, len(progressRows))
	for _, row := range progressRows {
		item := ProgressLevelResponse{
			LevelID:         row.LevelID,
			Title:           row.Title,
			TemplateType:    row.TemplateType,
			Difficulty:      row.Difficulty,
			BestScore:       row.BestScore,
			XPTotalForLevel: row.XpTotalForLevel,
			AttemptsCount:   row.AttemptsCount,
		}
		if row.FirstCompletedAt.Valid {
			t := row.FirstCompletedAt.Time
			item.FirstCompletedAt = &t
		}
		if row.LastCompletedAt.Valid {
			t := row.LastCompletedAt.Time
			item.LastCompletedAt = &t
		}
		levels = append(levels, item)
	}

	return ProfileProgressResponse{
		TotalXP:         totals.TotalXP,
		CompletedLevels: totals.CompletedLevels,
		TotalAttempts:   totals.TotalAttempts,
		CurrentStreak:   calculateCurrentStreak(streakDates),
		Badges:          badgesToResponse(badgeRows),
		Levels:          levels,
	}, nil
}
