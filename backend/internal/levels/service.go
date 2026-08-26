package levels

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/altair/usbi-anon-backend/internal/audit"
	"github.com/altair/usbi-anon-backend/internal/httpjson"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
	ErrForbidden  = errors.New("forbidden")
)

type Service struct {
	repo *repository.Queries
}

func NewService(repo *repository.Queries) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateLevel(ctx context.Context, adminID uuid.UUID, req CreateLevelRequest) (LevelResponse, error) {
	if req.SectionID == uuid.Nil {
		return LevelResponse{}, ErrValidation
	}
	if err := validateLevelInput(req.Title, req.Color, req.TemplateType, req.Difficulty, req.Content); err != nil {
		return LevelResponse{}, ErrValidation
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return LevelResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	levelID := newID()

	adminUUID := uuid.NullUUID{UUID: adminID, Valid: true}

	level, err := qtx.CreateLevelReturning(ctx, repository.CreateLevelReturningParams{
		ID:               levelID,
		SectionID:        req.SectionID,
		Title:            req.Title,
		Color:            req.Color,
		TemplateType:     req.TemplateType,
		Content:          req.Content,
		Difficulty:       req.Difficulty,
		IsPublished:      false,
		CreatedByAdminID: adminUUID,
	})
	if err != nil {
		return LevelResponse{}, err
	}
	resp := levelToResponse(level)

	if err := logAdminAudit(ctx, AuditParams{
		Repo:        qtx,
		ActorID:     adminID,
		Action:      "level.create",
		EntityType:  "level",
		EntityID:    level.ID,
		BeforeState: nil,
		AfterState:  levelAuditPayload(resp),
	}); err != nil {
		return LevelResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return LevelResponse{}, err
	}
	return resp, nil
}

func (s *Service) UpdateLevel(ctx context.Context, adminID, levelID uuid.UUID, req UpdateLevelRequest) (LevelResponse, error) {
	if levelID == uuid.Nil {
		return LevelResponse{}, ErrValidation
	}
	if err := validateLevelInput(req.Title, req.Color, req.TemplateType, req.Difficulty, req.Content); err != nil {
		return LevelResponse{}, ErrValidation
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return LevelResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	level, err := qtx.UpdateLevel(ctx, repository.UpdateLevelParams{
		ID:           levelID,
		Title:        req.Title,
		Color:        req.Color,
		TemplateType: req.TemplateType,
		Content:      req.Content,
		Difficulty:   req.Difficulty,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LevelResponse{}, ErrNotFound
		}
		return LevelResponse{}, err
	}
	resp := levelToResponse(level)
	if err := logAdminAudit(ctx, AuditParams{
		Repo:        qtx,
		ActorID:     adminID,
		Action:      "level.update",
		EntityType:  "level",
		EntityID:    level.ID,
		BeforeState: nil,
		AfterState:  levelAuditPayload(resp),
	}); err != nil {
		return LevelResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return LevelResponse{}, err
	}
	return resp, nil
}

func (s *Service) PublishLevel(ctx context.Context, adminID, levelID uuid.UUID) (LevelResponse, error) {
	if levelID == uuid.Nil {
		return LevelResponse{}, ErrValidation
	}
	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return LevelResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	level, err := qtx.PublishLevel(ctx, levelID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LevelResponse{}, ErrNotFound
		}
		return LevelResponse{}, err
	}
	resp := levelToResponse(level)
	if err := logAdminAudit(ctx, AuditParams{
		Repo:        qtx,
		ActorID:     adminID,
		Action:      "level.publish",
		EntityType:  "level",
		EntityID:    level.ID,
		BeforeState: nil,
		AfterState:  levelAuditPayload(resp),
	}); err != nil {
		return LevelResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return LevelResponse{}, err
	}
	return resp, nil
}

func (s *Service) UnpublishLevel(ctx context.Context, adminID, levelID uuid.UUID) (LevelResponse, error) {
	if levelID == uuid.Nil {
		return LevelResponse{}, ErrValidation
	}
	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return LevelResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	level, err := qtx.UnpublishLevel(ctx, levelID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LevelResponse{}, ErrNotFound
		}
		return LevelResponse{}, err
	}
	resp := levelToResponse(level)
	if err := logAdminAudit(ctx, AuditParams{
		Repo:        qtx,
		ActorID:     adminID,
		Action:      "level.unpublish",
		EntityType:  "level",
		EntityID:    level.ID,
		BeforeState: nil,
		AfterState:  levelAuditPayload(resp),
	}); err != nil {
		return LevelResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return LevelResponse{}, err
	}
	return resp, nil
}

