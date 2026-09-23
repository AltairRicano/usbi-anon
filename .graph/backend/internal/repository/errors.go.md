---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona la función auxiliar de comprobación de errores del repositorio, centralizando la verificación de ausencia de registros en la base de datos sin acoplar los paquetes superiores a los tipos internos de `database/sql`. [[backend/internal/sync/service.go.md#Service.ProcessSync|sync/service.go#Service.ProcessSync]] la invoca como `repository.IsNoRows` al validar que el dispositivo exista antes de procesar un evento de sincronización.
