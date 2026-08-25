// Package privacy contiene la saga de cancelación definitiva compartida por
// AMBOS caminos que la disparan: una ARCO cancelación aprobada por un
// administrador (internal/auth.Service.ResolveArcoRequest) y la retención
// legal automática (internal/maintenance.Service.cancelUser). En ../usbi
// existía la misma regla (audit finding A4: los dos caminos no deben poder
// divergir) implementada como una única *sql.Tx.
//
// Con dos bases separadas esa transacción única deja de ser posible
// (PostgreSQL no ofrece 2PC usable sobre un contenedor de ~256 MB — ver
// plan/02_Backend.md §5). El reemplazo es una saga de dos fases, cada una en
// su propia transacción de UNA sola base:
//
//	PurgeMain            [P] progreso, bitácoras, dispositivos
//	PseudonymizeIdentity [I] identidad, tutores, sesión
//
// La BASE PRINCIPAL VA PRIMERO a propósito (regla 1 de plan/02_Backend.md
// §5): si el proceso muere entre las dos fases, queda una cuenta VIVA con el
// progreso ya purgado — recuperable y reintentable — en vez de una cuenta ya
// bloqueada (identidad seudonimizada) que no podría volver a autenticarse
// para completar su propia cancelación.
//
// Ambas fases son idempotentes por construcción (regla 2): reintentar
// cualquiera de las dos no hace daño. Eso es lo que permite que el llamador
// persista un checkpoint entre fases (arco_requests.status en el caso de
// ARCO; un reintento natural en el siguiente RunOnce en el caso de la
// retención automática, que no tiene una fila de solicitud que actualizar) y
// reanude desde ahí sin re-ejecutar la fase ya confirmada — aunque hacerlo
// tampoco sería incorrecto.
package privacy

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/altair/usbi-anon-backend/internal/identityrepo"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

// CancelParams carga los identificadores y secretos que necesita la fase de
// identidad. Ya no lleva FullName/Phone: esas columnas no existen en
// `identities`.
type CancelParams struct {
	UserID           uuid.UUID
	Reason           string
	EncryptionKey    string
	BlindIndexSecret []byte
}

// PurgeMain ejecuta la fase [P] de la saga: purga el progreso no-ledger
// (player_progress, level_attempts, daily_streak, user_badges), pone a NULL
// al usuario en las bitácoras append-only (experience_history,
// admin_audit_log) preservando la evidencia de No-Repudio, y marca los
// dispositivos del usuario para wipe local. Reutiliza exactamente las
// mismas tres consultas [P] que ya existían (dos desde F3, más
// MarkUserDevicesForWipe, ya presente en device_queries.go desde ../usbi —
// no se dio de alta una función redundante).
//
// Idempotente: las tres consultas son DELETE/UPDATE ... WHERE user_id = $1
// sin precondición de estado, así que repetir la fase tras un reintento no
// falla ni corrompe nada.
func PurgeMain(ctx context.Context, main *repository.Queries, userID uuid.UUID) error {
	tx, err := main.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("beginning main-db tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	qtx := main.WithTx(tx)

	if err := qtx.NullUserInPseudonymizableLedgers(ctx, userID); err != nil {
		return fmt.Errorf("pseudonymizing ledgers: %w", err)
	}
	if err := qtx.PurgeUserProgressData(ctx, userID); err != nil {
		return fmt.Errorf("purging progress data: %w", err)
	}
	if err := qtx.MarkUserDevicesForWipe(ctx, userID); err != nil {
		return fmt.Errorf("marking devices for wipe: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing main-db purge: %w", err)
	}
	return nil
}

// PseudonymizeIdentity ejecuta la fase [I] de la saga: seudonimiza al
// usuario y sus consentimientos de tutor, y revoca todos sus refresh tokens
// (token_version + 1 incluido, dentro de PseudonymizeUser). Conserva
// `identities.id` — la seudonimización nunca borra la fila ni el UUID, así
// que la fase [P] siempre puede reintentarse con la misma llave si hiciera
// falta (regla 4 de plan/02_Backend.md §5).
//
// Idempotente vía el guard `AND deleted_at IS NULL` de PseudonymizeUser:
// reintentar tras ya haber pseudonimizado es un no-op, no un error.
func PseudonymizeIdentity(ctx context.Context, ident *identityrepo.Queries, p CancelParams) error {
	pseudonymEmail := "deleted-" + p.UserID.String() + "@pseudonymized.usbi.invalid"
	emailHash := crypto.BlindIndexHMAC(pseudonymEmail, p.BlindIndexSecret)

	tx, err := ident.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("beginning identity-db tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	qtx := ident.WithTx(tx)

	if err := qtx.PseudonymizeTutorConsents(ctx, identityrepo.PseudonymizeTutorConsentsParams{
		UserID:        p.UserID,
		EncryptionKey: p.EncryptionKey,
	}); err != nil {
		return fmt.Errorf("pseudonymizing tutor consents: %w", err)
	}
	if err := qtx.PseudonymizeUser(ctx, identityrepo.PseudonymizeUserParams{
		UserID:          p.UserID,
		PseudonymEmail:  pseudonymEmail,
		EmailLookupHash: emailHash,
		EncryptionKey:   p.EncryptionKey,
		DeletionReason:  p.Reason,
	}); err != nil {
		return fmt.Errorf("pseudonymizing identity: %w", err)
	}
	if err := qtx.RevokeRefreshTokensForUser(ctx, p.UserID); err != nil {
		return fmt.Errorf("revoking refresh tokens: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing identity-db pseudonymization: %w", err)
	}
	return nil
}

// ResumeArcoCancellation termina una solicitud ARCO de cancelación aprobada
// que quedó a medio camino, a partir de su checkpoint (status). La llaman
// DOS sitios: internal/auth.Service.ResolveArcoRequest, justo después de
// reclamar el trámite bajo lock (puede encontrarlo ya en purging_main o
// identity_pseudonymized si un intento previo murió a mitad de camino), y el
// job de reconciliación de internal/maintenance, que rebarre trámites
// atascados encontrados por identityrepo.ListStuckArcoRequests.
//
// HandledBy/ResponseSummary NO se tocan aquí: ya se persistieron en el
// momento del reclamo (identityrepo.ClaimArcoRequestForCancellation), a
// propósito, para que esta función pueda completarse sin depender de datos
// que solo existían en la petición HTTP original — el job de reconciliación
// no tiene ni actor ni resumen a mano, solo el checkpoint ya guardado.
func ResumeArcoCancellation(ctx context.Context, ident *identityrepo.Queries, main *repository.Queries, requestID uuid.UUID, checkpoint string, p CancelParams) error {
	if checkpoint == "purging_main" {
		if err := PurgeMain(ctx, main, p.UserID); err != nil {
			return fmt.Errorf("purging main data: %w", err)
		}
		if err := PseudonymizeIdentity(ctx, ident, p); err != nil {
			return fmt.Errorf("pseudonymizing identity: %w", err)
		}
		if err := ident.UpdateArcoRequestStatus(ctx, requestID, "identity_pseudonymized"); err != nil {
			return fmt.Errorf("checkpointing identity_pseudonymized: %w", err)
		}
	}
	if err := ident.MarkArcoRequestResolved(ctx, requestID); err != nil {
		return fmt.Errorf("marking arco request resolved: %w", err)
	}
	return nil
}