func (s *Service) ArchiveLevel(ctx context.Context, adminID, levelID uuid.UUID) (LevelResponse, error) {
	if adminID == uuid.Nil || levelID == uuid.Nil {
		return LevelResponse{}, ErrValidation
	}
	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return LevelResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	level, err := qtx.ArchiveLevel(ctx, repository.ArchiveLevelParams{
		ID:        levelID,
		DeletedBy: uuid.NullUUID{UUID: adminID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LevelResponse{}, ErrNotFound
		}
		return LevelResponse{}, err
	}
	resp := levelToResponse(level)
	if err := logAdminAudit(ctx, AuditParams{
		Repo:        qtx,
		ActorID:     adminID,
		Action:      "level.archive",
		EntityType:  "level",
		EntityID:    level.ID,
		BeforeState: nil,
		AfterState:  levelAuditPayload(resp),
	}); err != nil {
		return LevelResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return LevelResponse{}, err
	}
	return resp, nil
}

// ErrNotArchived se devuelve cuando se intenta purgar o restaurar contenido
// que no ha pasado primero por archivar — la purga irreversible solo opera
// sobre lo ya archivado (plan/05_Contenido_maker_y_juego.md §6).
var ErrNotArchived = errors.New("content is not archived")

// ErrSectionHasLevels se devuelve cuando se intenta purgar una sección que
// todavía tiene niveles (archivados o no) referenciándola. levels.section_id
// es ON DELETE RESTRICT a propósito: cada nivel se purga uno por uno, con sus
// contadores acumulados, nunca en cascada silenciosa.
var ErrSectionHasLevels = errors.New("section still has levels")

func (s *Service) UnarchiveLevel(ctx context.Context, adminID, levelID uuid.UUID) (LevelResponse, error) {
	if levelID == uuid.Nil {
		return LevelResponse{}, ErrValidation
	}
	current, err := s.repo.GetLevelByIDAny(ctx, levelID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LevelResponse{}, ErrNotFound
		}
		return LevelResponse{}, err
	}
	if !current.DeletedAt.Valid {
		return LevelResponse{}, ErrNotArchived
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return LevelResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	level, err := qtx.UnarchiveLevel(ctx, levelID)
	if err != nil {
		return LevelResponse{}, err
	}
	resp := levelToResponse(level)
	if err := logAdminAudit(ctx, AuditParams{
		Repo: qtx, ActorID: adminID, Action: "level.unarchive",
		EntityType: "level", EntityID: level.ID, AfterState: levelAuditPayload(resp),
	}); err != nil {
		return LevelResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return LevelResponse{}, err
	}
	return resp, nil
}

func (s *Service) ListArchivedLevels(ctx context.Context, sectionID uuid.UUID) (ArchivedLevelsResponse, error) {
	rows, err := s.repo.ListArchivedLevels(ctx, repository.ListArchivedLevelsParams{
		HasSectionID: sectionID != uuid.Nil,
		SectionID:    sectionID,
	})
	if err != nil {
		return ArchivedLevelsResponse{}, err
	}
	items := make([]LevelSummary, 0, len(rows))
	for _, r := range rows {
		items = append(items, LevelSummary{
			ID: r.ID, SectionID: r.SectionID, Title: r.Title, Color: r.Color,
			TemplateType: r.TemplateType, Difficulty: r.Difficulty,
			IsPublished: r.IsPublished, CreatedAt: r.CreatedAt,
		})
	}
	return ArchivedLevelsResponse{Items: items}, nil
}

// PurgeLevel es irreversible. Solo opera sobre un nivel ya archivado: acumula
// sus contadores de progreso en account_retired_progress ANTES del DELETE
// (dentro de la misma transacción, orden que importa) para que la XP y los
// "niveles completados" de cada jugador sobrevivan a la rotación de
// temporada — ver CLAUDE.md, "rotación de niveles por temporadas".
func (s *Service) PurgeLevel(ctx context.Context, adminID, levelID uuid.UUID) error {
	if levelID == uuid.Nil {
		return ErrValidation
	}
	current, err := s.repo.GetLevelByIDAny(ctx, levelID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if !current.DeletedAt.Valid {
		return ErrNotArchived
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	if err := qtx.AccumulateRetiredProgressForLevel(ctx, levelID); err != nil {
		return fmt.Errorf("accumulating retired progress: %w", err)
	}
	rowsAffected, err := qtx.PurgeLevel(ctx, levelID)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	beforeState := levelAuditPayload(levelToResponse(current))
	if err := logAdminAudit(ctx, AuditParams{
		Repo: qtx, ActorID: adminID, Action: "level.purge",
		EntityType: "level", EntityID: levelID, BeforeState: beforeState,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) UnarchiveSection(ctx context.Context, adminID, sectionID uuid.UUID) (SectionResponse, error) {
	if sectionID == uuid.Nil {
		return SectionResponse{}, ErrValidation
	}
	current, err := s.repo.GetSectionByIDAny(ctx, sectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SectionResponse{}, ErrNotFound
		}
		return SectionResponse{}, err
	}
	if !current.ArchivedAt.Valid {
		return SectionResponse{}, ErrNotArchived
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return SectionResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	section, err := qtx.UnarchiveSection(ctx, sectionID)
	if err != nil {
		return SectionResponse{}, err
	}
	resp := sectionToResponse(section)
	if err := logAdminAudit(ctx, AuditParams{
		Repo: qtx, ActorID: adminID, Action: "section.unarchive",
		EntityType: "section", EntityID: section.ID, AfterState: resp,
	}); err != nil {
		return SectionResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return SectionResponse{}, err
	}
	return resp, nil
}

func (s *Service) ListArchivedSections(ctx context.Context) (ArchivedSectionsResponse, error) {
	sections, err := s.repo.ListArchivedSections(ctx)
	if err != nil {
		return ArchivedSectionsResponse{}, err
	}
	items := make([]SectionResponse, 0, len(sections))
	for _, section := range sections {
		items = append(items, sectionToResponse(section))
	}
	return ArchivedSectionsResponse{Items: items}, nil
}

// PurgeSection es irreversible. Solo opera sobre una sección ya archivada y
// sin ningún nivel restante (archivado o no) referenciándola — levels.
// section_id es RESTRICT a propósito, así que primero hay que purgar sus
// niveles uno por uno.
func (s *Service) PurgeSection(ctx context.Context, adminID, sectionID uuid.UUID) error {
	if sectionID == uuid.Nil {
		return ErrValidation
	}
	current, err := s.repo.GetSectionByIDAny(ctx, sectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if !current.ArchivedAt.Valid {
		return ErrNotArchived
	}
	remaining, err := s.repo.CountLevelsBySection(ctx, sectionID)
	if err != nil {
		return err
	}
	if remaining > 0 {
		return ErrSectionHasLevels
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	rowsAffected, err := qtx.PurgeSection(ctx, sectionID)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	if err := logAdminAudit(ctx, AuditParams{
		Repo: qtx, ActorID: adminID, Action: "section.purge",
		EntityType: "section", EntityID: sectionID, BeforeState: sectionToResponse(current),
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) GetLevel(ctx context.Context, levelID uuid.UUID, includeUnpublished bool) (LevelResponse, error) {
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
	if !includeUnpublished && !level.IsPublished {
		return LevelResponse{}, ErrNotFound
	}
	return levelToResponse(level), nil
}

func (s *Service) ListLevels(ctx context.Context, cursor uuid.UUID, sectionID uuid.UUID, includeUnpublished bool, pageSize int32) (LevelsPage, error) {
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	rows, err := s.repo.ListLevels(ctx, repository.ListLevelsParams{
		IncludeUnpublished: includeUnpublished,
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

func (s *Service) CreateSection(ctx context.Context, adminID uuid.UUID, req CreateSectionRequest) (SectionResponse, error) {
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Color) == "" {
		return SectionResponse{}, ErrValidation
	}
	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return SectionResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	section, err := qtx.CreateSectionReturning(ctx, repository.CreateSectionReturningParams{
		ID:               newID(),
		Title:            strings.TrimSpace(req.Title),
		Description:      strings.TrimSpace(req.Description),
		Color:            strings.TrimSpace(req.Color),
		IsPublished:      false,
		CreatedByAdminID: uuid.NullUUID{UUID: adminID, Valid: true},
	})
	if err != nil {
		return SectionResponse{}, err
	}
	resp := sectionToResponse(section)
	if err := logAdminAudit(ctx, AuditParams{
		Repo:        qtx,
		ActorID:     adminID,
		Action:      "section.create",
		EntityType:  "section",
		EntityID:    section.ID,
		BeforeState: nil,
		AfterState:  resp,
	}); err != nil {
		return SectionResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return SectionResponse{}, err
	}
	return resp, nil
}

func (s *Service) ListSections(ctx context.Context, includeUnpublished bool) (SectionsResponse, error) {
	sections, err := s.repo.ListSections(ctx, repository.ListSectionsParams{IncludeUnpublished: includeUnpublished})
	if err != nil {
		return SectionsResponse{}, err
	}
	items := make([]SectionResponse, 0, len(sections))
	for _, section := range sections {
		items = append(items, sectionToResponse(section))
	}
	return SectionsResponse{Items: items}, nil
}

func (s *Service) UpdateSection(ctx context.Context, adminID, sectionID uuid.UUID, req UpdateSectionRequest) (SectionResponse, error) {
	if sectionID == uuid.Nil || strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Color) == "" {
		return SectionResponse{}, ErrValidation
	}
	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return SectionResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	section, err := qtx.UpdateSection(ctx, repository.UpdateSectionParams{
		ID:          sectionID,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Color:       strings.TrimSpace(req.Color),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SectionResponse{}, ErrNotFound
		}
		return SectionResponse{}, err
	}
	resp := sectionToResponse(section)
	if err := logAdminAudit(ctx, AuditParams{
		Repo:        qtx,
		ActorID:     adminID,
		Action:      "section.update",
		EntityType:  "section",
		EntityID:    section.ID,
		BeforeState: nil,
		AfterState:  resp,
	}); err != nil {
		return SectionResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return SectionResponse{}, err
	}
	return resp, nil
}

func (s *Service) PublishSection(ctx context.Context, adminID, sectionID uuid.UUID) (SectionResponse, error) {
	if sectionID == uuid.Nil {
		return SectionResponse{}, ErrValidation
	}
	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return SectionResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	section, err := qtx.PublishSection(ctx, sectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SectionResponse{}, ErrNotFound
		}
		return SectionResponse{}, err
	}
	resp := sectionToResponse(section)
	if err := logAdminAudit(ctx, AuditParams{
		Repo:        qtx,
		ActorID:     adminID,
		Action:      "section.publish",
		EntityType:  "section",
		EntityID:    section.ID,
		BeforeState: nil,
		AfterState:  resp,
	}); err != nil {
		return SectionResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return SectionResponse{}, err
	}
	return resp, nil
}

func (s *Service) UnpublishSection(ctx context.Context, adminID, sectionID uuid.UUID) (SectionResponse, error) {
	if sectionID == uuid.Nil {
		return SectionResponse{}, ErrValidation
	}
	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return SectionResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	section, err := qtx.UnpublishSection(ctx, sectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SectionResponse{}, ErrNotFound
		}
		return SectionResponse{}, err
	}
	resp := sectionToResponse(section)
	if err := logAdminAudit(ctx, AuditParams{
		Repo:        qtx,
		ActorID:     adminID,
		Action:      "section.unpublish",
		EntityType:  "section",
		EntityID:    section.ID,
		BeforeState: nil,
		AfterState:  resp,
	}); err != nil {
		return SectionResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return SectionResponse{}, err
	}
	return resp, nil
}

func (s *Service) ArchiveSection(ctx context.Context, adminID, sectionID uuid.UUID) (SectionResponse, error) {
	if sectionID == uuid.Nil {
		return SectionResponse{}, ErrValidation
	}
	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return SectionResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	section, err := qtx.ArchiveSection(ctx, sectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SectionResponse{}, ErrNotFound
		}
		return SectionResponse{}, err
	}

	err = qtx.ArchiveLevelsBySection(ctx, repository.ArchiveLevelsBySectionParams{
		SectionID: sectionID,
		DeletedBy: uuid.NullUUID{UUID: adminID, Valid: true},
	})
	if err != nil {
		return SectionResponse{}, err
	}

	resp := sectionToResponse(section)
	if err := logAdminAudit(ctx, AuditParams{
		Repo:        qtx,
		ActorID:     adminID,
		Action:      "section.archive",
		EntityType:  "section",
		EntityID:    section.ID,
		BeforeState: nil,
		AfterState:  resp,
	}); err != nil {
		return SectionResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return SectionResponse{}, err
	}
	return resp, nil
}

func (s *Service) CompleteLevel(ctx context.Context, userID, levelID uuid.UUID, req CompleteLevelRequest) (CompleteLevelResponse, error) {
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
	xpAwarded := CalculateXP(level.Difficulty, attemptNumber, req.Completed)

	if err := qtx.InsertLevelAttempt(ctx, repository.InsertLevelAttemptParams{
		ID:            newID(),
		UserID:        userID,
		LevelID:       levelID,
		AttemptDate:   attemptDate,
		AttemptNumber: attemptNumber,
		XpAwarded:     xpAwarded,
		Completed:     req.Completed,
	}); err != nil {
		return CompleteLevelResponse{}, err
	}

	if err := qtx.UpsertPlayerProgressForAttempt(ctx, repository.UpsertPlayerProgressForAttemptParams{
		UserID:          userID,
		LevelID:         levelID,
		BestScore:       req.Score,
		XpTotalForLevel: xpAwarded,
		Completed:       req.Completed,
	}); err != nil {
		return CompleteLevelResponse{}, err
	}

	if req.Completed {
		if err := qtx.UpsertDailyStreak(ctx, repository.UpsertDailyStreakParams{
			UserID:       userID,
			ActivityDate: attemptDate,
		}); err != nil {
			return CompleteLevelResponse{}, err
		}
	}

	eventType := "level_failed"
	if req.Completed {
		eventType = "level_completed"
	}
	if err := qtx.InsertExperienceHistory(ctx, repository.InsertExperienceHistoryParams{
		ID:                 newID(),
		UserID:             uuid.NullUUID{UUID: userID, Valid: true},
		LevelID:            levelID,
		EventType:          eventType,
		XpGained:           xpAwarded,
		Source:             "online",
		VerificationMethod: "online_direct",
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
		Completed:     req.Completed,
		AttemptNumber: attemptNumber,
		XPAwarded:     xpAwarded,
		TotalXP:       totals.TotalXP,
		CurrentStreak: calculateCurrentStreak(streakDates),
		BadgesAwarded: badgesToResponse(badgesAwarded),
	}, nil
}

func (s *Service) GetProfileProgress(ctx context.Context, userID uuid.UUID) (ProfileProgressResponse, error) {
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

func CalculateXP(difficulty int32, attemptNumber int32, completed bool) int32 {
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

func validateLevelInput(title, color, templateType string, difficulty int32, content json.RawMessage) error {
	if strings.TrimSpace(title) == "" || strings.TrimSpace(color) == "" || templateType == "" {
		log.Printf("validateLevelInput failed: basic empty fields")
		return ErrValidation
	}
	if difficulty < 1 || difficulty > 10 {
		log.Printf("validateLevelInput failed: difficulty out of bounds")
		return ErrValidation
	}
	if _, ok := AllowedTemplateTypes[templateType]; !ok {
		log.Printf("validateLevelInput failed: unknown template type %s", templateType)
		return ErrValidation
	}
	if len(content) == 0 || len(content) > 5*1024*1024 || !json.Valid(content) {
		log.Printf("validateLevelInput failed: content size or invalid JSON")
		return ErrValidation
	}
	if string(content) == "[]" || string(content) == "{}" || string(content) == "null" {
		log.Printf("validateLevelInput failed: empty content")
		return ErrValidation
	}
	switch templateType {
	case "trivia":
		return validateTriviaContent(content)
	case "memory":
		return validateMemoryContent(content)
	case "fake_news":
		return validateFakeNewsContent(content)
	case "word_search":
		return validateWordSearchContent(content)
	case "puzzle":
		return validatePuzzleContent(content)
	case "crossword":
		return validateCrosswordContent(content)
	case "snakes_ladders":
		return validateSnakesContent(content)
	default:
		return ErrValidation
	}
}

func validateTriviaContent(content json.RawMessage) error {
	type question struct {
		Question     string   `json:"question"`
		Options      []string `json:"options"`
		CorrectIndex int      `json:"correct_index"`
		MediaURL     string   `json:"media_url,omitempty"`
	}
	type questionEnvelope struct {
		Questions []question `json:"questions"`
	}

	var questions []question
	if err := decodeStrictContent(content, &questions); err != nil {
		var envelope questionEnvelope
		if envelopeErr := decodeStrictContent(content, &envelope); envelopeErr != nil {
			return ErrValidation
		}
		questions = envelope.Questions
	}
	if len(questions) == 0 {
		return ErrValidation
	}
	for _, q := range questions {
		if strings.TrimSpace(q.Question) == "" || len(q.Options) < 2 || len(q.Options) > 4 {
			return ErrValidation
		}
		if q.CorrectIndex < 0 || q.CorrectIndex >= len(q.Options) {
			return ErrValidation
		}
		for _, option := range q.Options {
			if strings.TrimSpace(option) == "" {
				return ErrValidation
			}
		}
	}
	return nil
}

func validateMemoryContent(content json.RawMessage) error {
	type pair struct {
		ID       string `json:"id"`
		Content1 string `json:"content1"`
		Content2 string `json:"content2"`
		Color    string `json:"color,omitempty"`
	}
	var payload struct {
		BackColor string `json:"back_color,omitempty"`
		Pairs     []pair `json:"pairs"`
	}
	if err := decodeStrictContent(content, &payload); err != nil || len(payload.Pairs) < 4 {
		return ErrValidation
	}
	for _, pair := range payload.Pairs {
		if strings.TrimSpace(pair.ID) == "" || strings.TrimSpace(pair.Content1) == "" || strings.TrimSpace(pair.Content2) == "" {
			return ErrValidation
		}
	}
	return nil
}

func validateFakeNewsContent(content json.RawMessage) error {
	type item struct {
		Title       string `json:"title"`
		Content     string `json:"content"`
		IsFake      *bool  `json:"isFake"`
		Explanation string `json:"explanation,omitempty"`
		ImageURL    string `json:"imageUrl,omitempty"`
		Reference   string `json:"reference"`
	}
	var payload struct {
		News []item `json:"news"`
	}
	if err := decodeStrictContent(content, &payload); err != nil || len(payload.News) == 0 {
		return ErrValidation
	}
	for _, item := range payload.News {
		if strings.TrimSpace(item.Title) == "" || strings.TrimSpace(item.Content) == "" || item.IsFake == nil || strings.TrimSpace(item.Reference) == "" {
			return ErrValidation
		}
		if strings.TrimSpace(item.ImageURL) != "" && !isHTTPURL(item.ImageURL) {
			return ErrValidation
		}
	}
	return nil
}

func validateWordSearchContent(content json.RawMessage) error {
	var payload struct {
		Words  []string `json:"words"`
		Width  *int32   `json:"width,omitempty"`
		Height *int32   `json:"height,omitempty"`
		Seed   *int32   `json:"seed,omitempty"`
	}
	if err := decodeStrictContent(content, &payload); err != nil || len(payload.Words) < 2 {
		return ErrValidation
	}
	for _, word := range payload.Words {
		if len([]rune(strings.TrimSpace(word))) < 2 {
			return ErrValidation
		}
	}
	if payload.Width != nil && (*payload.Width < 5 || *payload.Width > 24) {
		return ErrValidation
	}
	if payload.Height != nil && (*payload.Height < 5 || *payload.Height > 24) {
		return ErrValidation
	}
	return nil
}

func validatePuzzleContent(content json.RawMessage) error {
	var payload struct {
		Phrase string `json:"phrase"`
		Pieces *int32 `json:"pieces,omitempty"`
		Seed   *int32 `json:"seed,omitempty"`
	}
	if err := decodeStrictContent(content, &payload); err != nil {
		// Never log the raw content: it can be up to 5 MB and is attacker-influenced (B4).
		log.Printf("validatePuzzleContent unmarshal error: %v", err)
		return ErrValidation
	}
	if strings.TrimSpace(payload.Phrase) == "" {
		log.Printf("validatePuzzleContent empty phrase")
		return ErrValidation
	}
	if payload.Pieces != nil && (*payload.Pieces < 3 || *payload.Pieces > 20) {
		log.Printf("validatePuzzleContent pieces out of bounds: %v", *payload.Pieces)
		return ErrValidation
	}
	return nil
}

// maxCrosswordWords bounds the crossword builder's cubic-ish placement search.
const maxCrosswordWords = 30

func validateCrosswordContent(content json.RawMessage) error {
	type word struct {
		Word string `json:"word"`
		Clue string `json:"clue"`
	}
	var payload struct {
		Words []word `json:"words"`
	}
	// Cap the word count: canBuildConnectedCrossword is ~O(N^3·L^2), so a payload
	// of hundreds of short words (well under the 5 MB body cap) would peg the
	// single vCPU synchronously inside the handler (audit finding B6).
	if err := decodeStrictContent(content, &payload); err != nil || len(payload.Words) < 2 || len(payload.Words) > maxCrosswordWords {
		return ErrValidation
	}
	words := make([]crosswordCandidate, 0, len(payload.Words))
	seen := make(map[string]struct{}, len(payload.Words))
	for _, item := range payload.Words {
		answer := normalizeCrosswordAnswer(item.Word)
		if len([]rune(answer)) < 2 || strings.TrimSpace(item.Clue) == "" {
			return ErrValidation
		}
		if _, exists := seen[answer]; exists {
			return ErrValidation
		}
		seen[answer] = struct{}{}
		words = append(words, crosswordCandidate{Answer: answer, Clue: strings.TrimSpace(item.Clue)})
	}
	if !canBuildConnectedCrossword(words) {
		return ErrValidation
	}
	return nil
}

type crosswordCandidate struct {
	Answer string
	Clue   string
}

type crosswordPlaced struct {
	Answer   string
	X        int
	Y        int
	Vertical bool
}

type crosswordPoint struct {
	X int
	Y int
}

type crosswordPlacement struct {
	X        int
	Y        int
	Vertical bool
	Score    int
}

func canBuildConnectedCrossword(words []crosswordCandidate) bool {
	if len(words) < 2 {
		return false
	}
	sortedWords := append([]crosswordCandidate(nil), words...)
	sort.Slice(sortedWords, func(i, j int) bool {
		return len([]rune(sortedWords[i].Answer)) > len([]rune(sortedWords[j].Answer))
	})

	for firstIndex := range sortedWords {
		orderedWords := make([]crosswordCandidate, 0, len(sortedWords))
		orderedWords = append(orderedWords, sortedWords[firstIndex])
		orderedWords = append(orderedWords, sortedWords[:firstIndex]...)
		orderedWords = append(orderedWords, sortedWords[firstIndex+1:]...)
		if buildConnectedCrosswordCount(orderedWords) == len(words) {
			return true
		}
	}
	return false
}

func buildConnectedCrosswordCount(words []crosswordCandidate) int {
	if len(words) == 0 {
		return 0
	}
	grid := make(map[crosswordPoint]rune)
	placed := make([]crosswordPlaced, 0, len(words))

	placeCrosswordWord(words[0], 0, 0, false, grid, &placed)
	for _, word := range words[1:] {
		placement, ok := findCrosswordPlacement(word, grid, placed)
		if !ok {
			continue
		}
		placeCrosswordWord(word, placement.X, placement.Y, placement.Vertical, grid, &placed)
	}
	return len(placed)
}

func placeCrosswordWord(word crosswordCandidate, x, y int, vertical bool, grid map[crosswordPoint]rune, placed *[]crosswordPlaced) {
	*placed = append(*placed, crosswordPlaced{Answer: word.Answer, X: x, Y: y, Vertical: vertical})
	for i, letter := range []rune(word.Answer) {
		cx := x
		cy := y
		if vertical {
			cy += i
		} else {
			cx += i
		}
		grid[crosswordPoint{X: cx, Y: cy}] = letter
	}
}

func findCrosswordPlacement(word crosswordCandidate, grid map[crosswordPoint]rune, placed []crosswordPlaced) (crosswordPlacement, bool) {
	letters := []rune(word.Answer)
	var best crosswordPlacement
	found := false

	for wordIndex, letter := range letters {
		for _, placedWord := range placed {
			for placedIndex, placedLetter := range []rune(placedWord.Answer) {
				if placedLetter != letter {
					continue
				}
				intersectionX := placedWord.X
				intersectionY := placedWord.Y
				if placedWord.Vertical {
					intersectionY += placedIndex
				} else {
					intersectionX += placedIndex
				}
				vertical := !placedWord.Vertical
				x := intersectionX
				y := intersectionY
				if vertical {
					y -= wordIndex
				} else {
					x -= wordIndex
				}
				if !canPlaceCrosswordWord(letters, x, y, vertical, grid) {
					continue
				}
				score := absInt(x) + absInt(y)
				if !found || score < best.Score {
					best = crosswordPlacement{X: x, Y: y, Vertical: vertical, Score: score}
					found = true
				}
			}
		}
	}

	return best, found
}

func canPlaceCrosswordWord(letters []rune, startX, startY int, vertical bool, grid map[crosswordPoint]rune) bool {
	intersections := 0
	for i, letter := range letters {
		x := startX
		y := startY
		if vertical {
			y += i
		} else {
			x += i
		}
		point := crosswordPoint{X: x, Y: y}
		if existing, ok := grid[point]; ok {
			if existing != letter {
				return false
			}
			intersections++
			continue
		}

		adjacentA := crosswordPoint{X: x, Y: y - 1}
		adjacentB := crosswordPoint{X: x, Y: y + 1}
		if vertical {
			adjacentA = crosswordPoint{X: x - 1, Y: y}
			adjacentB = crosswordPoint{X: x + 1, Y: y}
		}
		if _, ok := grid[adjacentA]; ok {
			return false
		}
		if _, ok := grid[adjacentB]; ok {
			return false
		}
	}

	before := crosswordPoint{X: startX - 1, Y: startY}
	after := crosswordPoint{X: startX + len(letters), Y: startY}
	if vertical {
		before = crosswordPoint{X: startX, Y: startY - 1}
		after = crosswordPoint{X: startX, Y: startY + len(letters)}
	}
	if _, ok := grid[before]; ok {
		return false
	}
	if _, ok := grid[after]; ok {
		return false
	}

	return intersections > 0
}

func normalizeCrosswordAnswer(raw string) string {
	var builder strings.Builder
	for _, letter := range strings.ToUpper(strings.TrimSpace(raw)) {
		switch letter {
		case 'Á', 'À', 'Ä', 'Â':
			letter = 'A'
		case 'É', 'È', 'Ë', 'Ê':
			letter = 'E'
		case 'Í', 'Ì', 'Ï', 'Î':
			letter = 'I'
		case 'Ó', 'Ò', 'Ö', 'Ô':
			letter = 'O'
		case 'Ú', 'Ù', 'Ü', 'Û':
			letter = 'U'
		}
		if (letter >= 'A' && letter <= 'Z') || letter == 'Ñ' {
			builder.WriteRune(letter)
		}
	}
	return builder.String()
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func validateSnakesContent(content json.RawMessage) error {
	type item struct {
		Start int32 `json:"start"`
		End   int32 `json:"end"`
	}
	type aiConfig struct {
		Difficulty      string    `json:"difficulty"`
		FailProbability *float64  `json:"fail_probability,omitempty"`
		Weights         []float64 `json:"weights,omitempty"`
	}
	type question struct {
		Question     string   `json:"question"`
		Options      []string `json:"options"`
		CorrectIndex int      `json:"correct_index"`
	}
	var payload struct {
		BoardWidth    int32      `json:"board_width"`
		BoardHeight   int32      `json:"board_height"`
		StartPosition int32      `json:"start_position"`
		EndPosition   int32      `json:"end_position"`
		Seed          *int32     `json:"seed,omitempty"`
		Snakes        []item     `json:"snakes,omitempty"`
		Ladders       []item     `json:"ladders,omitempty"`
		AIConfig      *aiConfig  `json:"ai_config,omitempty"`
		Questions     []question `json:"questions,omitempty"`
	}
	if err := decodeStrictContent(content, &payload); err != nil {
		return ErrValidation
	}
	totalCells := payload.BoardWidth * payload.BoardHeight
	if payload.BoardWidth < 1 || payload.BoardHeight < 1 || payload.StartPosition < 1 || payload.EndPosition < 1 || payload.StartPosition > totalCells || payload.EndPosition > totalCells {
		return ErrValidation
	}

	origins := make(map[int32]struct{}, len(payload.Snakes)+len(payload.Ladders))
	for _, snake := range payload.Snakes {
		if snake.Start <= snake.End || !positionInBoard(snake.Start, totalCells) || !positionInBoard(snake.End, totalCells) {
			return ErrValidation
		}
		if _, exists := origins[snake.Start]; exists {
			return ErrValidation
		}
		origins[snake.Start] = struct{}{}
	}
	for _, ladder := range payload.Ladders {
		if ladder.Start >= ladder.End || !positionInBoard(ladder.Start, totalCells) || !positionInBoard(ladder.End, totalCells) {
			return ErrValidation
		}
		if _, exists := origins[ladder.Start]; exists {
			return ErrValidation
		}
		origins[ladder.Start] = struct{}{}
	}
	if payload.AIConfig != nil {
		if payload.AIConfig.Difficulty != "EASY" && payload.AIConfig.Difficulty != "MEDIUM" && payload.AIConfig.Difficulty != "HARD" {
			return ErrValidation
		}
		if payload.AIConfig.FailProbability != nil && (*payload.AIConfig.FailProbability < 0 || *payload.AIConfig.FailProbability > 1) {
			return ErrValidation
		}
		for _, weight := range payload.AIConfig.Weights {
			if weight < 0 {
				return ErrValidation
			}
		}
	}
	// Snakes & ladders gates every dice roll behind a two-option question drawn
	// from a queue (see frontend SnakeLadderGame); a minimum bank of 8 keeps
	// that queue from cycling through the same handful of questions too fast.
	if len(payload.Questions) < minSnakesQuestions {
		return ErrValidation
	}
	for _, question := range payload.Questions {
		if strings.TrimSpace(question.Question) == "" || len(question.Options) != 2 {
			return ErrValidation
		}
		if question.CorrectIndex < 0 || question.CorrectIndex >= len(question.Options) {
			return ErrValidation
		}
		for _, option := range question.Options {
			if strings.TrimSpace(option) == "" {
				return ErrValidation
			}
		}
	}
	return nil
}

const minSnakesQuestions = 8

func decodeStrictContent(content json.RawMessage, dst any) error {
	return httpjson.DecodeStrict(bytes.NewReader(content), dst)
}

func positionInBoard(position, totalCells int32) bool {
	return position >= 1 && position <= totalCells
}

func isHTTPURL(rawURL string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

type AuditParams struct {
	Repo        *repository.Queries
	ActorID     uuid.UUID
	Action      string
	EntityType  string
	EntityID    uuid.UUID
	BeforeState any
	AfterState  any
}

// logAdminAudit is a thin wrapper over the shared audit package, kept so the
// 10 existing call sites in this file stay unchanged. Content admin actions have
// no per-request IP/user-agent here, so audit.Log fills its placeholders.
func logAdminAudit(ctx context.Context, params AuditParams) error {
	return audit.Log(ctx, params.Repo, audit.Entry{
		ActorID:    params.ActorID,
		Action:     params.Action,
		EntityType: params.EntityType,
		EntityID:   params.EntityID,
		Before:     params.BeforeState,
		After:      params.AfterState,
	})
}

func levelAuditPayload(level LevelResponse) map[string]any {
	return map[string]any{
		"id":            level.ID,
		"section_id":    level.SectionID,
		"title":         level.Title,
		"color":         level.Color,
		"template_type": level.TemplateType,
		"difficulty":    level.Difficulty,
		"is_published":  level.IsPublished,
	}
}

func levelToResponse(level repository.Level) LevelResponse {
	resp := LevelResponse{
		ID:           level.ID,
		SectionID:    level.SectionID,
		Title:        level.Title,
		Color:        level.Color,
		TemplateType: level.TemplateType,
		Content:      level.Content,
		Difficulty:   level.Difficulty,
		IsPublished:  level.IsPublished,
		CreatedAt:    level.CreatedAt,
		UpdatedAt:    level.UpdatedAt,
	}
	if level.CreatedByAdminID.Valid {
		resp.CreatedByAdminID = level.CreatedByAdminID.UUID
	}
	return resp
}

func sectionToResponse(section repository.Section) SectionResponse {
	resp := SectionResponse{
		ID:          section.ID,
		Title:       section.Title,
		Description: section.Description,
		Color:       section.Color,
		IsPublished: section.IsPublished,
		CreatedAt:   section.CreatedAt,
	}
	if section.CreatedByAdminID.Valid {
		resp.CreatedByAdminID = section.CreatedByAdminID.UUID
	}
	return resp
}

func badgesToResponse(rows []repository.BadgeWithEarnedAt) []BadgeResponse {
	badges := make([]BadgeResponse, 0, len(rows))
	for _, row := range rows {
		badges = append(badges, BadgeResponse{
			ID:          row.ID,
			Name:        row.Name,
			XPThreshold: row.XpThreshold,
			IconKey:     row.IconKey,
			EarnedAt:    row.EarnedAt,
		})
	}
	return badges
}

func completionDate(clientFinishedAt string) time.Time {
	now := time.Now().UTC()
	if strings.TrimSpace(clientFinishedAt) != "" {
		if parsed, err := time.Parse(time.RFC3339, clientFinishedAt); err == nil {
			now = parsed.UTC()
		}
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func calculateCurrentStreak(dates []time.Time) int32 {
	if len(dates) == 0 {
		return 0
	}

	seen := make(map[string]struct{}, len(dates))
	for _, date := range dates {
		utc := date.UTC()
		key := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		seen[key] = struct{}{}
	}

	cursor := time.Now().UTC()
	today := time.Date(cursor.Year(), cursor.Month(), cursor.Day(), 0, 0, 0, 0, time.UTC)
	if _, ok := seen[today.Format("2006-01-02")]; !ok {
		yesterday := today.AddDate(0, 0, -1)
		if _, ok := seen[yesterday.Format("2006-01-02")]; !ok {
			return 0
		}
		today = yesterday
	}

	var streak int32
	for {
		if _, ok := seen[today.Format("2006-01-02")]; !ok {
			break
		}
		streak++
		today = today.AddDate(0, 0, -1)
	}
	return streak
}

func newID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}
