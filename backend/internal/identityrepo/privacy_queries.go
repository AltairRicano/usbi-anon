// Migrado desde ../usbi/backend/internal/repository/privacy_queries.go (todo
// menos NullUserInPseudonymizableLedgers y PurgeUserProgressData, que se
// quedan en [P] internal/repository — ver plan/02_Backend.md §3.3) más
// InsertArcoRequest, que en ../usbi vivía en query.sql.go (sqlc). Las cuatro
// consultas de arco_requests quedan reunidas aquí porque las cuatro hablan
// con la BASE DE IDENTIDAD, que es donde vive arco_requests en USBI-Anon.
//
// PseudonymizeUser pierde las líneas `full_name = ...` y
// `phone = NULL, phone_lookup_hash = NULL`: esas columnas no existen en
// `identities`. El resto de la seudonimización (correo, estado, revocación,
// token_version) es idéntico a ../usbi.
package identityrepo

import (
	"context"
	"net"
	"time"

	"github.com/google/uuid"
)

type InsertArcoRequestParams struct {
	ID            uuid.UUID
	UserID        uuid.NullUUID
	RequesterType string
	RequestType   string
	Status        string
	EvidenceHash  []byte
}

func (q *Queries) InsertArcoRequest(ctx context.Context, arg InsertArcoRequestParams) error {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO arco_requests (
    id, user_id, requester_type, request_type, status, evidence_hash
) VALUES (
    $1, $2, $3, $4, $5, $6
)
`, arg.ID, arg.UserID, arg.RequesterType, arg.RequestType, arg.Status, arg.EvidenceHash)
	return err
}

type InsertTutorConsentParams struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	TutorName            string
	TutorEmail           string
	PrivacyNoticeVersion string
	AcceptedAt           time.Time
	AcceptanceIP         net.IP
	AcceptanceUserAgent  string
	ConsentSignature     []byte
	CryptoKeyVersion     int16
	EncryptionKey        string
}

func (q *Queries) InsertTutorConsent(ctx context.Context, arg InsertTutorConsentParams) error {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO tutor_consents (
    id, user_id, tutor_name, tutor_email, privacy_notice_version, accepted_at,
    acceptance_ip, acceptance_user_agent, consent_signature, crypto_key_version
) VALUES (
    $1, $2,
    pgp_sym_encrypt($3::text, $11::text),
    pgp_sym_encrypt($4::text, $11::text),
    $5, $6, $7, $8, $9, $10
)
`, arg.ID, arg.UserID, arg.TutorName, arg.TutorEmail, arg.PrivacyNoticeVersion, arg.AcceptedAt,
		arg.AcceptanceIP, arg.AcceptanceUserAgent, arg.ConsentSignature, arg.CryptoKeyVersion, arg.EncryptionKey)
	return err
}

func (q *Queries) ActivateTutorConsentUser(ctx context.Context, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE identities
SET status = 'active', updated_at = NOW()
WHERE id = $1 AND status = 'pending_tutor_consent' AND deleted_at IS NULL
`, userID)
	return err
}

type ArcoRequestForResolution struct {
	ID          uuid.UUID
	UserID      uuid.NullUUID
	RequestType string
	Status      string
}

type ArcoPendingRequest struct {
	ID            uuid.UUID
	RequesterType string
	RequestType   string
	Status        string
	ReceivedAt    time.Time
}

func (q *Queries) ListPendingArcoRequests(ctx context.Context, limit int32) ([]ArcoPendingRequest, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT id, requester_type, request_type, status, received_at
FROM arco_requests
WHERE status = 'pending'
ORDER BY received_at ASC
LIMIT $1
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ArcoPendingRequest, 0)
	for rows.Next() {
		var item ArcoPendingRequest
		if err := rows.Scan(&item.ID, &item.RequesterType, &item.RequestType, &item.Status, &item.ReceivedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (q *Queries) GetArcoRequestForUpdate(ctx context.Context, id uuid.UUID) (ArcoRequestForResolution, error) {
	var req ArcoRequestForResolution
	err := q.db.QueryRowContext(ctx, `
SELECT id, user_id, request_type, status
FROM arco_requests
WHERE id = $1
FOR UPDATE
`, id).Scan(&req.ID, &req.UserID, &req.RequestType, &req.Status)
	return req, err
}

// UpdateArcoRequestStatus persiste un checkpoint de la saga de cancelación
// (plan/02_Backend.md §5): pending → purging_main → identity_pseudonymized →
// resolved/rejected. Deliberadamente NO toca resolved_at/handled_by/
// response_summary — eso lo hace ResolveArcoRequest solo en el paso final,
// para no fijar esos campos antes de que el trámite realmente concluya.
func (q *Queries) UpdateArcoRequestStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE arco_requests SET status = $2 WHERE id = $1
`, id, status)
	return err
}

