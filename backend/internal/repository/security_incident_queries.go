package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type InsertSecurityIncidentParams struct {
	ID                 uuid.UUID
	DetectedAt         time.Time
	Severity           string
	AffectedScope      string
	Description        string
	ContainmentActions string
	ReportedToCutai    bool
	EvidenceHash       []byte
}

// InsertSecurityIncident appends a row to the 5-year security incident log
// mandated by the Documento de Seguridad. Nullable follow-up columns
// (reported_at, resolved_at, notified_*) start NULL and are filled later.
func (q *Queries) InsertSecurityIncident(ctx context.Context, arg InsertSecurityIncidentParams) error {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO security_incidents (
    id, detected_at, severity, affected_scope, description,
    containment_actions, reported_to_cutai, evidence_hash
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
`, arg.ID, arg.DetectedAt, arg.Severity, arg.AffectedScope, arg.Description,
		arg.ContainmentActions, arg.ReportedToCutai, arg.EvidenceHash)
	return err
}

// ── Lectura y edición (B2, estado_proyecto.md 2026-09-09) ──────────────────
// Corre exclusivamente sobre el pool de moderador. usbi_app tiene REVOKE ALL
// sobre security_incidents desde 2026-09-08; usbi_moderador gana SELECT y
// UPDATE de tabla completa (decisión D1: el PATCH puede resellar la
// narrativa, no solo las columnas de seguimiento) — nunca DELETE, para
// ningún rol, reforzado además por un trigger BEFORE DELETE (migración
// 0006).

type SecurityIncident struct {
	ID                   uuid.UUID
	DetectedAt           time.Time
	ReportedAt           sql.NullTime
	Severity             string
	AffectedScope        string
	Description          string
	ContainmentActions   string
	ResolvedAt           sql.NullTime
	ReportedToCutai      bool
	NotifiedToCutaiAt    sql.NullTime
	NotifiedToSubjectsAt sql.NullTime
	EvidenceHash         []byte
}

func scanSecurityIncident(row scanner) (SecurityIncident, error) {
	var i SecurityIncident
	err := row.Scan(&i.ID, &i.DetectedAt, &i.ReportedAt, &i.Severity, &i.AffectedScope,
		&i.Description, &i.ContainmentActions, &i.ResolvedAt, &i.ReportedToCutai,
		&i.NotifiedToCutaiAt, &i.NotifiedToSubjectsAt, &i.EvidenceHash)
	return i, err
}

const securityIncidentColumns = `
    id, detected_at, reported_at, severity, affected_scope, description,
    containment_actions, resolved_at, reported_to_cutai, notified_to_cutai_at,
    notified_to_subjects_at, evidence_hash`

type ListSecurityIncidentsParams struct {
	CursorTime sql.NullTime
	CursorID   uuid.NullUUID
	PageSize   int32
}

// ListSecurityIncidents pagina por (detected_at, id) DESC — igual que
// audit_log, el id es uuid.New() (v4 aleatorio, ver incidents.CreateIncident),
// así que ordenar por id no aproxima ningún orden temporal.
func (q *Queries) ListSecurityIncidents(ctx context.Context, arg ListSecurityIncidentsParams) ([]SecurityIncident, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT `+securityIncidentColumns+`
FROM security_incidents
WHERE (
    $1::timestamptz IS NULL
    OR detected_at < $1
    OR (detected_at = $1 AND id < $2)
)
ORDER BY detected_at DESC, id DESC
LIMIT $3
`, arg.CursorTime, arg.CursorID, arg.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SecurityIncident
	for rows.Next() {
		item, err := scanSecurityIncident(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (q *Queries) GetSecurityIncident(ctx context.Context, id uuid.UUID) (SecurityIncident, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT `+securityIncidentColumns+`
FROM security_incidents
WHERE id = $1
`, id)
	return scanSecurityIncident(row)
}

type UpdateSecurityIncidentParams struct {
	ID                   uuid.UUID
	Severity             string
	AffectedScope        string
	Description          string
	ContainmentActions   string
	ReportedAt           sql.NullTime
	ResolvedAt           sql.NullTime
	ReportedToCutai      bool
	NotifiedToCutaiAt    sql.NullTime
	NotifiedToSubjectsAt sql.NullTime
	EvidenceHash         []byte
}

// UpdateSecurityIncident actualiza la narrativa (severity/affected_scope/
// description/containment_actions), las 5 columnas de seguimiento, y
// resella evidence_hash con el nuevo contenido — decisión D1
// (estado_proyecto.md 2026-09-09). detected_at NUNCA se edita: es el hecho
// histórico de cuándo se detectó, no parte de la narrativa que se corrige.
func (q *Queries) UpdateSecurityIncident(ctx context.Context, arg UpdateSecurityIncidentParams) (SecurityIncident, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE security_incidents
SET severity = $2, affected_scope = $3, description = $4, containment_actions = $5,
    reported_at = $6, resolved_at = $7, reported_to_cutai = $8,
    notified_to_cutai_at = $9, notified_to_subjects_at = $10, evidence_hash = $11
WHERE id = $1
RETURNING `+securityIncidentColumns,
		arg.ID, arg.Severity, arg.AffectedScope, arg.Description, arg.ContainmentActions,
		arg.ReportedAt, arg.ResolvedAt, arg.ReportedToCutai, arg.NotifiedToCutaiAt,
		arg.NotifiedToSubjectsAt, arg.EvidenceHash)
	return scanSecurityIncident(row)
}
