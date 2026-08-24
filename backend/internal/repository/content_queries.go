package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrTransactionsUnsupported = errors.New("repository: configured DBTX does not support transactions")

type txBeginner interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

func (q *Queries) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	db, ok := q.db.(txBeginner)
	if !ok {
		return nil, ErrTransactionsUnsupported
	}
	return db.BeginTx(ctx, opts)
}

type CreateSectionReturningParams struct {
	ID               uuid.UUID
	Title            string
	Description      string
	Color            string
	IsPublished      bool
	CreatedByAdminID uuid.NullUUID
}

func (q *Queries) CreateSectionReturning(ctx context.Context, arg CreateSectionReturningParams) (Section, error) {
	row := q.db.QueryRowContext(ctx, `
INSERT INTO sections (id, title, description, color, is_published, created_by_admin_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, title, description, color, created_by_admin_id, is_published, created_at, deleted_at, archived_at
`, arg.ID, arg.Title, arg.Description, arg.Color, arg.IsPublished, arg.CreatedByAdminID)
	return scanSection(row)
}

type ListSectionsParams struct {
	IncludeUnpublished bool
}

func (q *Queries) ListSections(ctx context.Context, arg ListSectionsParams) ([]Section, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT id, title, description, color, created_by_admin_id, is_published, created_at, deleted_at, archived_at
FROM sections
WHERE deleted_at IS NULL
  AND archived_at IS NULL
  AND ($1::boolean OR is_published = true)
ORDER BY created_at DESC
`, arg.IncludeUnpublished)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []Section
	for rows.Next() {
		section, err := scanSection(rows)
		if err != nil {
			return nil, err
		}
		sections = append(sections, section)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sections, nil
}

type UpdateSectionParams struct {
	ID    uuid.UUID
	Title       string
	Description string
	Color       string
}

func (q *Queries) UpdateSection(ctx context.Context, arg UpdateSectionParams) (Section, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE sections
SET title = $2, description = $3, color = $4
WHERE id = $1 AND deleted_at IS NULL AND archived_at IS NULL
RETURNING id, title, description, color, created_by_admin_id, is_published, created_at, deleted_at, archived_at
`, arg.ID, arg.Title, arg.Description, arg.Color)
	return scanSection(row)
}

func (q *Queries) PublishSection(ctx context.Context, id uuid.UUID) (Section, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE sections
SET is_published = true
WHERE id = $1 AND deleted_at IS NULL AND archived_at IS NULL
RETURNING id, title, description, color, created_by_admin_id, is_published, created_at, deleted_at, archived_at
`, id)
	return scanSection(row)
}

func (q *Queries) UnpublishSection(ctx context.Context, id uuid.UUID) (Section, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE sections
SET is_published = false
WHERE id = $1 AND deleted_at IS NULL AND archived_at IS NULL
RETURNING id, title, description, color, created_by_admin_id, is_published, created_at, deleted_at, archived_at
`, id)
	return scanSection(row)
}

