// Package identityrepo agrupa la migración de F2 (ver
// plan/00_Plan_maestro.md y plan/02_Backend.md §3.2).
//
// En ../usbi, 6 de estas consultas (CreateUser, GetUserByEmailHash,
// GetUserTokenVersion, IncrementTokenVersion, UpdateUserAdultStatus,
// IncrementAgeUpAttempts, InsertArcoRequest) las generaba sqlc a partir de
// sql/query.sql + sqlc.yaml. Aquí se escriben a mano, siguiendo el mismo
// patrón (ctx, Params) -> (Result, error) que repository/doc.go documenta y
// valida como equivalente: no se reprodujo el pipeline sqlc para 7 consultas
// que no van a cambiar de forma con el tiempo, y porque más de la mitad de
// este paquete (auth, tutor consent, privacy, maintenance) ya era
// hand-written en ../usbi. Si en el futuro crece el número de consultas
// generables, reevaluar sqlc aquí — no antes.
package identityrepo
