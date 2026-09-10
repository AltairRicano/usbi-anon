package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
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
	ID          uuid.UUID
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

// GetSectionByIDAny, a diferencia de ListSections, no filtra por archived_at:
// la purga y el unarchive necesitan poder leer una sección archivada para
// decidir si la operación es válida (plan/05_Contenido_maker_y_juego.md §6).
func (q *Queries) GetSectionByIDAny(ctx context.Context, id uuid.UUID) (Section, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT id, title, description, color, created_by_admin_id, is_published, created_at, deleted_at, archived_at
FROM sections
WHERE id = $1
`, id)
	return scanSection(row)
}

func (q *Queries) ListArchivedSections(ctx context.Context) ([]Section, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT id, title, description, color, created_by_admin_id, is_published, created_at, deleted_at, archived_at
FROM sections
WHERE deleted_at IS NULL AND archived_at IS NOT NULL
ORDER BY archived_at DESC
`)
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
	return sections, rows.Err()
}

func (q *Queries) UnarchiveSection(ctx context.Context, id uuid.UUID) (Section, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE sections
SET archived_at = NULL
WHERE id = $1 AND deleted_at IS NULL AND archived_at IS NOT NULL
RETURNING id, title, description, color, created_by_admin_id, is_published, created_at, deleted_at, archived_at
`, id)
	return scanSection(row)
}

// CountLevelsBySection cuenta TODOS los niveles de la sección, archivados o
// no: levels.section_id es ON DELETE RESTRICT a propósito (ver comentario en
// la migración 0001, sección "CONTENIDO EDUCATIVO"), así que una sección con
// cualquier nivel restante — vivo o archivado — no se puede purgar todavía.
func (q *Queries) CountLevelsBySection(ctx context.Context, sectionID uuid.UUID) (int64, error) {
	var count int64
	err := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM levels WHERE section_id = $1`, sectionID).Scan(&count)
	return count, err
}

