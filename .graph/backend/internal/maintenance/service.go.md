---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa las tareas periódicas de retención legal y purga de datos del sistema. Es responsable de suspender automáticamente cuentas inactivas tras un período configurable, procesar en lotes la cancelación definitiva e idempotente de cuentas suspendidas superado el límite de retención, y purgar tokens de refresco expirados o revocados.

## Funciones

### NewService
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicializa el servicio de mantenimiento asignando valores por defecto a la configuración si no son especificados (365 días para suspensión por inactividad, 30 días para cancelación y lotes de 100 registros).

### Service.RunOnce
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Ejecuta un ciclo completo de mantenimiento: suspende cuentas inactivas con [[backend/internal/repository/maintenance_queries.go.md#Queries.SuspendInactivePlayers|Queries.SuspendInactivePlayers]], obtiene con [[backend/internal/repository/maintenance_queries.go.md#Queries.ListSuspendedUsersForCancellation|Queries.ListSuspendedUsersForCancellation]] y cancela en lotes las cuentas cuyo periodo de suspensión expiró, y purga los refresh tokens obsoletos con [[backend/internal/repository/auth_queries.go.md#Queries.PurgeExpiredRefreshTokens|Queries.PurgeExpiredRefreshTokens]], retornando el resumen de la operación.

### Service.cancelAccount
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Delega la cancelación de la cuenta a [[backend/internal/privacy/privacy.go.md#CancelAccount|privacy.CancelAccount]] con el motivo de expiración por inactividad, garantizando la misma lógica de borrado seguro e idempotencia que el flujo autoservicio (`DELETE /api/v1/auth/me`).

### StartScheduler
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicia un temporizador en un goroutine en segundo plano que invoca periódicamente a `RunOnce` según el intervalo configurado (por defecto 24 horas) y registra métricas o advertencias en los logs.
