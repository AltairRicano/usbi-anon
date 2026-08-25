// arco_queries.go es NUEVO en F9. F7 deliberadamente no lo trajo de vuelta
// desde identityrepo (ver privacy_queries.go): estas consultas solo las
// necesita internal/auth (Arco/ListPendingArco/ResolveArco), que es F9.
//
// arco_requests se simplificó frente al diseño de dos bases: sin saga de
// tres pasos (pending → purging_main → identity_pseudonymized → resolved),
// porque la cancelación ya no pasa por aquí — es autoservicio inmediato vía
// DELETE /auth/me (internal/privacy.CancelAccount). Estas cuatro consultas
// solo sirven a acceso/rectificacion/oposicion, que sí siguen necesitando
// intervención de un admin.
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type InsertArcoRequestParams struct {
	ID            uuid.UUID
	UserID        uuid.NullUUID
	RequesterType string
	RequestType   string
	EvidenceHash  []byte
}

// InsertArcoRequest nace siempre en 'pending' — el esquema ya no tiene
// estados intermedios de saga.
func (q *Queries) InsertArcoRequest(ctx context.Context, arg InsertArcoRequestParams) error {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO arco_requests (id, user_id, requester_type, request_type, status, evidence_hash)
VALUES ($1, $2, $3, $4, 'pending', $5)
`, arg.ID, arg.UserID, arg.RequesterType, arg.RequestType, arg.EvidenceHash)
	return err
}

type ArcoPendingRequest struct {
	ID            uuid.UUID
	RequesterType string
	RequestType   string
	Status        string
	ReceivedAt    time.Time
}

func (q *Queries) ListPendingArcoRequests(ctx context.Context, limit int32) ([]ArcoPendingRequest, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT id, requester_type, request_type, status, received_at
FROM arco_requests
WHERE status = 'pending'
ORDER BY received_at ASC
LIMIT $1
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ArcoPendingRequest, 0)
	for rows.Next() {
		var item ArcoPendingRequest
		if err := rows.Scan(&item.ID, &item.RequesterType, &item.RequestType, &item.Status, &item.ReceivedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type ArcoRequestForResolution struct {
	ID          uuid.UUID
	UserID      uuid.NullUUID
	RequestType string
	Status      string
}

// GetArcoRequestForUpdate bloquea la fila bajo FOR UPDATE — el llamador debe
// envolverla en una transacción, igual que el guard de mínimo 4 preguntas.
// Sin saga que reanudar, el único motivo del lock es evitar que dos admins
// resuelvan el mismo trámite a la vez.
func (q *Queries) GetArcoRequestForUpdate(ctx context.Context, id uuid.UUID) (ArcoRequestForResolution, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT id, user_id, request_type, status
FROM arco_requests
WHERE id = $1
FOR UPDATE
`, id)
	var r ArcoRequestForResolution
	err := row.Scan(&r.ID, &r.UserID, &r.RequestType, &r.Status)
	return r, err
}

type ResolveArcoRequestParams struct {
	ID              uuid.UUID
	HandledBy       uuid.NullUUID
	Status          string
	ResponseSummary string
}

func (q *Queries) ResolveArcoRequest(ctx context.Context, arg ResolveArcoRequestParams) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE arco_requests
SET status = $2, resolved_at = NOW(), handled_by = $3, response_summary = $4
WHERE id = $1
`, arg.ID, arg.Status, arg.HandledBy, arg.ResponseSummary)
	return err
}
