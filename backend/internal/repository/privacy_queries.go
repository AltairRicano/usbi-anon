// F2 (dos bases) había migrado a identityrepo InsertArcoRequest,
// InsertTutorConsent, ActivateTutorConsentUser, ListPendingArcoRequests,
// GetArcoRequestForUpdate, ResolveArcoRequest, PseudonymizeUser y
// PseudonymizeTutorConsents; NullUserInPseudonymizableLedgers y
// PurgeUserProgressData se quedaron aquí por operar sobre tablas de la base
// principal. Con el rediseño de identidad (plan/04_Rediseno_identidad_gustos.md)
// esa distinción desaparece — ya solo hay una base — y F7 devuelve a este
// archivo lo que necesita internal/privacy.CancelAccount: DeactivateAccount y
// PurgeAccountQuizAnswers reemplazan a PseudonymizeUser/InsertTutorConsent
// (ya no hay email que ofuscar ni tutor que pseudonimizar; cancelar sobrescribe
// el nickname — ver §1.1 punto 2). InsertTutorConsent, ActivateTutorConsentUser
// y PseudonymizeTutorConsents no vuelven: el flujo de tutor se eliminó
// completo. InsertArcoRequest/ListPendingArcoRequests/GetArcoRequestForUpdate/
// ResolveArcoRequest tampoco vuelven aquí todavía — F9 los añade junto con
// internal/auth.Arco/ListPendingArco/ResolveArco, que son quienes los usan.
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
		`UPDATE audit_log SET actor_account_id = NULL WHERE actor_account_id = $1`, userID); err != nil {
		return err
	}
	return nil
}

type DeactivateAccountParams struct {
	ID             uuid.UUID
	RandomNickname string // relleno [a-z0-9] de 20 caracteres, generado por el llamador
	DeletionReason string
}

// DeactivateAccount reemplaza a la PseudonymizeUser del diseño de dos bases:
// ya no hay email que ofuscar, así que "cancelar" es sobrescribir el
// nickname con relleno aleatorio (§1.1 punto 2 — libera el valor original
// para reuso y no deja rastro de las respuestas que lo originaron), marcar
// status='deleted' y forzar token_version+1 para invalidar cualquier JWT
// vivo. La fila NUNCA se borra: es la misma garantía de no repudio que
// ../usbi ya aplicaba a `users`.
func (q *Queries) DeactivateAccount(ctx context.Context, arg DeactivateAccountParams) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE accounts
SET nickname        = $2,
    status          = 'deleted',
    deleted_at      = COALESCE(deleted_at, NOW()),
    deletion_reason = $3,
    token_version   = token_version + 1,
    updated_at      = NOW()
WHERE id = $1 AND deleted_at IS NULL
`, arg.ID, arg.RandomNickname, arg.DeletionReason)
	return err
}

// PurgeAccountQuizAnswers borra las respuestas del cuestionario de registro.
// account_quiz_answers existe SOLO para que un admin pueda comparar
// respuestas y recuperar una cuenta viva (§1 decisión 4) — una vez cancelada
// la cuenta esa recuperación ya no aplica, así que conservarlas sería
// retener datos sin propósito. No es una bitácora append-only: a diferencia
// de experience_history/audit_log, aquí sí toca DELETE, no SET NULL.
func (q *Queries) PurgeAccountQuizAnswers(ctx context.Context, accountID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`DELETE FROM account_quiz_answers WHERE account_id = $1`, accountID)
	return err
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
