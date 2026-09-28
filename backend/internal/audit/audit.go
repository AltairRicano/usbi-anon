// Package audit centraliza las escrituras en audit_log para que toda operación
// sensible registre evidencia de No-Repudio a través de una ruta única y consistente.
package audit

import (
	"context"
	"encoding/json"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

// Entry es un único registro de auditoría. Before/After se serializan a JSON; nil
// se convierte en un objeto vacío. IP/UserAgent tienen valores por defecto para
// acciones internas del backend que no tienen contexto de petición HTTP.
type Entry struct {
	ActorID    uuid.UUID
	Action     string
	EntityType string
	EntityID   uuid.UUID
	Before     any
	After      any
	IP         string
	UserAgent  string
}

// Log añade una entrada a audit_log usando el repositorio dado (posiblemente transaccional).
// El trigger de solo-añadir en la tabla garantiza que la fila
// nunca pueda ser actualizada o eliminada después.
func Log(ctx context.Context, repo *repository.Queries, e Entry) error {
	before, err := marshalState(e.Before)
	if err != nil {
		return err
	}
	after, err := marshalState(e.After)
	if err != nil {
		return err
	}
	ip := e.IP
	if ip == "" {
		ip = "0.0.0.0"
	}
	userAgent := e.UserAgent
	if userAgent == "" {
		userAgent = "backend-service"
	}
	return repo.LogAuditEntry(ctx, repository.LogAuditEntryParams{
		ID:             uuid.New(),
		ActorAccountID: uuid.NullUUID{UUID: e.ActorID, Valid: e.ActorID != uuid.Nil},
		Action:         e.Action,
		EntityType:     e.EntityType,
		EntityID:       uuid.NullUUID{UUID: e.EntityID, Valid: e.EntityID != uuid.Nil},
		BeforeState:    before,
		AfterState:     after,
		IpAddress:      ip,
		UserAgent:      userAgent,
	})
}

func marshalState(value any) (pqtype.NullRawMessage, error) {
	if value == nil {
		return pqtype.NullRawMessage{RawMessage: json.RawMessage(`{}`), Valid: true}, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return pqtype.NullRawMessage{}, err
	}
	return pqtype.NullRawMessage{RawMessage: data, Valid: true}, nil
}
