---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Contiene los métodos generados por sqlc para operaciones sobre el historial de experiencia, intentos de niveles, registros de auditoría unificados (`audit_log`), estados de sincronización y rachas diarias. `InsertExperienceHistory`, `InsertLevelAttempt`, `UpdateSyncEventStatus` y `UpsertDailyStreak` los invoca [[backend/internal/sync/service.go.md#Service.ProcessSync|sync/service.go#Service.ProcessSync]] en el camino offline, y [[backend/internal/levels/player_service.go.md|levels/player_service.go]] en el camino online. `LogAuditEntry` es la base sobre la que [[backend/internal/audit/audit.go.md|audit/audit.go]] construye su helper genérico de auditoría, usado por prácticamente todo `admin_service.go` del backend.

## Funciones

### Queries.InsertExperienceHistory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Registra una entrada en la tabla de historial de experiencia asociando usuario, nivel, evento, XP obtenida y método de verificación.

### Queries.InsertLevelAttempt
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Guarda un registro correspondiente al intento de un nivel por parte de un usuario con sus métricas obtenidas.

### Queries.LogAuditEntry
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inserta un registro en la bitácora unificada de auditoría `audit_log`, capturando la cuenta del actor, acción, entidad y estados previo y posterior.

### Queries.UpdateSyncEventStatus
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza el estado de procesamiento de un evento de sincronización registrando la fecha y hora actual.

### Queries.UpsertDailyStreak
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inserta un registro de actividad diaria de un usuario o lo ignora si ya existe una entrada para la misma fecha.
