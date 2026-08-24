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
