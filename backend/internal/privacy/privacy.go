// Package privacy contiene la cancelación definitiva y autoservicio de una cuenta.
// CancelAccount se ejecuta en una única transacción serializable (*sql.Tx), garantizando atomicidad
// completa sin estados intermedios.
package privacy

import (
	"context"
	cryptorand "crypto/rand"
	"database/sql"
	"fmt"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

// CancelAccountParams identifica la cuenta a cancelar y el motivo a registrar.
type CancelAccountParams struct {
	AccountID uuid.UUID
	Reason    string
}

// CancelAccount ejecuta, en una sola transacción serializable:
//   - purga el progreso no-ledger (player_progress, level_attempts,
//     daily_streak, user_badges) y las respuestas del cuestionario;
//   - pone a NULL al usuario en las bitácoras append-only
//     (experience_history, audit_log), preservando la evidencia de No-Repudio;
//   - marca los dispositivos del usuario para wipe local y revoca todos sus refresh tokens;
//   - sobrescribe accounts.nickname con relleno aleatorio, marca status='deleted'
//     y fuerza token_version+1 (dentro de DeactivateAccount).
//
// La fila de accounts NUNCA se borra — se conserva seudonimizada, para que
// las bitácoras que la referencian sigan teniendo integridad referencial.
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

// randomNicknameFill genera el relleno que sobrescribe accounts.nickname al cancelar.
// Usa crypto/rand: este valor nunca se muestra ni se recuerda — solo tiene que
// cumplir el CHECK de longitud/formato y no colisionar.
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
