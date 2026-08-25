// account_queries.go es NUEVO en USBI-Anon — no existe en ../usbi, donde no
// hacía falta una tabla ancla porque las FK apuntaban directo a `users`. Ver
// plan/00_Plan_maestro.md §2 y plan/02_Backend.md §3.3/§4.
package repository

import (
	"context"
	cryptorand "crypto/rand"
	"math/big"
	"time"

	"github.com/google/uuid"
)

// aliasVocabularySize es el número de palabras sembradas en alias_adjectives
// y alias_nouns (ids 1..24, ver migrations/main/0001_esquema_principal.up.sql).
// Vive aquí, no en una consulta, porque el rol de aplicación no tiene permiso
// de escritura sobre esos vocabularios (backend/sql/00_roles_principal.sql):
// Go sortea índices dentro de rangos conocidos, nunca los lee de la base.
const aliasVocabularySize = 24

// RandomAlias sortea un id de adjetivo, un id de sustantivo y un número para
// el alias visible de una cuenta nueva. Se llama en CADA Login/Refresh, no
// solo en el alta: UpsertAccount descarta estos valores vía ON CONFLICT DO
// UPDATE cuando la cuenta ya existe, así que generarlos de más es más barato
// que consultar primero si la fila existe.
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

type Account struct {
	ID               uuid.UUID
	Role             string
	Status           string
	IsAdult          bool
	AliasAdjectiveID int16
	AliasNounID      int16
	AliasNumber      int16
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UpsertAccountParams struct {
	ID               uuid.UUID
	Role             string
	Status           string
	IsAdult          bool
	AliasAdjectiveID int16
	AliasNounID      int16
	AliasNumber      int16
}

// UpsertAccount da de alta o refresca la réplica no autoritativa de
// `identities` en la base principal. Se llama SOLO desde Login/Refresh (ver
// plan/02_Backend.md §3.5, decisión de rendimiento: una sola consulta extra
// por login/refresh, ninguna en el resto de requests autenticados).
//
// El ON CONFLICT DO UPDATE deliberadamente NO toca las tres columnas de
// alias: el alias se sortea una sola vez, en el alta, y debe permanecer
// estable para la persona usuaria (plan/02_Backend.md §3.3).
func (q *Queries) UpsertAccount(ctx context.Context, arg UpsertAccountParams) error {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO accounts (
    id, role, status, is_adult, alias_adjective_id, alias_noun_id, alias_number
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (id) DO UPDATE SET
    role       = EXCLUDED.role,
    status     = EXCLUDED.status,
    is_adult   = EXCLUDED.is_adult,
    updated_at = NOW()
`, arg.ID, arg.Role, arg.Status, arg.IsAdult, arg.AliasAdjectiveID, arg.AliasNounID, arg.AliasNumber)
	return err
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
