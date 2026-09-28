// account_queries.go es NUEVO en USBI-Anon — no existe en ../usbi, donde no
// hacía falta una tabla ancla porque las FK apuntaban directo a `users`. Ver
// plan/00_Plan_maestro.md §2.
//
// F7 (rediseño de identidad, plan/04_Rediseno_identidad_gustos.md §2)
// absorbió aquí CreateAccount/FindAccountByNickname, que antes vivían
// repartidas entre identityrepo (CreateIdentity/GetIdentityByEmailHash) y
// UpsertAccount de este mismo archivo. Con una sola base ya no existe la
// réplica no autoritativa que UpsertAccount mantenía sincronizada en cada
// login/refresh — accounts es ahora la única fila, se crea una vez en el
// registro y se lee directo por nickname en el login.
package repository

import (
	"context"
	cryptorand "crypto/rand"
	"database/sql"
	"math/big"
	"time"

	"github.com/google/uuid"
)

// aliasVocabularySize es el número de palabras sembradas en alias_adjectives
// y alias_nouns (ids 1..24, ver migrations/0001_esquema_unificado.up.sql).
// Vive aquí, no en una consulta, porque el rol de aplicación no tiene permiso
// de escritura sobre esos vocabularios (backend/sql/00_roles_unificado.sql):
// Go sortea índices dentro de rangos conocidos, nunca los lee de la base.
const aliasVocabularySize = 24

// RandomAlias sortea un id de adjetivo, un id de sustantivo y un número para
// el alias visible de una cuenta nueva. Se llama UNA SOLA VEZ, en el INSERT
// de registro (CreateAccount) — con una sola base ya no hay una réplica que
// reconciliar en cada login/refresh, así que generar el alias de más ya no
// tiene sentido (plan/04_Rediseno_identidad_gustos.md §1).
func RandomAlias() (adjectiveID, nounID, number int16, err error) {
	adj, err := cryptorand.Int(cryptorand.Reader, big.NewInt(aliasVocabularySize))
	if err != nil {
		return 0, 0, 0, err
	}
	noun, err := cryptorand.Int(cryptorand.Reader, big.NewInt(aliasVocabularySize))
	if err != nil {
		return 0, 0, 0, err
	}
	num, err := cryptorand.Int(cryptorand.Reader, big.NewInt(1000))
	if err != nil {
		return 0, 0, 0, err
	}
	return int16(adj.Int64()) + 1, int16(noun.Int64()) + 1, int16(num.Int64()), nil
}

// Account es la fila completa de `accounts` — id, credencial de login y
// metadatos de cuenta. Es un tipo interno de acceso a datos, no el DTO
// público (ese es domain.User): incluye PasswordHash y el resto de columnas
// sensibles porque solo lo consumen internal/auth e internal/privacy, nunca
// se serializa directo a un response HTTP.
type Account struct {
	ID                      uuid.UUID
	Nickname                string
	PasswordHash            string
	TokenVersion            int32
	IsAdult                 bool
	Role                    string
	Status                  string
	AgeUpAttempts           int16
	AliasAdjectiveID        int16
	AliasNounID             int16
	AliasNumber             int16
	PrivacyNoticeVersion    string
	PrivacyNoticeAcceptedAt time.Time
	PrivacyAcceptanceHash   []byte
	CryptoKeyVersion        int16
	CreatedAt               time.Time
	UpdatedAt               time.Time
	LastLoginAt             sql.NullTime
	DeletedAt               sql.NullTime
	DeletionReason          sql.NullString
}

const accountColumns = `
    id, nickname, password_hash, token_version, is_adult, role, status,
    age_up_attempts, alias_adjective_id, alias_noun_id, alias_number,
    privacy_notice_version, privacy_notice_accepted_at, privacy_acceptance_hash,
    crypto_key_version, created_at, updated_at, last_login_at, deleted_at,
    deletion_reason`

func scanAccount(row scanner) (Account, error) {
	var a Account
	err := row.Scan(
		&a.ID, &a.Nickname, &a.PasswordHash, &a.TokenVersion, &a.IsAdult, &a.Role, &a.Status,
		&a.AgeUpAttempts, &a.AliasAdjectiveID, &a.AliasNounID, &a.AliasNumber,
		&a.PrivacyNoticeVersion, &a.PrivacyNoticeAcceptedAt, &a.PrivacyAcceptanceHash,
		&a.CryptoKeyVersion, &a.CreatedAt, &a.UpdatedAt, &a.LastLoginAt, &a.DeletedAt,
		&a.DeletionReason,
	)
	return a, err
}

type CreateAccountParams struct {
	ID                      uuid.UUID
	Nickname                string
	PasswordHash            string
	IsAdult                 bool
	Role                    string
	AliasAdjectiveID        int16
	AliasNounID             int16
	AliasNumber             int16
	PrivacyNoticeVersion    string
	PrivacyNoticeAcceptedAt time.Time
	PrivacyAcceptanceHash   []byte
	CryptoKeyVersion        int16
}

