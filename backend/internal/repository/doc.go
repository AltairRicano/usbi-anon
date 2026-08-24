// Package repository es la capa de acceso a datos de la BASE PRINCIPAL
// (usbi_anon_db): progreso, contenido, dispositivos, incidentes y
// admin_audit_log. Es la contraparte de internal/identityrepo, que cubre la
// BASE DE IDENTIDAD — ver plan/02_Backend.md §3.
//
// # Origen: partición de ../usbi/backend/internal/repository (F3)
//
// En ../usbi, este paquete mezclaba consultas de las 17 tablas sobre una
// sola base. content_queries.go, badge_queries.go, device_queries.go,
// security_incident_queries.go y db.go no tocaban `users` — verificado con
// grep — así que se copiaron aquí verbatim, sin editar una línea.
//
// query.sql.go y models.go sí mezclaban ambas bases (los generaba sqlc a
// partir de sql/query.sql, que cubre 18 consultas: 7 de identidad y 11 de
// progreso/contenido). Aquí solo sobreviven los 11 fragmentos [P]
// (CreateLevel, CreateSection, GetLevelAttemptsByDate,
// InsertExperienceHistory, InsertLevelAttempt, InsertSyncEvent,
// ListPublishedLevels, LogAdminAudit, UpdateSyncEventStatus,
// UpsertDailyStreak, UpsertPlayerProgress) y los 12 tipos de fila de
// progreso/contenido/auditoría de la base principal (AdminAuditLog, Badge,
// DailyStreak, Device, ExperienceHistory, Level, LevelAttempt,
// PlayerProgress, Section, SecurityIncident, SyncEvent, UserBadge) —
// byte-idénticos a como sqlc los generó en ../usbi, sin regenerar el
// pipeline. Los 7 fragmentos de identidad (CreateUser, GetUserByEmailHash,
// GetUserTokenVersion, IncrementTokenVersion, UpdateUserAdultStatus,
// IncrementAgeUpAttempts, InsertArcoRequest) y los tipos User, ArcoRequest,
// TutorConsent migraron a identityrepo en F2.
//
// De los 11 fragmentos [P], 4 (GetLevelAttemptsByDate, InsertSyncEvent,
// ListPublishedLevels, UpsertPlayerProgress) ya estaban muertos en ../usbi
// —superseded por CountLevelAttemptsByDate, InsertSyncEventWithPayload y
// UpsertPlayerProgressForAttempt en content_queries.go, documentado en el
// doc.go original—. Se conservan igual, por fidelidad con el recorte "la
// parte de query.sql.go/models.go referida a progreso y contenido" que pide
// plan/02_Backend.md §2.1, no por uso real.
//
// auth_queries.go, tutor_consent_queries.go, maintenance_queries.go y
// privacy_queries.go (salvo NullUserInPseudonymizableLedgers y
// PurgeUserProgressData, que sí operan sobre la base principal y siguen
// aquí) migraron íntegros a identityrepo en F2 — ningún archivo de este
// paquete debe volver a mencionar `identities`, `refresh_tokens`,
// `tutor_consent*` ni `arco_requests` (criterio 2 de plan/02_Backend.md §8).
package repository
