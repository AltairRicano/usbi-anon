---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define los DTOs (`EntryResponse`, `Page`, `Filters`) utilizados en la consulta paginada del registro de auditoría administrativo. Garantiza la seguridad e independencia de esquemas manteniendo `actor_account_id` como UUID crudo sin resolver nicknames en la tabla de cuentas, y define la estructura de cursores opacos basados en marcas de tiempo RFC3339Nano y UUIDs. `toResponse` traduce cada fila `repository.AuditLogEntry` (modelo devuelto por [[backend/internal/repository/content_queries.go.md#Queries.ListAuditLog|Queries.ListAuditLog]]) al DTO público; [[backend/internal/auditlog/service.go.md#AdminService.List|AdminService.List]] es el único consumidor de estos tipos.
