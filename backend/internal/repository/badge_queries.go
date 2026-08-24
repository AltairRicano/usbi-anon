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
