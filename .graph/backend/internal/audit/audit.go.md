---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Centraliza la persistencia de registros de auditoría en la tabla audit_log para respaldar el no-repudio de operaciones sensibles. Asigna valores por defecto ("0.0.0.0" y "backend-service") para peticiones internas sin contexto HTTP, maneja UUIDs opcionales y serializa estados previos y posteriores a JSON.

## Funciones

### Log
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Registra una entrada en la tabla de auditoría a través del repositorio especificado, llamando a [[backend/internal/repository/query.sql.go.md#Queries.LogAuditEntry|Queries.LogAuditEntry]]. Configura valores predeterminados para IP y User-Agent si están vacíos, maneja la opcionalidad de UUIDs para actor/entidad y delega la serialización del estado previo y posterior. Es la puerta de escritura que usan servicios de todo el backend (p. ej. [[backend/internal/auditlog/service.go.md#AdminService.List|AdminService.List]] para auditar su propia lectura) cada vez que necesitan dejar evidencia de una operación sensible.

### marshalState
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Convierte un valor de tipo any en un objeto pqtype.NullRawMessage en formato JSON; si el valor recibido es nil, retorna una estructura válida conteniendo un objeto JSON vacío ("{}").
