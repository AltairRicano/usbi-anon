package suggestions

import (
	"time"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

// SubmitRequest es el cuerpo de POST /suggestions. No incluye ningún campo de
// identidad ya que las sugerencias son anónimas por diseño.
type SubmitRequest struct {
	Description string `json:"description"`
}

// SuggestionResponse es la representación de una sugerencia para el panel
// admin. No existe lectura para el jugador, garantizando el anonimato de quien envía.
type SuggestionResponse struct {
	ID                      uuid.UUID `json:"id"`
	Description             string    `json:"description"`
	LevelsCompletedSnapshot int32     `json:"levels_completed_snapshot"`
	XPSnapshot              int32     `json:"xp_snapshot"`
	SubmittedAt             time.Time `json:"submitted_at"`
}

// SubmitResponse es la confirmación mínima que recibe el jugador tras enviar su sugerencia,
// sin eco del contenido ni del snapshot de progreso.
type SubmitResponse struct {
	ID          uuid.UUID `json:"id"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// Page es la respuesta paginada de GET /admin/suggestions.
type Page struct {
	Items      []SuggestionResponse `json:"items"`
	NextCursor string               `json:"next_cursor,omitempty"`
}

func toResponse(s repository.Suggestion) SuggestionResponse {
	return SuggestionResponse{
		ID:                      s.ID,
		Description:             s.Description,
		LevelsCompletedSnapshot: s.LevelsCompletedSnapshot,
		XPSnapshot:              s.XPSnapshot,
		SubmittedAt:             s.SubmittedAt,
	}
}
