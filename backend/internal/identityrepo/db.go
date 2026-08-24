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
