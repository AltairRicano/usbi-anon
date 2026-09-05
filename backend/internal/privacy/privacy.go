// Package privacy contiene la cancelación definitiva de una cuenta.
//
// Hasta F5 (dos bases separadas) esto era una saga reanudable de dos fases
// —PurgeMain [P] y PseudonymizeIdentity [I], cada una en su propia
// transacción de una sola base, porque PostgreSQL no ofrece 2PC usable entre
// dos instancias separadas— con checkpoints en arco_requests.status y un job
// de reconciliación en internal/maintenance para trámites que quedaban a
// medio camino si el proceso moría entre fases.
//
// Con una sola base (F5, plan/04_Rediseno_identidad_gustos.md) ese problema
// desaparece: CancelAccount es una única *sql.Tx. Se lleva con ella toda la
// maquinaria de checkpoint/reconciliación — ya no hay "a medio camino"
// posible, un rollback de Postgres deshace la cancelación completa o no
// deshace nada.
//
// Además, con el rediseño, la cancelación es autoservicio inmediato
// (DELETE /auth/me, decisión 7 del rediseño) — ya no depende de que un admin
// apruebe una solicitud ARCO de tipo cancelación primero. (Útil)
package privacy

import (
	"context"
	cryptorand "crypto/rand"
	"database/sql"
	"fmt"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

// CancelAccountParams identifica la cuenta a cancelar y el motivo a
// registrar. No lleva ningún secreto de cifrado: sin email no hay nada que
// ofuscar con una clave (§1 del rediseño). (Útil)
type CancelAccountParams struct {
	AccountID uuid.UUID
	Reason    string
}

// CancelAccount ejecuta, en una sola transacción, todo lo que antes eran las
// fases [P] e [I] de la saga:
//   - purga el progreso no-ledger (player_progress, level_attempts,
//     daily_streak, user_badges) y las respuestas del cuestionario;
//   - pone a NULL al usuario en las bitácoras append-only
//     (experience_history, audit_log), preservando la evidencia de
//     No-Repudio;
//   - marca los dispositivos del usuario para wipe local y revoca todos sus
//     refresh tokens;
//   - sobrescribe accounts.nickname con relleno aleatorio, marca
//     status='deleted' y fuerza token_version+1 (dentro de DeactivateAccount).
//
// La fila de accounts NUNCA se borra — se conserva seudonimizada, para que
// las bitácoras que la referencian sigan teniendo integridad referencial. (Útil)
func CancelAccount(ctx context.Context, repo *repository.Queries, p CancelAccountParams) error {
	randomNickname, err := randomNicknameFill()
	if err != nil {
		return fmt.Errorf("generating random nickname fill: %w", err)
	}

	tx, err := repo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("beginning cancellation tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	qtx := repo.WithTx(tx)

	if err := qtx.NullUserInPseudonymizableLedgers(ctx, p.AccountID); err != nil {
		return fmt.Errorf("pseudonymizing ledgers: %w", err)
	}
	if err := qtx.PurgeUserProgressData(ctx, p.AccountID); err != nil {
		return fmt.Errorf("purging progress data: %w", err)
	}
	if err := qtx.PurgeAccountQuizAnswers(ctx, p.AccountID); err != nil {
		return fmt.Errorf("purging quiz answers: %w", err)
	}
	if err := qtx.MarkUserDevicesForWipe(ctx, p.AccountID); err != nil {
		return fmt.Errorf("marking devices for wipe: %w", err)
	}
	if err := qtx.RevokeRefreshTokensForUser(ctx, p.AccountID); err != nil {
		return fmt.Errorf("revoking refresh tokens: %w", err)
	}
	if err := qtx.DeactivateAccount(ctx, repository.DeactivateAccountParams{
		ID:             p.AccountID,
		RandomNickname: randomNickname,
		DeletionReason: p.Reason,
	}); err != nil {
		return fmt.Errorf("deactivating account: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing cancellation: %w", err)
	}
	return nil
}

const nicknameFillAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
const nicknameFillLength = 20

// randomNicknameFill genera el relleno que sobrescribe accounts.nickname al
// cancelar (§1.1 punto 2 del rediseño). Usa crypto/rand, no math/rand: a
// diferencia del nickname candidato que ve la persona usuaria en el
// registro, este valor nunca se muestra ni se recuerda — solo tiene que
// cumplir el CHECK y no colisionar, y no hay razón para no usar el generador
// criptográfico por defecto del proyecto. (Útil)
func randomNicknameFill() (string, error) {
	b := make([]byte, nicknameFillLength)
	if _, err := cryptorand.Read(b); err != nil {
		return "", err
	}
	for i, v := range b {
		b[i] = nicknameFillAlphabet[int(v)%len(nicknameFillAlphabet)]
	}
	return string(b), nil
}
