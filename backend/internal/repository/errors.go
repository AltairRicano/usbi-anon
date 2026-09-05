// IsNoRows es un helper genérico sin acoplamiento a ninguna tabla.
// internal/sync (paquete [P], copiado verbatim en F3) la invoca como
// repository.IsNoRows — igual que en ../usbi/backend/internal/repository,
// donde vivía junto a las consultas de auth. Entre F2 y F6 existió una copia
// duplicada en identityrepo por la misma razón que DBTX (ver el db.go que
// tenía ese paquete); con identityrepo eliminado (F6) ya no hay nada que
// duplicar. (Útil)
package repository

import "database/sql"

func IsNoRows(err error) bool {
	return err == sql.ErrNoRows
}
