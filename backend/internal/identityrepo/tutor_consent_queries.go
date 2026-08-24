// Migrado íntegro desde
// ../usbi/backend/internal/repository/tutor_consent_queries.go. Único cambio:
// GetUserStatusByID consulta `identities` en vez de `users`. Las cinco
// consultas del doble opt-in (tutor_consent_tokens) no cambian: esa tabla ya
// vivía fuera de `users` en ../usbi y su esquema es idéntico aquí.
package identityrepo

import (
	"context"
	"database/sql"
	"net"
	"time"

	"github.com/google/uuid"
)

// GetUserStatusByID returns the lifecycle status of a non-deleted identity.
// It is used by the tutor-consent flow to decide whether to issue a
// verification token without leaking (via differing responses) whether an
// account exists.
func (q *Queries) GetUserStatusByID(ctx context.Context, id uuid.UUID) (string, error) {
	var status string
	err := q.db.QueryRowContext(ctx, `
SELECT status FROM identities WHERE id = $1 AND deleted_at IS NULL
`, id).Scan(&status)
	return status, err
}

type InsertTutorConsentTokenParams struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	TutorName            string
	TutorEmail           string
	PrivacyNoticeVersion string
	TokenHash            []byte
	RequestedIP          net.IP
	RequestedUserAgent   string
	CryptoKeyVersion     int16
	ExpiresAt            time.Time
	EncryptionKey        string
}

// InsertTutorConsentToken stores a pending tutor-consent request. The tutor's
// name/email are encrypted with pgcrypto; only an HMAC of the raw token is kept.
func (q *Queries) InsertTutorConsentToken(ctx context.Context, arg InsertTutorConsentTokenParams) error {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO tutor_consent_tokens (
    id, user_id, tutor_name, tutor_email, privacy_notice_version, token_hash,
    requested_ip, requested_user_agent, crypto_key_version, created_at, expires_at
) VALUES (
    $1, $2,
    pgp_sym_encrypt($3::text, $11::text),
    pgp_sym_encrypt($4::text, $11::text),
    $5, $6, $7, $8, $9, NOW(), $10
)
`, arg.ID, arg.UserID, arg.TutorName, arg.TutorEmail, arg.PrivacyNoticeVersion,
		arg.TokenHash, arg.RequestedIP, arg.RequestedUserAgent, arg.CryptoKeyVersion,
		arg.ExpiresAt, arg.EncryptionKey)
	return err
}

// DeleteUnverifiedTutorConsentTokensForUser removes any still-pending tokens for
// the user so a re-submission (resend) supersedes previous, un-clicked links.
func (q *Queries) DeleteUnverifiedTutorConsentTokensForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
DELETE FROM tutor_consent_tokens WHERE user_id = $1 AND verified_at IS NULL
`, userID)
	return err
}

type TutorConsentTokenRow struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	TutorName            string // decrypted
	TutorEmail           string // decrypted
	PrivacyNoticeVersion string
	ExpiresAt            time.Time
	VerifiedAt           sql.NullTime
	CryptoKeyVersion     int16
}

// GetTutorConsentTokenByHashForUpdate fetches and row-locks a token by its HMAC.
// FOR UPDATE serialises concurrent clicks so an account can't be activated twice.
// Must be called within a transaction.
func (q *Queries) GetTutorConsentTokenByHashForUpdate(ctx context.Context, tokenHash []byte, encryptionKey string) (TutorConsentTokenRow, error) {
	var row TutorConsentTokenRow
	err := q.db.QueryRowContext(ctx, `
SELECT id, user_id,
       pgp_sym_decrypt(tutor_name, $2::text),
       pgp_sym_decrypt(tutor_email, $2::text),
       privacy_notice_version, expires_at, verified_at, crypto_key_version
FROM tutor_consent_tokens
WHERE token_hash = $1
FOR UPDATE
`, tokenHash, encryptionKey).Scan(
		&row.ID, &row.UserID, &row.TutorName, &row.TutorEmail,
		&row.PrivacyNoticeVersion, &row.ExpiresAt, &row.VerifiedAt, &row.CryptoKeyVersion,
	)
	return row, err
}

type MarkTutorConsentTokenVerifiedParams struct {
	ID                    uuid.UUID
	VerifiedAt            time.Time
	VerificationIP        net.IP
	VerificationUserAgent string
}

// MarkTutorConsentTokenVerified records the tutor's click as legal evidence.
func (q *Queries) MarkTutorConsentTokenVerified(ctx context.Context, arg MarkTutorConsentTokenVerifiedParams) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE tutor_consent_tokens
SET verified_at = $2, verification_ip = $3, verification_user_agent = $4
WHERE id = $1
`, arg.ID, arg.VerifiedAt, arg.VerificationIP, arg.VerificationUserAgent)
	return err
}

// PurgeExpiredTutorConsentTokens deletes unverified tokens past their expiry.
// Returns the number of rows removed. Intended for the legal-maintenance job.
func (q *Queries) PurgeExpiredTutorConsentTokens(ctx context.Context) (int64, error) {
	res, err := q.db.ExecContext(ctx, `
DELETE FROM tutor_consent_tokens WHERE verified_at IS NULL AND expires_at < NOW()
`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