func (q *Queries) ArchiveSection(ctx context.Context, id uuid.UUID) (Section, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE sections
SET archived_at = NOW(), is_published = false
WHERE id = $1 AND deleted_at IS NULL AND archived_at IS NULL
RETURNING id, title, description, color, created_by_admin_id, is_published, created_at, deleted_at, archived_at
`, id)
	return scanSection(row)
}

type CreateLevelReturningParams struct {
	ID               uuid.UUID
	SectionID        uuid.UUID
	Title            string
	Color            string
	TemplateType     string
	Content          json.RawMessage
	Difficulty       int32
	IsPublished      bool
	CreatedByAdminID uuid.NullUUID
}

func (q *Queries) CreateLevelReturning(ctx context.Context, arg CreateLevelReturningParams) (Level, error) {
	row := q.db.QueryRowContext(ctx, `
INSERT INTO levels (
    id, section_id, title, color, template_type, content, difficulty, is_published, created_by_admin_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id, section_id, title, color, template_type, content, difficulty, is_published,
          created_by_admin_id, created_at, updated_at, deleted_at, deleted_by
`, arg.ID, arg.SectionID, arg.Title, arg.Color, arg.TemplateType, []byte(arg.Content), arg.Difficulty, arg.IsPublished, arg.CreatedByAdminID)
	return scanLevel(row)
}

type ListLevelsParams struct {
	IncludeUnpublished bool
	HasSectionID       bool
	SectionID          uuid.UUID
	Cursor             uuid.UUID
	PageSize           int32
}

type ListLevelsRow struct {
	ID           uuid.UUID
	SectionID    uuid.UUID
	Title        string
	Color        string
	TemplateType string
	Difficulty   int32
	IsPublished  bool
	CreatedAt    time.Time
}

func (q *Queries) ListLevels(ctx context.Context, arg ListLevelsParams) ([]ListLevelsRow, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT id, section_id, title, color, template_type, difficulty, is_published, created_at
FROM levels
WHERE deleted_at IS NULL
  AND ($1::boolean OR is_published = true)
  AND (NOT $2::boolean OR section_id = $3)
  AND (id > $4 OR $4 = '00000000-0000-0000-0000-000000000000'::uuid)
ORDER BY id ASC
LIMIT $5
`, arg.IncludeUnpublished, arg.HasSectionID, arg.SectionID, arg.Cursor, arg.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []ListLevelsRow
	for rows.Next() {
		var item ListLevelsRow
		if err := rows.Scan(
			&item.ID,
			&item.SectionID,
			&item.Title,
			&item.Color,
			&item.TemplateType,
			&item.Difficulty,
			&item.IsPublished,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		levels = append(levels, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return levels, nil
}

func (q *Queries) GetLevelByID(ctx context.Context, id uuid.UUID) (Level, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT id, section_id, title, color, template_type, content, difficulty, is_published,
       created_by_admin_id, created_at, updated_at, deleted_at, deleted_by
FROM levels
WHERE id = $1 AND deleted_at IS NULL
`, id)
	return scanLevel(row)
}

type UpdateLevelParams struct {
	ID           uuid.UUID
	Title        string
	Color        string
	TemplateType string
	Content      json.RawMessage
	Difficulty   int32
}

func (q *Queries) UpdateLevel(ctx context.Context, arg UpdateLevelParams) (Level, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE levels
SET title = $2,
    color = $3,
    template_type = $4,
    content = $5,
    difficulty = $6,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, section_id, title, color, template_type, content, difficulty, is_published,
          created_by_admin_id, created_at, updated_at, deleted_at, deleted_by
`, arg.ID, arg.Title, arg.Color, arg.TemplateType, []byte(arg.Content), arg.Difficulty)
	return scanLevel(row)
}

func (q *Queries) PublishLevel(ctx context.Context, id uuid.UUID) (Level, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE levels
SET is_published = true, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, section_id, title, color, template_type, content, difficulty, is_published,
          created_by_admin_id, created_at, updated_at, deleted_at, deleted_by
`, id)
	return scanLevel(row)
}

func (q *Queries) UnpublishLevel(ctx context.Context, id uuid.UUID) (Level, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE levels
SET is_published = false, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, section_id, title, color, template_type, content, difficulty, is_published,
          created_by_admin_id, created_at, updated_at, deleted_at, deleted_by
`, id)
	return scanLevel(row)
}

type ArchiveLevelParams struct {
	ID        uuid.UUID
	DeletedBy uuid.NullUUID
}

func (q *Queries) ArchiveLevel(ctx context.Context, arg ArchiveLevelParams) (Level, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE levels
SET deleted_at = NOW(), deleted_by = $2, is_published = false, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, section_id, title, color, template_type, content, difficulty, is_published,
          created_by_admin_id, created_at, updated_at, deleted_at, deleted_by
`, arg.ID, arg.DeletedBy)
	return scanLevel(row)
}

type ArchiveLevelsBySectionParams struct {
	SectionID uuid.UUID
	DeletedBy uuid.NullUUID
}

func (q *Queries) ArchiveLevelsBySection(ctx context.Context, arg ArchiveLevelsBySectionParams) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE levels
SET deleted_at = NOW(), deleted_by = $2, is_published = false, updated_at = NOW()
WHERE section_id = $1 AND deleted_at IS NULL
`, arg.SectionID, arg.DeletedBy)
	return err
}

type LockLevelAttemptParams struct {
	UserID      uuid.UUID
	LevelID     uuid.UUID
	AttemptDate time.Time
}

func (q *Queries) LockLevelAttempt(ctx context.Context, arg LockLevelAttemptParams) error {
	key := fmt.Sprintf("%s:%s:%s", arg.UserID, arg.LevelID, arg.AttemptDate.Format("2006-01-02"))
	_, err := q.db.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, key)
	return err
}

type CountLevelAttemptsByDateParams struct {
	UserID      uuid.UUID
	LevelID     uuid.UUID
	AttemptDate time.Time
}

func (q *Queries) CountLevelAttemptsByDate(ctx context.Context, arg CountLevelAttemptsByDateParams) (int64, error) {
	var count int64
	err := q.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM level_attempts
WHERE user_id = $1 AND level_id = $2 AND attempt_date = $3
`, arg.UserID, arg.LevelID, arg.AttemptDate).Scan(&count)
	return count, err
}

type UpsertPlayerProgressForAttemptParams struct {
	UserID          uuid.UUID
	LevelID         uuid.UUID
	BestScore       int32
	XpTotalForLevel int32
	Completed       bool
}

func (q *Queries) UpsertPlayerProgressForAttempt(ctx context.Context, arg UpsertPlayerProgressForAttemptParams) error {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO player_progress (
    user_id, level_id, best_score, xp_total_for_level, attempts_count, first_completed_at, last_completed_at
) VALUES (
    $1, $2, $3, $4, 1,
    CASE WHEN $5::boolean THEN NOW() ELSE NULL END,
    CASE WHEN $5::boolean THEN NOW() ELSE NULL END
) ON CONFLICT (user_id, level_id) DO UPDATE SET
    best_score = GREATEST(player_progress.best_score, EXCLUDED.best_score),
    xp_total_for_level = player_progress.xp_total_for_level + EXCLUDED.xp_total_for_level,
    attempts_count = player_progress.attempts_count + 1,
    first_completed_at = CASE
        WHEN $5::boolean AND player_progress.first_completed_at IS NULL THEN NOW()
        ELSE player_progress.first_completed_at
    END,
    last_completed_at = CASE
        WHEN $5::boolean THEN NOW()
        ELSE player_progress.last_completed_at
    END
`, arg.UserID, arg.LevelID, arg.BestScore, arg.XpTotalForLevel, arg.Completed)
	return err
}

