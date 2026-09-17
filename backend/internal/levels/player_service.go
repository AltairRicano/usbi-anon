// PlayerService agrupa las operaciones de internal/levels que corren con el
// pool de usbi_app (F3, 2026-09-09): jugar y consultar el propio progreso.
// Nunca ve contenido no publicado — a diferencia de AdminService (
// admin_service.go), no recibe ningún parámetro para pedirlo, así que un bug
// en el caller no puede colar un include_unpublished=true por esta vía. (Útil)
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

// GetLevel nunca devuelve un nivel no publicado — a diferencia de
// AdminService.GetLevel, no acepta includeUnpublished: el jugador jamás
// necesita verlo, así que ni siquiera existe la posibilidad de pasarlo por
// error. (Útil)
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

// ListLevels siempre filtra a solo publicados — sin parámetro
// includeUnpublished, a diferencia de AdminService.ListLevels. (Útil)
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

// ListSections siempre filtra a solo publicadas — sin parámetro
// includeUnpublished, a diferencia de AdminService.ListSections. (Útil)
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

	// M1 Fase B (D-03): el servidor recalcula completed/score contra
	// level.content cuando la plantilla y las respuestas enviadas lo permiten
	// (verify.go); si no, cae al comportamiento anterior (confiar en el
	// cliente) y lo deja registrado como tal en verification_method.
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
