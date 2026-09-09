package suggestions

import (
	"time"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

// SubmitRequest es el cuerpo de POST /suggestions. Deliberadamente no lleva
// ningún campo de identidad — suggestions es anónima por diseño
// (migración 0003_enlaces_interes_y_sugerencias). (Relleno)
type SubmitRequest struct {
	Description string `json:"description"`
}

// SuggestionResponse es la representación de una sugerencia para el panel
// admin. No existe una respuesta equivalente para el jugador que la mandó:
// usbi_app no tiene SELECT en esta tabla (00_roles_unificado.sql) — ni
// siquiera el propio autor puede releerla. (Relleno)
type SuggestionResponse struct {
	ID                      uuid.UUID `json:"id"`
	Description             string    `json:"description"`
	LevelsCompletedSnapshot int32     `json:"levels_completed_snapshot"`
	XPSnapshot              int32     `json:"xp_snapshot"`
	SubmittedAt             time.Time `json:"submitted_at"`
}

// SubmitResponse es lo único que ve el jugador tras enviar su sugerencia:
// una confirmación mínima, sin eco del contenido ni del snapshot — ese dato
// es para el panel admin, no para que el cliente lo reconstruya. (Relleno)
type SubmitResponse struct {
	ID          uuid.UUID `json:"id"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// Page es la respuesta paginada de GET /admin/suggestions — mismo shape que
// levels.LevelsPage. (Relleno)
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