// ClaimArcoRequestForCancellation transiciona pending → purging_main Y
// persiste handled_by/response_summary EN ESE MISMO UPDATE, no en el paso
// final de la saga. Es lo que le permite al job de reconciliación (o a un
// segundo llamador que estaba bloqueado detrás del FOR UPDATE de
// GetArcoRequestForUpdate) terminar un trámite atascado sin depender de
// información que solo existía en la petición HTTP original del admin, que
// para entonces ya pudo haber terminado o el proceso haber muerto.
func (q *Queries) ClaimArcoRequestForCancellation(ctx context.Context, id uuid.UUID, handledBy uuid.NullUUID, responseSummary string) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE arco_requests
SET status = 'purging_main', handled_by = $2, response_summary = $3
WHERE id = $1 AND status = 'pending'
`, id, handledBy, responseSummary)
	return err
}

// MarkArcoRequestResolved cierra el último paso de una cancelación aprobada
// cuyo handled_by/response_summary ya se guardaron en ClaimArcoRequestForCancellation.
// Deliberadamente separado de ResolveArcoRequest (que sí recibe esos dos
// campos): esta versión la puede llamar el job de reconciliación, que no
// tiene ni el actor ni el resumen de la solicitud original, solo el
// checkpoint ya persistido.
func (q *Queries) MarkArcoRequestResolved(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE arco_requests SET status = 'resolved', resolved_at = NOW() WHERE id = $1
`, id)
	return err
}

// ListStuckArcoRequests alimenta el job de reconciliación (regla 5 de
// plan/02_Backend.md §5): solicitudes que quedaron a medio camino porque el
// proceso murió entre el paso [P] y el paso [I] de la saga. olderThan evita
// que el job se pise con una resolución que apenas está en curso en otro
// request del mismo proceso.
func (q *Queries) ListStuckArcoRequests(ctx context.Context, olderThan time.Time, limit int32) ([]ArcoRequestForResolution, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT id, user_id, request_type, status
FROM arco_requests
WHERE status IN ('purging_main', 'identity_pseudonymized')
  AND received_at < $1
ORDER BY received_at ASC
LIMIT $2
`, olderThan, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ArcoRequestForResolution, 0)
	for rows.Next() {
		var item ArcoRequestForResolution
		if err := rows.Scan(&item.ID, &item.UserID, &item.RequestType, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type ResolveArcoRequestParams struct {
	ID              uuid.UUID
	HandledBy       uuid.NullUUID
	Status          string
	ResponseSummary string
}

func (q *Queries) ResolveArcoRequest(ctx context.Context, arg ResolveArcoRequestParams) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE arco_requests
SET status = $2,
    resolved_at = NOW(),
    handled_by = $3,
    response_summary = $4
WHERE id = $1
`, arg.ID, arg.Status, arg.HandledBy, arg.ResponseSummary)
	return err
}

type PseudonymizeUserParams struct {
	UserID          uuid.UUID
	PseudonymEmail  string
	EmailLookupHash []byte
	EncryptionKey   string
	DeletionReason  string
}

func (q *Queries) PseudonymizeUser(ctx context.Context, arg PseudonymizeUserParams) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE identities
SET email = pgp_sym_encrypt($2::text, $4::text),
    email_lookup_hash = $3,
    status = 'deleted',
    deleted_at = COALESCE(deleted_at, NOW()),
    deletion_reason = $5,
    token_version = token_version + 1,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
`, arg.UserID, arg.PseudonymEmail, arg.EmailLookupHash, arg.EncryptionKey, arg.DeletionReason)
	return err
}

type PseudonymizeTutorConsentsParams struct {
	UserID        uuid.UUID
	EncryptionKey string
}

func (q *Queries) PseudonymizeTutorConsents(ctx context.Context, arg PseudonymizeTutorConsentsParams) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE tutor_consents
SET tutor_name = pgp_sym_encrypt('Tutor pseudonimizado'::text, $2::text),
    tutor_email = pgp_sym_encrypt('tutor-pseudonimizado@example.invalid'::text, $2::text),
    revoked_at = COALESCE(revoked_at, NOW())
WHERE user_id = $1
`, arg.UserID, arg.EncryptionKey)
	return err
}
