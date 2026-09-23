---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa la consulta y paginación del historial de eventos de sincronización offline para el jugador autenticado. Se ejecuta sobre el pool de base de datos del usuario (`usbi_app`) asegurando aislamiento por `user_id`, y omite los datos JSONB pesados para retornar estructuras ligeras paginadas por cursor. Es un método de `Service`, el mismo tipo definido en [[backend/internal/sync/service.go.md|service.go]]; invocado por [[backend/internal/sync/handler.go.md#Handler.ListMyHistory|handler.go#Handler.ListMyHistory]].

## Funciones

### Service.ListMySyncEvents
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Consulta el historial de eventos de sincronización del usuario autenticado de forma paginada por cursor (`received_at`) mediante [[backend/internal/repository/content_queries.go.md#Queries.ListSyncEventsForUser|Queries.ListSyncEventsForUser]]. Valida que el `user_id` esté presente, restringe el tamaño de página entre 1 y 50 (por defecto 20), consulta el repositorio agregando un margen para verificar si hay más registros y calcula la marca de tiempo `NextCursor` en formato RFC3339Nano.
