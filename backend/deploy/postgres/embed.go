// Package pgconf empotra las plantillas de configuración de Postgres
// (M3.7 del plan de maduración): overrides de postgresql.conf y una
// plantilla parametrizada de pg_hba.conf, nunca una copia literal de los
// archivos completos — eso los ataría a una versión concreta de Postgres.
package pgconf

import "embed"

//go:embed *.tmpl
var FS embed.FS
