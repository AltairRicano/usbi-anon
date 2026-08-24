package repository

import (
	"context"
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
