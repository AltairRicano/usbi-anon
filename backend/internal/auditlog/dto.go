package auditlog

import (
	"time"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

// EntryResponse es una fila de audit_log para el panel de administración.
// actor_account_id viaja como UUID crudo: usbi_moderador no tiene ningún
// GRANT sobre accounts, así que este paquete no puede resolverlo a un
// nickname sin cruzar al pool de jugador — justo lo que la separación de
// pools evita (estado_proyecto.md 2026-09-09, sección B1). (Relleno)
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
// opaco: "<RFC3339Nano de created_at>_<id>" — no reveses en cursores por
// simple id como suggestions, porque audit_log usa UUIDv4 (ver
// repository.ListAuditLogParams). (Relleno)
type Page struct {
	Items      []EntryResponse `json:"items"`
	NextCursor string          `json:"next_cursor,omitempty"`
}

// Filters es el conjunto de filtros opcionales de GET /admin/audit-log,
// todos combinables. (Relleno)
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
