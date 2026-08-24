// IsNoRows vive duplicada aquí y en identityrepo por la misma razón que DBTX
// (ver identityrepo/db.go): es un helper genérico sin acoplamiento a ninguna
// tabla, y internal/sync (paquete [P], copiado verbatim en F3) la invoca como
// repository.IsNoRows — igual que en ../usbi/backend/internal/repository,
// donde vivía junto a las consultas de auth antes de que F2 migrara ese
// archivo entero a identityrepo.
package repository

import "database/sql"

func IsNoRows(err error) bool {
	return err == sql.ErrNoRows
}