type InsertSyncEventWithPayloadParams struct {
	ID               uuid.UUID
	DeviceID         uuid.UUID
	UserID           uuid.UUID
	Payload          json.RawMessage
	PayloadHash      []byte
	HmacSignature    []byte
	CryptoKeyVersion int16
	HmacValid        bool
	Status           string
}

func (q *Queries) InsertSyncEventWithPayload(ctx context.Context, arg InsertSyncEventWithPayloadParams) error {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO sync_events (
    id, device_id, user_id, payload, payload_hash, hmac_signature,
    crypto_key_version, hmac_valid, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
`, arg.ID, arg.DeviceID, arg.UserID, []byte(arg.Payload), arg.PayloadHash, arg.HmacSignature, arg.CryptoKeyVersion, arg.HmacValid, arg.Status)
	return err
}

type UpdateSyncEventRejectedParams struct {
	ID     uuid.UUID
	Reason string
}

func (q *Queries) UpdateSyncEventRejected(ctx context.Context, arg UpdateSyncEventRejectedParams) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE sync_events
SET status = 'rejected', processed_at = NOW(), rejection_reason = $2
WHERE id = $1
`, arg.ID, arg.Reason)
	return err
}

type UserProgressTotals struct {
	TotalXP         int32
	CompletedLevels int32
	TotalAttempts   int32
}

func (q *Queries) GetUserProgressTotals(ctx context.Context, userID uuid.UUID) (UserProgressTotals, error) {
	var totals UserProgressTotals
	err := q.db.QueryRowContext(ctx, `
SELECT
    COALESCE((SELECT SUM(xp_gained)::int FROM experience_history WHERE user_id = $1), 0) AS total_xp,
    COALESCE((SELECT COUNT(*)::int FROM player_progress WHERE user_id = $1 AND first_completed_at IS NOT NULL), 0) AS completed_levels,
    COALESCE((SELECT SUM(attempts_count)::int FROM player_progress WHERE user_id = $1), 0) AS total_attempts
`, userID).Scan(&totals.TotalXP, &totals.CompletedLevels, &totals.TotalAttempts)
	return totals, err
}

func (q *Queries) ListDailyStreakDates(ctx context.Context, userID uuid.UUID, limit int32) ([]time.Time, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT activity_date
FROM daily_streak
WHERE user_id = $1
ORDER BY activity_date DESC
LIMIT $2
`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []time.Time
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err != nil {
			return nil, err
		}
		dates = append(dates, date)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dates, nil
}

type UserProgressLevelRow struct {
	LevelID          uuid.UUID
	Title            string
	TemplateType     string
	Difficulty       int32
	BestScore        int32
	XpTotalForLevel  int32
	AttemptsCount    int32
	FirstCompletedAt sql.NullTime
	LastCompletedAt  sql.NullTime
}

func (q *Queries) ListUserProgressLevels(ctx context.Context, userID uuid.UUID) ([]UserProgressLevelRow, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT l.id, l.title, l.template_type, l.difficulty,
       pp.best_score, pp.xp_total_for_level, pp.attempts_count,
       pp.first_completed_at, pp.last_completed_at
FROM player_progress pp
JOIN levels l ON l.id = pp.level_id
WHERE pp.user_id = $1
ORDER BY pp.last_completed_at DESC NULLS LAST, l.title ASC
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []UserProgressLevelRow
	for rows.Next() {
		var item UserProgressLevelRow
		if err := rows.Scan(
			&item.LevelID,
			&item.Title,
			&item.TemplateType,
			&item.Difficulty,
			&item.BestScore,
			&item.XpTotalForLevel,
			&item.AttemptsCount,
			&item.FirstCompletedAt,
			&item.LastCompletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanSection(row scanner) (Section, error) {
	var section Section
	err := row.Scan(
		&section.ID,
		&section.Title,
		&section.Description,
		&section.Color,
		&section.CreatedByAdminID,
		&section.IsPublished,
		&section.CreatedAt,
		&section.DeletedAt,
		&section.ArchivedAt,
	)
	return section, err
}

func scanLevel(row scanner) (Level, error) {
	var level Level
	err := row.Scan(
		&level.ID,
		&level.SectionID,
		&level.Title,
		&level.Color,
		&level.TemplateType,
		&level.Content,
		&level.Difficulty,
		&level.IsPublished,
		&level.CreatedByAdminID,
		&level.CreatedAt,
		&level.UpdatedAt,
		&level.DeletedAt,
		&level.DeletedBy,
	)
	return level, err
}
