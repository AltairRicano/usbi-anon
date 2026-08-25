// auth_queries.go absorbe las consultas de refresh_tokens que antes vivían en
// identityrepo/auth_queries.go (F2, dos bases). Migradas sin más cambio que
// el JOIN contra `accounts` en vez de `identities` — la tabla ya traía la FK
// renombrada a account_id desde F5 (ver
// plan/04_Rediseno_identidad_gustos.md §1). IsNoRows no se duplica aquí: ya
// vive en errors.go, en este mismo paquete.
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type InsertRefreshTokenParams struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	TokenHash []byte
	ExpiresAt time.Time
}

func (q *Queries) InsertRefreshToken(ctx context.Context, arg InsertRefreshTokenParams) error {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO refresh_tokens (id, account_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4)
`, arg.ID, arg.AccountID, arg.TokenHash, arg.ExpiresAt)
	return err
}

// RefreshTokenAccount es la proyección mínima que necesita el flujo de
// refresh: no trae nickname ni ningún otro dato que no haga falta para
// revalidar la sesión y volver a emitir el JWT.
type RefreshTokenAccount struct {
	TokenID      uuid.UUID
	AccountID    uuid.UUID
	IsAdult      bool
	Role         string
	Status       string
	TokenVersion int32
	CreatedAt    time.Time
}

func (q *Queries) GetRefreshTokenAccount(ctx context.Context, tokenHash []byte) (RefreshTokenAccount, error) {
	var row RefreshTokenAccount
	err := q.db.QueryRowContext(ctx, `
SELECT rt.id, a.id, a.is_adult, a.role, a.status, a.token_version, a.created_at
FROM refresh_tokens rt
JOIN accounts a ON a.id = rt.account_id
WHERE rt.token_hash = $1
  AND rt.revoked_at IS NULL
  AND rt.expires_at > NOW()
  AND a.deleted_at IS NULL
`, tokenHash).Scan(
		&row.TokenID,
		&row.AccountID,
		&row.IsAdult,
		&row.Role,
		&row.Status,
		&row.TokenVersion,
		&row.CreatedAt,
	)
	return row, err
}

func (q *Queries) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE id = $1 AND revoked_at IS NULL
`, id)
	return err
}

func (q *Queries) RevokeRefreshTokensForAccount(ctx context.Context, accountID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE account_id = $1 AND revoked_at IS NULL
`, accountID)
	return err
}

// PurgeExpiredRefreshTokens borra tokens vencidos, o revocados hace más de
// una semana, para que la tabla no crezca sin límite (audit finding A8).
// Devuelve el número de filas eliminadas.
func (q *Queries) PurgeExpiredRefreshTokens(ctx context.Context) (int64, error) {
	res, err := q.db.ExecContext(ctx, `
DELETE FROM refresh_tokens
WHERE expires_at < NOW()
   OR (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '7 days')
`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
