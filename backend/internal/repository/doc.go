// Package repository es la capa de acceso a datos de la ÚNICA base del
// sistema (usbi_anon_db): identidad (accounts, refresh_tokens), progreso,
// contenido, dispositivos, incidentes y audit_log.
//
// internal/identityrepo existió entre F2 y F5 para la base de identidad
// separada que el rediseño descartó (ver
// plan/04_Rediseno_identidad_gustos.md); F7 absorbió aquí lo reutilizable de
// ese paquete (CreateAccount/FindAccountByNickname en account_queries.go,
// refresh tokens en auth_queries.go, DeactivateAccount/
// PurgeAccountQuizAnswers en privacy_queries.go) y lo demás no volvió: el
// flujo de tutor por correo se eliminó completo, y las consultas de ARCO las
// añade F9 junto con internal/auth, que es quien las usa.
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
// ListPublishedLevels, LogAuditEntry (entonces LogAdminAudit),
// UpdateSyncEventStatus, UpsertDailyStreak, UpsertPlayerProgress) y los 12
// tipos de fila de progreso/contenido/auditoría de la base principal
// (AuditLog — entonces AdminAuditLog, renombrada por F7 al fusionarse con
// identity_audit_log —, Badge,
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
// PurgeUserProgressData) migraron íntegros a identityrepo en F2. ESTA
// RESTRICCIÓN YA NO APLICA: con una sola base (F5) e identityrepo eliminado
// todo se vuelve a juntar aquí en internal/repository. (Útil)esh_tokens a este paquete (auth_queries.go) y
// accounts ya vivía aquí desde F1 (account_queries.go). `tutor_consent*` no
// vuelve — el flujo de tutor se eliminó por completo, no se fusionó.
// `arco_requests` tampoco vuelve todavía: la trae F9 junto con
// internal/auth, que es su único consumidor.
package repository
