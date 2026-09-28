// Package suggestions implementa el buzón de sugerencias anónimo.
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

// PlayerService corre sobre el pool de jugador (usbi_app), que
// solo tiene permisos de inserción en suggestions para preservar el anonimato.
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

// Submit calcula el snapshot de progreso en el momento del envío y lo
// congela en la fila — la sugerencia queda desvinculada de userID desde
// este punto, sin ninguna clave foránea ni columna que lo conserve. No se registra en audit.Log
// para evitar que una marca temporal y actor_account_id correlacionen la sugerencia con el autor.
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
