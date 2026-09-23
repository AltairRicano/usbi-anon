---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa la lógica de negocio y las reglas de seguridad para la administración de dispositivos, invocada desde handler.go. Valida el vocabulario de tipos de dispositivo y plataformas contra un catálogo permitido antes de interactuar con la base de datos, gestiona el flujo de upsert seguro para inicios de sesión recurrentes y ejecuta la revocación lógica transaccional junto con el registro auditado de acciones con [[backend/internal/audit/audit.go.md#Log|audit.Log]].

## Funciones

### Service.RegisterDevice
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida que el tipo de dispositivo pertenezca al conjunto permitido (movil, tablet, laptop, escritorio, otro) y que la plataforma sea web o tauri. Si se proporciona un ID de dispositivo existente, intenta actualizar la fecha de último acceso con [[backend/internal/repository/device_queries.go.md#Queries.TouchDeviceReturning|Queries.TouchDeviceReturning]]; si el ID no existe, pertenece a otro usuario o está revocado, crea un nuevo registro con [[backend/internal/repository/device_queries.go.md#Queries.CreateDevice|Queries.CreateDevice]].

### Service.RevokeDevice
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Realiza la revocación lógica de un dispositivo con [[backend/internal/repository/device_queries.go.md#Queries.RevokeDevice|Queries.RevokeDevice]] para conservar la integridad referencial de la clave foránea en eventos de sincronización. Ejecuta una transacción SQL que establece la fecha de revocación y registra un evento de auditoría ("device.revoke") con [[backend/internal/audit/audit.go.md#Log|audit.Log]] antes de confirmar los cambios.

### Service.ListDevices
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica la validez del identificador del usuario y obtiene con [[backend/internal/repository/device_queries.go.md#Queries.ListDevices|Queries.ListDevices]] la lista completa de sus dispositivos asociados (tanto activos como revocados), mapeándolos a DTOs de respuesta definidos en [[backend/internal/devices/dto.go.md|dto.go]].
