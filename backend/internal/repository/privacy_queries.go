// El resto de privacy_queries.go (InsertArcoRequest, InsertTutorConsent,
// ActivateTutorConsentUser, ListPendingArcoRequests, GetArcoRequestForUpdate,
// ResolveArcoRequest, PseudonymizeUser, PseudonymizeTutorConsents) migró
// íntegro a identityrepo en F2 — ver plan/02_Backend.md §3.2 y §3.3. Estas
// dos funciones se quedan aquí porque operan sobre tablas de la base
// principal.
package repository

import (
	"context"

	"github.com/google/uuid"
)

// NullUserInPseudonymizableLedgers scrubs the user from the append-only ledgers
// while preserving the rows for No-Repudio. Run as two separate statements: with
// lib/pq's extended protocol a single parameterised query may contain only one
// command, and each UPDATE matches exactly the SET-NULL pattern the append-only
// trigger permits.
func (q *Queries) NullUserInPseudonymizableLedgers(ctx context.Context, userID uuid.UUID) error {
	if _, err := q.db.ExecContext(ctx,
		`UPDATE experience_history SET user_id = NULL WHERE user_id = $1`, userID); err != nil {
		return err
	}
	if _, err := q.db.ExecContext(ctx,
		`UPDATE admin_audit_log SET actor_user_id = NULL WHERE actor_user_id = $1`, userID); err != nil {
		return err
	}
	return nil
}

// PurgeUserProgressData deletes the user's non-ledger personal progress data for
// data minimization on definitive cancellation. It intentionally excludes:
//   - experience_history / admin_audit_log — append-only ledgers, pseudonymized
//     via NullUserInPseudonymizableLedgers instead of deleted;
//   - sync_events — deleting it would cascade SET NULL onto
//     experience_history.sync_event_id, which the append-only trigger rejects.
func (q *Queries) PurgeUserProgressData(ctx context.Context, userID uuid.UUID) error {
	stmts := []string{
		`DELETE FROM player_progress WHERE user_id = $1`,
		`DELETE FROM level_attempts WHERE user_id = $1`,
		`DELETE FROM daily_streak WHERE user_id = $1`,
		`DELETE FROM user_badges WHERE user_id = $1`,
	}
	for _, stmt := range stmts {
		if _, err := q.db.ExecContext(ctx, stmt, userID); err != nil {
			return err
		}
	}
	return nil
}
