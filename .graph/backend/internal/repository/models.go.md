---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Contiene las estructuras de datos Go (`Device`, `Level`, `Section`) generadas por sqlc que representan tablas principales del sistema. Define los campos de cada entidad, incluyendo la adaptación de `Device` para utilizar una enumeración de tipos de dispositivo (`device_kind`) en lugar de texto libre. `Device` es consumida por [[backend/internal/devices/service.go.md|devices/service.go]] y por [[backend/internal/sync/service.go.md#Service.ProcessSync|sync/service.go#Service.ProcessSync]] (campo `WipeLocalData` de la respuesta de sincronización); `Level` y `Section` son la base de [[backend/internal/levels/admin_service.go.md|levels/admin_service.go]] y [[backend/internal/levels/player_service.go.md|levels/player_service.go]].
