---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Gestiona el registro, estado de actividad y la revocación de dispositivos móviles o clientes conectados (`devices`). Controla la presencia del dispositivo (`last_seen_at`), permite marcar la orden de borrado de datos locales (`wipe_local_data`) y asegura que los dispositivos revocados no puedan realizar acciones de sincronización. [[backend/internal/devices/service.go.md|devices/service.go]] es el consumidor CRUD principal; [[backend/internal/sync/service.go.md#Service.ProcessSync|sync/service.go#Service.ProcessSync]] llama `GetActiveDevice` y `TouchDevice` en cada sincronización, y [[backend/internal/privacy/privacy.go.md|privacy/privacy.go]] usa `MarkUserDevicesForWipe` al cancelar una cuenta.

## Funciones

### Queries.CreateDevice
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Registra un nuevo dispositivo asociado al usuario especificando tipo de dispositivo y plataforma.

### Queries.GetActiveDevice
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la información de un dispositivo activo validando que pertenezca al usuario y no esté revocado.

### Queries.TouchDevice
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza la fecha de última actividad (`last_seen_at`) de un dispositivo no revocado.

### Queries.TouchDeviceReturning
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza la fecha de actividad y retorna el estado del dispositivo en una sola consulta atómica.

### Queries.MarkUserDevicesForWipe
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Establece la bandera `wipe_local_data` en verdadero para todos los dispositivos activos del usuario.

### Queries.RevokeDevice
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Revoca un dispositivo especificando la fecha actual e instruye el borrado de datos locales en la siguiente sincronización.

### Queries.ListDevices
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Lista todos los dispositivos activos no revocados de un usuario ordenados por su última fecha de conexión.
