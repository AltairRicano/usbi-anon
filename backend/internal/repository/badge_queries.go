package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type BadgeWithEarnedAt struct {
	ID          uuid.UUID
	Name        string
	XpThreshold int32
	IconKey     string
	EarnedAt    time.Time
}

type AwardEligibleBadgesParams struct {
	UserID  uuid.UUID
	TotalXP int32
}

func (q *Queries) AwardEligibleBadges(ctx context.Context, arg AwardEligibleBadgesParams) ([]BadgeWithEarnedAt, error) {
	rows, err := q.db.QueryContext(ctx, `
WITH awarded AS (
    INSERT INTO user_badges (user_id, badge_id)
    SELECT $1, b.id
    FROM badges b
    WHERE b.xp_threshold <= $2
    ON CONFLICT (user_id, badge_id) DO NOTHING
    RETURNING badge_id, earned_at
)
SELECT b.id, b.name, b.xp_threshold, b.icon_key, awarded.earned_at
FROM awarded
JOIN badges b ON b.id = awarded.badge_id
ORDER BY b.xp_threshold ASC
`, arg.UserID, arg.TotalXP)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var badges []BadgeWithEarnedAt
	for rows.Next() {
		var badge BadgeWithEarnedAt
		if err := rows.Scan(&badge.ID, &badge.Name, &badge.XpThreshold, &badge.IconKey, &badge.EarnedAt); err != nil {
			return nil, err
		}
		badges = append(badges, badge)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return badges, nil
}

func (q *Queries) ListUserBadges(ctx context.Context, userID uuid.UUID) ([]BadgeWithEarnedAt, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT b.id, b.name, b.xp_threshold, b.icon_key, ub.earned_at
FROM user_badges ub
JOIN badges b ON b.id = ub.badge_id
WHERE ub.user_id = $1
ORDER BY b.xp_threshold ASC, ub.earned_at ASC
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var badges []BadgeWithEarnedAt
	for rows.Next() {
		var badge BadgeWithEarnedAt
		if err := rows.Scan(&badge.ID, &badge.Name, &badge.XpThreshold, &badge.IconKey, &badge.EarnedAt); err != nil {
			return nil, err
		}
		badges = append(badges, badge)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return badges, nil
}

// ── Admin CRUD (B3, estado_proyecto.md 2026-09-09) ─────────────────────────
// Corre exclusivamente sobre el pool de moderador: usbi_app perdió
// INSERT/UPDATE/DELETE en badges desde el 2026-09-02 (decisión de producto:
// el equipo de USBI gestiona insignias sin depender de un ingeniero).

type Badge struct {
	ID          uuid.UUID
	Name        string
	XpThreshold int32
	IconKey     string
}

func scanBadge(row scanner) (Badge, error) {
	var b Badge
	err := row.Scan(&b.ID, &b.Name, &b.XpThreshold, &b.IconKey)
	return b, err
}

const badgeColumns = `id, name, xp_threshold, icon_key`

// ListBadges ordena por xp_threshold: es el mismo orden en que se conceden
// (AwardEligibleBadges) y en que ListUserBadges se las muestra al jugador.
func (q *Queries) ListBadges(ctx context.Context) ([]Badge, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT `+badgeColumns+`
FROM badges
ORDER BY xp_threshold ASC, name ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Badge
	for rows.Next() {
		item, err := scanBadge(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type CreateBadgeParams struct {
	ID          uuid.UUID
	Name        string
	XpThreshold int32
	IconKey     string
}

func (q *Queries) CreateBadge(ctx context.Context, arg CreateBadgeParams) (Badge, error) {
	row := q.db.QueryRowContext(ctx, `
INSERT INTO badges (id, name, xp_threshold, icon_key)
VALUES ($1, $2, $3, $4)
RETURNING `+badgeColumns,
		arg.ID, arg.Name, arg.XpThreshold, arg.IconKey)
	return scanBadge(row)
}

type UpdateBadgeParams struct {
	ID          uuid.UUID
	Name        string
	XpThreshold int32
	IconKey     string
}

func (q *Queries) UpdateBadge(ctx context.Context, arg UpdateBadgeParams) (Badge, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE badges
SET name = $2, xp_threshold = $3, icon_key = $4
WHERE id = $1
RETURNING `+badgeColumns,
		arg.ID, arg.Name, arg.XpThreshold, arg.IconKey)
	return scanBadge(row)
}

// CountUserBadgesByBadge soporta el mismo guard explícito que
// CountInterestLinksByCategory: en vez de dejar que el DELETE choque con el
// ON DELETE RESTRICT de user_badges.badge_id y traducir el código de error
// de Postgres, el servicio pregunta primero y decide con un error de
// dominio propio. Una insignia ya ganada nunca se revoca (misma lógica de
// negocio que "retirar un nivel no quita XP").
func (q *Queries) CountUserBadgesByBadge(ctx context.Context, badgeID uuid.UUID) (int64, error) {
	var count int64
	err := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_badges WHERE badge_id = $1`, badgeID).Scan(&count)
	return count, err
}

func (q *Queries) DeleteBadge(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, `DELETE FROM badges WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
