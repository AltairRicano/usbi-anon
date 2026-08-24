package identityrepo

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Identity es la fila completa de `identities` — la tabla que reemplaza a
// `users` de ../usbi. Sin FullName ni Phone/PhoneLookupHash: esa PII se
// eliminó por diseño (ver migrations/identity/0001_esquema_identidad.up.sql).
type Identity struct {
	ID                      uuid.UUID
	Email                   []byte // cifrado con pgcrypto, nunca en claro
	EmailLookupHash         []byte
	PasswordHash            string
	TokenVersion            int32
	IsAdult                 bool
	Role                    string
	PrivacyNoticeVersion    string
	PrivacyNoticeAcceptedAt time.Time
	PrivacyAcceptanceHash   []byte
	CryptoKeyVersion        int16
	Status                  string
	AgeUpAttempts           int16
	CreatedAt               time.Time
	UpdatedAt               time.Time
	LastLoginAt             sql.NullTime
	DeletedAt               sql.NullTime
	DeletionReason          sql.NullString
}

type CreateIdentityParams struct {
	ID                      uuid.UUID
	EmailPlaintext          string // se cifra en el propio INSERT con pgp_sym_encrypt
	EmailLookupHash         []byte
	PasswordHash            string
	TokenVersion            int32
	IsAdult                 bool
	Role                    string
	PrivacyNoticeVersion    string
	PrivacyNoticeAcceptedAt time.Time
	PrivacyAcceptanceHash   []byte
	CryptoKeyVersion        int16
	Status                  string
	EncryptionKey           string
}

// CreateIdentity reemplaza a repository.CreateUser. Migrada desde
// ../usbi/backend/internal/repository/query.sql.go (generada por sqlc); aquí
// se escribe a mano porque este paquete no reproduce el pipeline sqlc — ver
// nota en doc.go.
func (q *Queries) CreateIdentity(ctx context.Context, arg CreateIdentityParams) (Identity, error) {
	row := q.db.QueryRowContext(ctx, `
INSERT INTO identities (
    id, email, email_lookup_hash, password_hash, token_version, is_adult, role,
    privacy_notice_version, privacy_notice_accepted_at, privacy_acceptance_hash,
    crypto_key_version, status
) VALUES (
    $1, pgp_sym_encrypt($2::text, $13::text), $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING id, email, email_lookup_hash, password_hash, token_version, is_adult, role,
    privacy_notice_version, privacy_notice_accepted_at, privacy_acceptance_hash,
    crypto_key_version, status, age_up_attempts, created_at, updated_at, last_login_at,
    deleted_at, deletion_reason
`,
		arg.ID,
		arg.EmailPlaintext,
		arg.EmailLookupHash,
		arg.PasswordHash,
		arg.TokenVersion,
		arg.IsAdult,
		arg.Role,
		arg.PrivacyNoticeVersion,
		arg.PrivacyNoticeAcceptedAt,
		arg.PrivacyAcceptanceHash,
		arg.CryptoKeyVersion,
		arg.Status,
		arg.EncryptionKey,
	)
	var i Identity
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.EmailLookupHash,
		&i.PasswordHash,
		&i.TokenVersion,
		&i.IsAdult,
		&i.Role,
		&i.PrivacyNoticeVersion,
		&i.PrivacyNoticeAcceptedAt,
		&i.PrivacyAcceptanceHash,
		&i.CryptoKeyVersion,
		&i.Status,
		&i.AgeUpAttempts,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.LastLoginAt,
		&i.DeletedAt,
		&i.DeletionReason,
	)
	return i, err
}

type GetIdentityByEmailHashParams struct {
	EmailLookupHash []byte
	EncryptionKey   string
}

// GetIdentityByEmailHashRow es como Identity pero con el correo YA
// DESCIFRADO (pgp_sym_decrypt en el propio SELECT) — mismo motivo que en
// ../usbi para que el login pueda comparar/mostrar el correo en claro sin un
// segundo roundtrip.
type GetIdentityByEmailHashRow struct {
	ID                      uuid.UUID
	Email                   string
	EmailLookupHash         []byte
	PasswordHash            string
	TokenVersion            int32
	IsAdult                 bool
	Role                    string
	PrivacyNoticeVersion    string
	PrivacyNoticeAcceptedAt time.Time
	PrivacyAcceptanceHash   []byte
	CryptoKeyVersion        int16
	Status                  string
	AgeUpAttempts           int16
	CreatedAt               time.Time
	UpdatedAt               time.Time
	LastLoginAt             sql.NullTime
	DeletedAt               sql.NullTime
	DeletionReason          sql.NullString
}

// GetIdentityByEmailHash reemplaza a repository.GetUserByEmailHash.
func (q *Queries) GetIdentityByEmailHash(ctx context.Context, arg GetIdentityByEmailHashParams) (GetIdentityByEmailHashRow, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT id, pgp_sym_decrypt(email::bytea, $2::text) as email, email_lookup_hash,
    password_hash, token_version, is_adult, role, privacy_notice_version,
    privacy_notice_accepted_at, privacy_acceptance_hash, crypto_key_version,
    status, age_up_attempts, created_at, updated_at, last_login_at, deleted_at,
    deletion_reason
FROM identities
WHERE email_lookup_hash = $1 AND deleted_at IS NULL
`, arg.EmailLookupHash, arg.EncryptionKey)
	var i GetIdentityByEmailHashRow
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.EmailLookupHash,
		&i.PasswordHash,
		&i.TokenVersion,
		&i.IsAdult,
		&i.Role,
		&i.PrivacyNoticeVersion,
		&i.PrivacyNoticeAcceptedAt,
		&i.PrivacyAcceptanceHash,
		&i.CryptoKeyVersion,
		&i.Status,
		&i.AgeUpAttempts,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.LastLoginAt,
		&i.DeletedAt,
		&i.DeletionReason,
	)
	return i, err
}

// GetUserTokenVersion, IncrementTokenVersion, UpdateUserAdultStatus e
// IncrementAgeUpAttempts NO se renombran — plan/02_Backend.md §3.2 solo marca
// CreateUser y GetUserByEmailHash como renombradas. Diff mínimo: solo cambia
// `users` → `identities` en el cuerpo SQL, igual que ya se hizo con la
// columna `user_id` en el esquema.

func (q *Queries) GetUserTokenVersion(ctx context.Context, id uuid.UUID) (int32, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT token_version FROM identities WHERE id = $1 AND deleted_at IS NULL
`, id)
	var tokenVersion int32
	err := row.Scan(&tokenVersion)
	return tokenVersion, err
}

func (q *Queries) IncrementTokenVersion(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE identities SET token_version = token_version + 1, updated_at = NOW() WHERE id = $1
`, id)
	return err
}

func (q *Queries) UpdateUserAdultStatus(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE identities SET is_adult = true, status = 'active', updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL
`, id)
	return err
}

func (q *Queries) IncrementAgeUpAttempts(ctx context.Context, id uuid.UUID) (int16, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE identities SET age_up_attempts = age_up_attempts + 1, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL RETURNING age_up_attempts
`, id)
	var ageUpAttempts int16
	err := row.Scan(&ageUpAttempts)
	return ageUpAttempts, err
}
