// AdminService agrupa las operaciones de internal/levels que corren con el
// pool de usbi_moderador (F3, 2026-09-09): todo lo que antes vivía detrás de
// canManageContent/canArchiveContent en el Handler. A diferencia de
// PlayerService (player_service.go), GetLevel siempre puede ver contenido no
// publicado y ListLevels/ListSections lo exponen bajo un parámetro explícito
// — coherente con el comportamiento que ya tenía el Service único antes de
// partirse en dos. (Útil)
package levels

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

type AdminService struct {
	repo *repository.Queries
}

func NewAdminService(repo *repository.Queries) *AdminService {
	return &AdminService{repo: repo}
}

func (s *AdminService) CreateLevel(ctx context.Context, adminID uuid.UUID, req CreateLevelRequest) (LevelResponse, error) {
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

func (s *AdminService) UpdateLevel(ctx context.Context, adminID, levelID uuid.UUID, req UpdateLevelRequest) (LevelResponse, error) {
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

func (s *AdminService) PublishLevel(ctx context.Context, adminID, levelID uuid.UUID) (LevelResponse, error) {
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

func (s *AdminService) UnpublishLevel(ctx context.Context, adminID, levelID uuid.UUID) (LevelResponse, error) {
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

func (s *AdminService) ArchiveLevel(ctx context.Context, adminID, levelID uuid.UUID) (LevelResponse, error) {
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

func (s *AdminService) UnarchiveLevel(ctx context.Context, adminID, levelID uuid.UUID) (LevelResponse, error) {
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

func (s *AdminService) ListArchivedLevels(ctx context.Context, sectionID uuid.UUID) (ArchivedLevelsResponse, error) {
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
// temporada — ver CLAUDE.md, "rotación de niveles por temporadas". (Útil)
func (s *AdminService) PurgeLevel(ctx context.Context, adminID, levelID uuid.UUID) error {
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

func (s *AdminService) UnarchiveSection(ctx context.Context, adminID, sectionID uuid.UUID) (SectionResponse, error) {
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

func (s *AdminService) ListArchivedSections(ctx context.Context) (ArchivedSectionsResponse, error) {
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
// niveles uno por uno. (Útil)
func (s *AdminService) PurgeSection(ctx context.Context, adminID, sectionID uuid.UUID) error {
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

// GetLevel siempre puede ver contenido no publicado — el Handler solo llama a
// este método cuando canManageContent(claims.Role) ya dio true, así que no
// hace falta un parámetro includeUnpublished aquí (a diferencia de
// ListLevels/ListSections, donde el admin sigue optando explícitamente vía
// query param). (Útil)
func (s *AdminService) GetLevel(ctx context.Context, levelID uuid.UUID) (LevelResponse, error) {
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
	return levelToResponse(level), nil
}

func (s *AdminService) ListLevels(ctx context.Context, cursor uuid.UUID, sectionID uuid.UUID, includeUnpublished bool, pageSize int32) (LevelsPage, error) {
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

func (s *AdminService) CreateSection(ctx context.Context, adminID uuid.UUID, req CreateSectionRequest) (SectionResponse, error) {
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

func (s *AdminService) ListSections(ctx context.Context, includeUnpublished bool) (SectionsResponse, error) {
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

func (s *AdminService) UpdateSection(ctx context.Context, adminID, sectionID uuid.UUID, req UpdateSectionRequest) (SectionResponse, error) {
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

func (s *AdminService) PublishSection(ctx context.Context, adminID, sectionID uuid.UUID) (SectionResponse, error) {
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

func (s *AdminService) UnpublishSection(ctx context.Context, adminID, sectionID uuid.UUID) (SectionResponse, error) {
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

func (s *AdminService) ArchiveSection(ctx context.Context, adminID, sectionID uuid.UUID) (SectionResponse, error) {
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
