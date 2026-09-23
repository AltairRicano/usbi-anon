---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona las consultas de retención legal automática e inactividad de jugadores. Permite la suspensión de jugadores inactivos invalidando sus sesiones activas al incrementar la versión de sus tokens y lista cuentas aptas para su posterior cancelación definitiva. Único consumidor: [[backend/internal/maintenance/service.go.md|maintenance/service.go]], que ejecuta ambas consultas dentro de su tarea periódica de retención.

## Funciones

### Queries.SuspendInactivePlayers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Suspende cuentas de jugadores sin actividad previa a una fecha límite, incrementa `token_version` para revocar sesiones y marca el motivo de eliminación como pendiente de cancelación.

### Queries.ListSuspendedUsersForCancellation
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene los identificadores de cuentas de jugadores suspendidos por inactividad cuyo periodo de gracia expiró antes de la fecha límite especificada.
