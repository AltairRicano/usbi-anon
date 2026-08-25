// LogIdentityAudit es la contraparte de repository.LogAdminAudit, pero
// escribe en identity_audit_log en vez de admin_audit_log. Existe porque en
// ../usbi había una sola bitácora para cualquier acción administrativa,
// incluidas las de identidad (auth.AgeUp, auth.ResolveArcoRequest,
// cmd/create_admin); aquí esas tres acciones seudonimizarían — o peor,
// filtrarían — antes/after states de identidad si escribieran en la base
// principal (ver plan/01_Base_de_datos.md §2.3). internal/audit no se tocó
// (debe seguir con diff cero contra ../usbi — plan/02_Backend.md §8 criterio
// 4) porque su función Log() está atada a *repository.Queries; los tres
// call-sites de identidad llaman a LogIdentityAudit directamente en vez de
// pasar por ese paquete compartido.
package identityrepo

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

// IdentityAuditEntry espeja audit.Entry (internal/audit) para no obligar a
// los llamadores de identidad a importar ese paquete solo por el tipo Entry.
type IdentityAuditEntry struct {
	ActorID    uuid.UUID
	Action     string
	EntityType string
	EntityID   uuid.UUID
	Before     any
	After      any
	IP         string
	UserAgent  string
}

// LogIdentityAudit inserta en identity_audit_log. El trigger append-only de
// la base de identidad (migrations/identity/0001) garantiza que la fila
// nunca se actualiza ni se borra después.
func (q *Queries) LogIdentityAudit(ctx context.Context, e IdentityAuditEntry) error {
	before, err := marshalAuditState(e.Before)
	if err != nil {
		return err
	}
	after, err := marshalAuditState(e.After)
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
	_, err = q.db.ExecContext(ctx, `
INSERT INTO identity_audit_log (
    id, actor_user_id, action, entity_type, entity_id, before_state, after_state, ip_address, user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
`,
		uuid.New(),
		uuid.NullUUID{UUID: e.ActorID, Valid: e.ActorID != uuid.Nil},
		e.Action,
		e.EntityType,
		uuid.NullUUID{UUID: e.EntityID, Valid: e.EntityID != uuid.Nil},
		before,
		after,
		ip,
		userAgent,
	)
	return err
}

func marshalAuditState(value any) (pqtype.NullRawMessage, error) {
	if value == nil {
		return pqtype.NullRawMessage{RawMessage: json.RawMessage(`{}`), Valid: true}, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return pqtype.NullRawMessage{}, err
	}
	return pqtype.NullRawMessage{RawMessage: data, Valid: true}, nil
}
