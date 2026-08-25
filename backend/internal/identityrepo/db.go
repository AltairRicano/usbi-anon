// Package identityrepo is the data-access layer para la BASE DE IDENTIDAD
// (usbi_ident_db): identities, refresh_tokens, tutor_consents,
// tutor_consent_tokens, arco_requests, identity_audit_log.
//
// Es la contraparte de internal/repository, que sigue existiendo para la BASE
// PRINCIPAL. Ningún archivo de este paquete debe nombrar una tabla que viva en
// esa otra base (progreso, contenido o su bitácora de auditoría) — ver
// plan/02_Backend.md §8, criterio 3.
//
// DBTX reutiliza exactamente el mismo contrato que repository.DBTX (no hace
// falta reinventarlo: ambos paquetes reciben un *sql.DB o una *sql.Tx de su
// propia base). No hay un tipo compartido entre los dos paquetes a propósito:
// mezclar un *sql.Tx de una base con Queries de la otra sería un error de
// compilación imposible, no solo un error en tiempo de ejecución.
package identityrepo

import (
	"context"
	"database/sql"
	"errors"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

type Queries struct {
	db DBTX
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db: tx,
	}
}

// BeginTx/txBeginner/ErrTransactionsUnsupported son el mismo patrón que
// repository.BeginTx (ver internal/repository/content_queries.go, heredado
// de ../usbi): no existían aquí en F2 porque ninguna consulta de F2 abría su
// propia transacción. F4 los necesita porque AgeUp, el doble opt-in de tutor
// y la fase de identidad de la saga ARCO (internal/privacy) sí la abren.
var ErrTransactionsUnsupported = errors.New("identityrepo: configured DBTX does not support transactions")

type txBeginner interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

func (q *Queries) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	db, ok := q.db.(txBeginner)
	if !ok {
		return nil, ErrTransactionsUnsupported
	}
	return db.BeginTx(ctx, opts)
}