// CreateAccount da de alta la cuenta. Es la ÚNICA inserción de la fila en
// toda la vida de la cuenta: status nace 'active' siempre (ya no hay
// 'pending_tutor_consent' — un menor autoreportado juega de inmediato) y
// token_version nace en 1 vía DEFAULT del esquema.
func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	row := q.db.QueryRowContext(ctx, `
INSERT INTO accounts (
    id, nickname, password_hash, is_adult, role, status,
    alias_adjective_id, alias_noun_id, alias_number,
    privacy_notice_version, privacy_notice_accepted_at, privacy_acceptance_hash,
    crypto_key_version
) VALUES (
    $1, $2, $3, $4, $5, 'active', $6, $7, $8, $9, $10, $11, $12
)
RETURNING `+accountColumns,
		arg.ID, arg.Nickname, arg.PasswordHash, arg.IsAdult, arg.Role,
		arg.AliasAdjectiveID, arg.AliasNounID, arg.AliasNumber,
		arg.PrivacyNoticeVersion, arg.PrivacyNoticeAcceptedAt, arg.PrivacyAcceptanceHash,
		arg.CryptoKeyVersion,
	)
	return scanAccount(row)
}

// FindAccountByNickname es la consulta del login: busca la credencial
// directo por nickname (sin blind index — el nickname no es PII cifrada, ver
// plan/04_Rediseno_identidad_gustos.md §1) y excluye cuentas ya canceladas.
func (q *Queries) FindAccountByNickname(ctx context.Context, nickname string) (Account, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT `+accountColumns+`
FROM accounts
WHERE nickname = $1 AND deleted_at IS NULL
`, nickname)
	return scanAccount(row)
}

// GetAccountByID lee la cuenta por id — la usa el middleware de auth en cada
// request autenticado para revalidar token_version/status/role contra la
// base, no solo contra lo que dice el JWT.
func (q *Queries) GetAccountByID(ctx context.Context, id uuid.UUID) (Account, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT `+accountColumns+`
FROM accounts
WHERE id = $1 AND deleted_at IS NULL
`, id)
	return scanAccount(row)
}

// GetAccountAlias lee la vista account_aliases, que compone "Jaguar Azul 42"
// sin que ninguna tabla almacene la cadena. El alias no es único: no debe
// usarse jamás como clave de búsqueda, solo como saludo de UI.
func (q *Queries) GetAccountAlias(ctx context.Context, userID uuid.UUID) (string, error) {
	var alias string
	err := q.db.QueryRowContext(ctx, `
SELECT display_alias FROM account_aliases WHERE id = $1
`, userID).Scan(&alias)
	return alias, err
}

// TouchAccountLastLogin registra el instante de un login/refresh exitoso.
// Nunca existió en el diseño de dos bases (identityrepo no la tenía) — sin
// ella, accounts_inactive_players_idx (COALESCE(last_login_at, created_at))
// mide inactividad desde la fecha de alta para siempre, sin importar cuánto
// juegue la persona. F9 la añade porque es la primera fase que de verdad
// implementa Login/Refresh contra esta tabla.
func (q *Queries) TouchAccountLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE accounts SET last_login_at = NOW() WHERE id = $1
`, id)
	return err
}

// IncrementAccountTokenVersion invalida todos los JWT vivos de la cuenta —
// logout y reseteo de password la usan igual.
func (q *Queries) IncrementAccountTokenVersion(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE accounts SET token_version = token_version + 1, updated_at = NOW() WHERE id = $1
`, id)
	return err
}

// IncrementAgeUpAttempts es el contador de intentos de transición a mayoría
// de edad (Ley 251, máx. 3) — internal/auth.AgeUp lo revisa antes de aplicar
// el cambio.
func (q *Queries) IncrementAgeUpAttempts(ctx context.Context, id uuid.UUID) (int16, error) {
	var attempts int16
	err := q.db.QueryRowContext(ctx, `
UPDATE accounts SET age_up_attempts = age_up_attempts + 1, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING age_up_attempts
`, id).Scan(&attempts)
	return attempts, err
}

// MarkAccountAdult aplica la transición a mayoría de edad. Sin flujo de
// tutor que activar (a diferencia del diseño de dos bases): solo cambia
// is_adult, el status ya era 'active' desde el registro.
func (q *Queries) MarkAccountAdult(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE accounts SET is_adult = true, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
`, id)
	return err
}

// SetAccountPassword reemplaza el hash y fuerza token_version+1 en la misma
// UPDATE — un reseteo de password (autoservicio o por un admin) debe
// invalidar cualquier sesión que sobreviva con el password anterior.
func (q *Queries) SetAccountPassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE accounts
SET password_hash = $2, token_version = token_version + 1, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
`, id, passwordHash)
	return err
}

type UpdatePrivacyAcceptanceParams struct {
	ID         uuid.UUID
	Version    string
	AcceptedAt time.Time
	Hash       []byte
}

// UpdatePrivacyAcceptance registra que la cuenta aceptó una versión nueva del
// aviso de privacidad mientras ya tenía sesión — el banner informativo de
// cambio de versión (M2.5, D-06), no el registro inicial (ese sello lo pone
// CreateAccount).
func (q *Queries) UpdatePrivacyAcceptance(ctx context.Context, arg UpdatePrivacyAcceptanceParams) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE accounts
SET privacy_notice_version = $2, privacy_notice_accepted_at = $3, privacy_acceptance_hash = $4, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
`, arg.ID, arg.Version, arg.AcceptedAt, arg.Hash)
	return err
}
