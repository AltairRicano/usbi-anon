// Migrado desde ../usbi/backend/internal/repository/maintenance_queries.go.
// Único cambio: las tres consultas apuntan a `identities` en vez de `users`.
package identityrepo

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

func (q *Queries) ListPendingTutorConsentUsers(ctx context.Context, cutoff time.Time, limit int32) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT id
FROM identities
WHERE status = 'pending_tutor_consent'
  AND deleted_at IS NULL
  AND created_at < $1
ORDER BY created_at ASC
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

func (q *Queries) SuspendInactivePlayers(ctx context.Context, cutoff time.Time) (int64, error) {
	result, err := q.db.ExecContext(ctx, `
UPDATE identities
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
	return rowsAffected(result)
}

func (q *Queries) ListSuspendedUsersForCancellation(ctx context.Context, cutoff time.Time, limit int32) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT id
FROM identities
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

func rowsAffected(result sql.Result) (int64, error) {
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}
