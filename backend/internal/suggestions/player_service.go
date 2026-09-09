// Package suggestions implementa el buzón de sugerencias anónimo (migración
// 0003_enlaces_interes_y_sugerencias, F4 — estado_proyecto.md 2026-09-09).
package suggestions

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

var ErrValidation = errors.New("validation error")

const descriptionMaxLen = 1000

// PlayerService corre sobre el pool de jugador (usbi_app), que en
// 00_roles_unificado.sql solo tiene INSERT en suggestions — ni siquiera
// SELECT, así que este Service no expone ningún método de lectura.
type PlayerService struct {
	repo *repository.Queries
}

func NewPlayerService(repo *repository.Queries) *PlayerService {
	return &PlayerService{repo: repo}
}

func newID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}

// Submit calcula el snapshot de progreso (misma fórmula que
// GetProfileProgress, GetUserProgressTotals) en el momento del envío y lo
// congela en la fila — la sugerencia queda desvinculada de userID desde
// este punto, sin ninguna FK ni columna que lo conserve (comentario de la
// migración 0003: "es anónimo por diseño"). No se llama a audit.Log aquí a
// propósito: una entrada de auditoría con actor_account_id + marca de
// tiempo casi idéntica a submitted_at recrearía el vínculo que la ausencia
// de account_id en la tabla busca evitar.
func (s *PlayerService) Submit(ctx context.Context, userID uuid.UUID, req SubmitRequest) (SubmitResponse, error) {
	description := strings.TrimSpace(req.Description)
	if description == "" || len(description) > descriptionMaxLen {
		return SubmitResponse{}, ErrValidation
	}

	totals, err := s.repo.GetUserProgressTotals(ctx, userID)
	if err != nil {
		return SubmitResponse{}, err
	}

	suggestion, err := s.repo.CreateSuggestion(ctx, repository.CreateSuggestionParams{
		ID:                      newID(),
		Description:             description,
		LevelsCompletedSnapshot: totals.CompletedLevels,
		XPSnapshot:              totals.TotalXP,
		SubmittedAt:             time.Now().UTC(),
	})
	if err != nil {
		return SubmitResponse{}, err
	}
	return SubmitResponse{ID: suggestion.ID, SubmittedAt: suggestion.SubmittedAt}, nil
}
