// maintenance_queries.go absorbe de identityrepo/maintenance_queries.go
// (F2, dos bases) las dos consultas de retención legal automática que siguen
// aplicando tras el rediseño de identidad: suspender jugadores inactivos y,
// más tarde, cancelarlos. ListPendingTutorConsentUsers no vuelve — el flujo
// de tutor por correo se eliminó por completo (§1 del rediseño), así que ya
// no existe el estado 'pending_tutor_consent' que esa consulta filtraba.
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SuspendInactivePlayers implementa la retención legal automática: un jugador
// sin actividad desde `cutoff` pasa a suspended y pierde sus sesiones vivas
// (token_version+1), quedando marcado para cancelación definitiva tras el
// segundo plazo (ver ListSuspendedUsersForCancellation).
func (q *Queries) SuspendInactivePlayers(ctx context.Context, cutoff time.Time) (int64, error) {
	result, err := q.db.ExecContext(ctx, `
UPDATE accounts
SET status = 'suspended',
    token_version = token_version + 1,
    deletion_reason = 'inactive_suspension_pending_cancel',
    updated_at = NOW()
WHERE role = 'player'
  AND status = 'active'
  AND deleted_at IS NULL
  AND COALESCE(last_login_at, created_at) < $1
`, cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (q *Queries) ListSuspendedUsersForCancellation(ctx context.Context, cutoff time.Time, limit int32) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT id
FROM accounts
WHERE role = 'player'
  AND status = 'suspended'
  AND deletion_reason = 'inactive_suspension_pending_cancel'
  AND deleted_at IS NULL
  AND updated_at < $1
ORDER BY updated_at ASC
LIMIT $2
`, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
