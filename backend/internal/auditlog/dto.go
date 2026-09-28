package auditlog

import (
	"time"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

// EntryResponse es una fila de audit_log para el panel de administración.
// actor_account_id viaja como UUID crudo: usbi_moderador no tiene ningún
// GRANT sobre accounts, evitando que este paquete resuelva nicknames cruzando al pool de jugador.
type EntryResponse struct {
	ID             uuid.UUID  `json:"id"`
	ActorAccountID *uuid.UUID `json:"actor_account_id,omitempty"`
	Action         string     `json:"action"`
	EntityType     string     `json:"entity_type"`
	EntityID       *uuid.UUID `json:"entity_id,omitempty"`
	BeforeState    any        `json:"before_state,omitempty"`
	AfterState     any        `json:"after_state,omitempty"`
	IPAddress      string     `json:"ip_address"`
	UserAgent      string     `json:"user_agent"`
	CreatedAt      time.Time  `json:"created_at"`
}

// Page es la respuesta paginada de GET /admin/audit-log. NextCursor es
// opaco: "<RFC3339Nano de created_at>_<id>", ya que audit_log usa UUIDv4.
type Page struct {
	Items      []EntryResponse `json:"items"`
	NextCursor string          `json:"next_cursor,omitempty"`
}

// Filters es el conjunto de filtros opcionales de GET /admin/audit-log,
// todos combinables.
type Filters struct {
	ActorAccountID uuid.UUID
	Action         string
	EntityType     string
	From           time.Time
	To             time.Time
	PageSize       int32
	Cursor         string
}

func toResponse(e repository.AuditLogEntry) EntryResponse {
	resp := EntryResponse{
		ID:         e.ID,
		Action:     e.Action,
		EntityType: e.EntityType,
		IPAddress:  e.IPAddress,
		UserAgent:  e.UserAgent,
		CreatedAt:  e.CreatedAt,
	}
	if e.ActorAccountID.Valid {
		id := e.ActorAccountID.UUID
		resp.ActorAccountID = &id
	}
	if e.EntityID.Valid {
		id := e.EntityID.UUID
		resp.EntityID = &id
	}
	if e.BeforeState.Valid {
		resp.BeforeState = e.BeforeState.RawMessage
	}
	if e.AfterState.Valid {
		resp.AfterState = e.AfterState.RawMessage
	}
	return resp
}
