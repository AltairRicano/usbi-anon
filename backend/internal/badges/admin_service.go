package badges

import (
	"context"
	"database/sql"
	"errors"

	"github.com/altair/usbi-anon-backend/internal/audit"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// AdminService corre sobre el pool de moderador (usbi_moderador), que tiene
// CRUD completo en badges desde el 2026-09-02 (00_roles_unificado.sql) —
// B3 solo construye la capa Go que faltaba, sin tocar la matriz de roles.
type AdminService struct {
	repo *repository.Queries
}

func NewAdminService(repo *repository.Queries) *AdminService {
	return &AdminService{repo: repo}
}

func newID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}

func (s *AdminService) List(ctx context.Context) ([]BadgeResponse, error) {
	items, err := s.repo.ListBadges(ctx)
	if err != nil {
		return nil, err
	}
	resp := make([]BadgeResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toResponse(item))
	}
	return resp, nil
}

func (s *AdminService) Create(ctx context.Context, adminID uuid.UUID, req CreateBadgeRequest) (BadgeResponse, error) {
	name, iconKey := trimBadgeInput(req.Name, req.IconKey)
	if err := validateBadgeInput(name, iconKey, req.XPThreshold); err != nil {
		return BadgeResponse{}, err
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return BadgeResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	badge, err := qtx.CreateBadge(ctx, repository.CreateBadgeParams{
		ID:          newID(),
		Name:        name,
		XpThreshold: req.XPThreshold,
		IconKey:     iconKey,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return BadgeResponse{}, ErrValidation
		}
		return BadgeResponse{}, err
	}
	resp := toResponse(badge)
	if err := audit.Log(ctx, qtx, audit.Entry{
		ActorID:    adminID,
		Action:     "badge.create",
		EntityType: "badge",
		EntityID:   badge.ID,
		After:      badgeAuditPayload(resp),
	}); err != nil {
		return BadgeResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return BadgeResponse{}, err
	}
	return resp, nil
}

func (s *AdminService) Update(ctx context.Context, adminID, badgeID uuid.UUID, req UpdateBadgeRequest) (BadgeResponse, error) {
	name, iconKey := trimBadgeInput(req.Name, req.IconKey)
	if err := validateBadgeInput(name, iconKey, req.XPThreshold); err != nil {
		return BadgeResponse{}, err
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return BadgeResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	badge, err := qtx.UpdateBadge(ctx, repository.UpdateBadgeParams{
		ID:          badgeID,
		Name:        name,
		XpThreshold: req.XPThreshold,
		IconKey:     iconKey,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return BadgeResponse{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return BadgeResponse{}, ErrValidation
		}
		return BadgeResponse{}, err
	}
	resp := toResponse(badge)
	if err := audit.Log(ctx, qtx, audit.Entry{
		ActorID:    adminID,
		Action:     "badge.update",
		EntityType: "badge",
		EntityID:   badge.ID,
		After:      badgeAuditPayload(resp),
	}); err != nil {
		return BadgeResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return BadgeResponse{}, err
	}
	return resp, nil
}

// Delete pregunta primero por titulares vivos (mismo patrón que
// interestlinks.DeleteCategory/levels.PurgeSection) en vez de dejar que el
// ON DELETE RESTRICT de user_badges.badge_id devuelva un error crudo de
// Postgres. Una insignia ya ganada nunca se revoca: si al menos un jugador
// la tiene, el borrado se rechaza con un error de dominio propio, nunca con
// un 500.
func (s *AdminService) Delete(ctx context.Context, adminID, badgeID uuid.UUID) error {
	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	count, err := qtx.CountUserBadgesByBadge(ctx, badgeID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrBadgeHasHolder
	}

	rows, err := qtx.DeleteBadge(ctx, badgeID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	if err := audit.Log(ctx, qtx, audit.Entry{
		ActorID:    adminID,
		Action:     "badge.delete",
		EntityType: "badge",
		EntityID:   badgeID,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

// isUniqueViolation traduce el 23505 de Postgres (badges.name ahora único,
// migración 0006) en vez de dejar pasar el error crudo del driver — mismo
// patrón que interestlinks.isUniqueViolation.
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