// PurgeSection es el DELETE físico — irreversible, libera almacenamiento de
// verdad. Solo puede ejecutarse sobre una sección ya archivada; RESTRICT en
// levels.section_id la rechaza además si queda cualquier nivel referenciándola
// (ver CountLevelsBySection). rowsAffected en 0 significa "no existe" o "no
// estaba archivada" — el servicio decide cuál con una lectura previa.
func (q *Queries) PurgeSection(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, `
DELETE FROM sections WHERE id = $1 AND deleted_at IS NULL AND archived_at IS NOT NULL
`, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
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

// GetLevelByIDAny, a diferencia de GetLevelByID, no filtra por deleted_at: la
// purga y el unarchive necesitan poder leer un nivel archivado (levels no
// tiene archived_at propio — su archivo ES deleted_at, ver comentario
// "CONTENIDO EDUCATIVO" en la migración 0001).
func (q *Queries) GetLevelByIDAny(ctx context.Context, id uuid.UUID) (Level, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT id, section_id, title, color, template_type, content, difficulty, is_published,
       created_by_admin_id, created_at, updated_at, deleted_at, deleted_by
FROM levels
WHERE id = $1
`, id)
	return scanLevel(row)
}

type ListArchivedLevelsParams struct {
	HasSectionID bool
	SectionID    uuid.UUID
}

func (q *Queries) ListArchivedLevels(ctx context.Context, arg ListArchivedLevelsParams) ([]ListLevelsRow, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT id, section_id, title, color, template_type, difficulty, is_published, created_at
FROM levels
WHERE deleted_at IS NOT NULL
  AND (NOT $1::boolean OR section_id = $2)
ORDER BY deleted_at DESC
`, arg.HasSectionID, arg.SectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []ListLevelsRow
	for rows.Next() {
		var item ListLevelsRow
		if err := rows.Scan(
			&item.ID, &item.SectionID, &item.Title, &item.Color,
			&item.TemplateType, &item.Difficulty, &item.IsPublished, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		levels = append(levels, item)
	}
	return levels, rows.Err()
}

func (q *Queries) UnarchiveLevel(ctx context.Context, id uuid.UUID) (Level, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE levels
SET deleted_at = NULL, deleted_by = NULL, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NOT NULL
RETURNING id, section_id, title, color, template_type, content, difficulty, is_published,
          created_by_admin_id, created_at, updated_at, deleted_at, deleted_by
`, id)
	return scanLevel(row)
}

// AccumulateRetiredProgressForLevel debe ejecutarse dentro de la misma
// transacción que PurgeLevel, y ANTES del DELETE — es la consulta exacta
// documentada en el comentario de account_retired_progress en la migración
// 0001. Sin esto, purgar un nivel le bajaría a cada jugador su contador de
// "niveles completados" e "intentos totales" (el XP no corre este riesgo:
// vive en experience_history, que sobrevive vía SET NULL).
func (q *Queries) AccumulateRetiredProgressForLevel(ctx context.Context, levelID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO account_retired_progress AS arp
    (user_id, levels_completed, attempts_total)
SELECT user_id,
       COUNT(*) FILTER (WHERE first_completed_at IS NOT NULL),
       COALESCE(SUM(attempts_count), 0)
FROM player_progress
WHERE level_id = $1
GROUP BY user_id
ON CONFLICT (user_id) DO UPDATE SET
    levels_completed = arp.levels_completed + EXCLUDED.levels_completed,
    attempts_total   = arp.attempts_total   + EXCLUDED.attempts_total,
    updated_at       = NOW()
`, levelID)
	return err
}

// PurgeLevel es el DELETE físico. Dispara los CASCADE de level_attempts y
// player_progress y el SET NULL de experience_history.level_id — por eso
// AccumulateRetiredProgressForLevel debe correr antes, en la misma tx. Solo
// afecta niveles ya archivados (deleted_at IS NOT NULL).
func (q *Queries) PurgeLevel(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, `DELETE FROM levels WHERE id = $1 AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
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

// GetUserProgressTotals suma niveles VIVOS (player_progress) con niveles ya
// PURGADOS por rotación de temporada (account_retired_progress) — ver
// plan/04_Rediseno_identidad_gustos.md §1.2. El XP no necesita esa suma: sale
// entero de experience_history, cuyas filas sobreviven con level_id NULL
// cuando su nivel se purga (SET NULL), así que SUM(xp_gained) ya cuenta el XP
// de niveles purgados sin ningún cambio. Antes de esta corrección
// completed_levels/total_attempts solo miraban player_progress y le "robaban"
// al jugador el progreso de cualquier nivel ya retirado.
func (q *Queries) GetUserProgressTotals(ctx context.Context, userID uuid.UUID) (UserProgressTotals, error) {
	var totals UserProgressTotals
	err := q.db.QueryRowContext(ctx, `
SELECT
    COALESCE((SELECT SUM(xp_gained)::int FROM experience_history WHERE user_id = $1), 0) AS total_xp,
    COALESCE((SELECT COUNT(*)::int FROM player_progress WHERE user_id = $1 AND first_completed_at IS NOT NULL), 0)
        + COALESCE((SELECT levels_completed FROM account_retired_progress WHERE user_id = $1), 0) AS completed_levels,
    COALESCE((SELECT SUM(attempts_count)::int FROM player_progress WHERE user_id = $1), 0)
        + COALESCE((SELECT attempts_total FROM account_retired_progress WHERE user_id = $1), 0) AS total_attempts
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

// ListUserProgressLevels intencionalmente sigue con INNER JOIN, no LEFT JOIN:
// player_progress.level_id es ON DELETE CASCADE (ver plan/04 §1.2), así que
// una fila de player_progress SIEMPRE tiene un nivel vivo detrás — cuando el
// nivel se purga, la fila de player_progress se va con él, no sobrevive con
// level_id NULL. La tolerancia a level_id NULL que pide §1.2 aplica a
// experience_history (SET NULL), no a esta consulta; ninguna consulta de
// historial existe todavía en este paquete — la añadirá la fase que primero
// necesite listar experience_history (F9 o F10), con su propio LEFT JOIN.
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

// ── Historial de sincronización del jugador (B4, estado_proyecto.md
// 2026-09-09) ────────────────────────────────────────────────────────────
// Corre exclusivamente sobre el pool de jugador: usbi_moderador no tiene
// ningún GRANT sobre sync_events, y no lo necesita — "cada quien ve solo lo
// suyo, un admin también" ya lo impone la base, no un filtro en Go. Sin
// SELECT del payload JSONB (decisión D3): es el delta completo de progreso
// y crece sin techo; si algún día hace falta depurar un evento concreto, se
// agrega un endpoint de detalle por ID, no se infla este listado.

type SyncEventSummary struct {
	ID              uuid.UUID
	DeviceID        uuid.UUID
	Status          string
	HmacValid       bool
	ReceivedAt      time.Time
	ProcessedAt     sql.NullTime
	RejectionReason sql.NullString
}

func scanSyncEventSummary(row scanner) (SyncEventSummary, error) {
	var e SyncEventSummary
	err := row.Scan(&e.ID, &e.DeviceID, &e.Status, &e.HmacValid, &e.ReceivedAt, &e.ProcessedAt, &e.RejectionReason)
	return e, err
}

const syncEventSummaryColumns = `id, device_id, status, hmac_valid, received_at, processed_at, rejection_reason`

type ListSyncEventsForUserParams struct {
	UserID   uuid.UUID
	DeviceID uuid.NullUUID
	Cursor   sql.NullTime // received_at del último elemento ya visto; NULL pide la primera página
	PageSize int32
}

// ListSyncEventsForUser pagina por received_at DESC — el orden en que un
// jugador esperaría ver "mis últimas sincronizaciones" — apoyado en
// sync_events_user_received_idx (migración 0006). El filtro por device_id
// es opcional.
func (q *Queries) ListSyncEventsForUser(ctx context.Context, arg ListSyncEventsForUserParams) ([]SyncEventSummary, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT `+syncEventSummaryColumns+`
FROM sync_events
WHERE user_id = $1
  AND ($2::uuid IS NULL OR device_id = $2)
  AND ($3::timestamptz IS NULL OR received_at < $3)
ORDER BY received_at DESC, id DESC
LIMIT $4
`, arg.UserID, arg.DeviceID, arg.Cursor, arg.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SyncEventSummary
	for rows.Next() {
		item, err := scanSyncEventSummary(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ── Lectura de audit_log (B1, estado_proyecto.md 2026-09-09) ───────────────
// Corre exclusivamente sobre el pool de moderador — usbi_app tiene REVOKE
// SELECT sobre audit_log desde el 2026-09-02 (solo INSERT). Requiere el
// GRANT SELECT nuevo documentado en 00_roles_unificado.sql.

type AuditLogEntry struct {
	ID             uuid.UUID
	ActorAccountID uuid.NullUUID
	Action         string
	EntityType     string
	EntityID       uuid.NullUUID
	BeforeState    pqtype.NullRawMessage
	AfterState     pqtype.NullRawMessage
	IPAddress      string
	UserAgent      string
	CreatedAt      time.Time
}

func scanAuditLogEntry(row scanner) (AuditLogEntry, error) {
	var e AuditLogEntry
	err := row.Scan(&e.ID, &e.ActorAccountID, &e.Action, &e.EntityType, &e.EntityID,
		&e.BeforeState, &e.AfterState, &e.IPAddress, &e.UserAgent, &e.CreatedAt)
	return e, err
}

const auditLogColumns = `id, actor_account_id, action, entity_type, entity_id, before_state, after_state, ip_address, user_agent, created_at`

// ListAuditLogParams filtra por rango de fechas, actor, action y
// entity_type. Todos los filtros son opcionales (NULL = sin filtrar).
//
// Cursor compuesto (created_at, id), NO cursor por id como suggestions: los
// ID de audit_log los genera audit.Log con uuid.New() (v4 aleatorio), no
// UUIDv7 — ORDER BY id DESC no aproxima ningún orden temporal aquí.
type ListAuditLogParams struct {
	ActorAccountID uuid.NullUUID
	Action         sql.NullString
	EntityType     sql.NullString
	From           sql.NullTime
	To             sql.NullTime
	CursorTime     sql.NullTime
	CursorID       uuid.NullUUID
	PageSize       int32
}

func (q *Queries) ListAuditLog(ctx context.Context, arg ListAuditLogParams) ([]AuditLogEntry, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT `+auditLogColumns+`
FROM audit_log
WHERE ($1::uuid IS NULL OR actor_account_id = $1)
  AND ($2::text IS NULL OR action = $2)
  AND ($3::text IS NULL OR entity_type = $3)
  AND ($4::timestamptz IS NULL OR created_at >= $4)
  AND ($5::timestamptz IS NULL OR created_at <= $5)
  AND (
    $6::timestamptz IS NULL
    OR created_at < $6
    OR (created_at = $6 AND id < $7)
  )
ORDER BY created_at DESC, id DESC
LIMIT $8
`, arg.ActorAccountID, arg.Action, arg.EntityType, arg.From, arg.To,
		arg.CursorTime, arg.CursorID, arg.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AuditLogEntry
	for rows.Next() {
		item, err := scanAuditLogEntry(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
