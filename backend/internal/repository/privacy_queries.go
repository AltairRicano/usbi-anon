package repository

import (
	"context"

	"github.com/google/uuid"
)

// NullUserInPseudonymizableLedgers scrubs the user from append-only ledgers
// while preserving rows for non-repudiation.
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

// DeactivateAccount cancela la cuenta sobrescribiendo el nickname con un valor
// aleatorio (para permitir su reuso sin dejar rastro), marca status='deleted' y
// forzar la invalidación del token_version. No elimina la fila por garantía de no-repudio.
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

// PurgeAccountQuizAnswers borra de manera definitiva las respuestas del cuestionario
// asociadas a la cuenta.
//
// usbi_app no tiene DELETE directo en account_quiz_answers — llama a la
// función SECURITY DEFINER purge_account_quiz_answers (migración 0004),
// propiedad de usbi_moderador, en vez de tocar la tabla por su cuenta. Corre
// dentro de la misma transacción de CancelAccount: no hace falta un segundo
// pool.
func (q *Queries) PurgeAccountQuizAnswers(ctx context.Context, accountID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`SELECT purge_account_quiz_answers($1)`, accountID)
	return err
}

// PurgeUserProgressData elimina los datos de progreso personal del usuario.
// Excluye intencionalmente las tablas de auditoría (experience_history, admin_audit_log, sync_events)
// ya que estas se anonimizan vía NullUserInPseudonymizableLedgers.
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
