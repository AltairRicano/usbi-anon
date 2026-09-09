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
// ResolveArcoRequest tampoco vuelven: el flujo ARCO se abandonó junto con el
// tutor, y F9 (ya cerrada) reescribió internal/auth sin necesitarlos —
// `arco_requests` no existe en ningún esquema vigente. (Útil)
package repository

import (
	"context"

	"github.com/google/uuid"
)

// NullUserInPseudonymizableLedgers scrubs the user from the append-only ledgers
// while preserving the rows for No-Repudio. usbi_app (el rol que ejecuta la
// cancelación de cuenta autoservicio) no tiene UPDATE directo en audit_log ni
// experience_history vía este camino — llama a la función SECURITY DEFINER
// null_user_in_pseudonymizable_ledgers (migración 0004), propiedad de
// usbi_moderador, en vez de tocar esas tablas por su cuenta. Corre dentro de
// la misma transacción de CancelAccount: no hace falta un segundo pool. (Útil)
func (q *Queries) NullUserInPseudonymizableLedgers(ctx context.Context, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`SELECT null_user_in_pseudonymizable_ledgers($1)`, userID)
	return err
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
// ../usbi ya aplicaba a `users`. (Útil)
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
//
// usbi_app no tiene DELETE directo en account_quiz_answers — llama a la
// función SECURITY DEFINER purge_account_quiz_answers (migración 0004),
// propiedad de usbi_moderador, en vez de tocar la tabla por su cuenta. Corre
// dentro de la misma transacción de CancelAccount: no hace falta un segundo
// pool. (Útil)
func (q *Queries) PurgeAccountQuizAnswers(ctx context.Context, accountID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`SELECT purge_account_quiz_answers($1)`, accountID)
	return err
}

// PurgeUserProgressData deletes the user's non-ledger personal progress data for
// data minimization on definitive cancellation. It intentionally excludes:
//   - experience_history / admin_audit_log — append-only ledgers, pseudonymized
//     via NullUserInPseudonymizableLedgers instead of deleted;
//   - sync_events — deleting it would cascade SET NULL onto
//     experience_history.sync_event_id, which the append-only trigger rejects. (Útil)
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
