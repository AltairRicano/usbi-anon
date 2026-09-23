---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define las estructuras de datos de transferencia (DTOs) empleadas en las peticiones y respuestas HTTP del módulo de dispositivos. Permite restringir el vocabulario de tipos de dispositivo y soporta el envío de un ID opcional para actualizaciones automáticas de sesión sin permitir la suplantación o apropiación de dispositivos entre usuarios. `deviceToResponse` en [[backend/internal/devices/service.go.md|service.go]] traduce el modelo `repository.Device` (definido en [[backend/internal/repository/device_queries.go.md|device_queries.go]]) a `DeviceResponse`.
