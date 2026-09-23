---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Ofrece operaciones para el registro, consulta y modificación de la bitácora de incidentes de seguridad. Opera bajo el pool de conexiones de `usbi_moderador`, previene borrados de registros mediante triggers a nivel de esquema y mantiene inmutable el campo `detected_at`. Consumida por [[backend/internal/incidents/service.go.md|incidents/service.go]] (creación y actualización) y [[backend/internal/incidents/admin_read.go.md|incidents/admin_read.go]] (listado y lectura individual).

## Funciones

### Queries.InsertSecurityIncident
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inserta un nuevo incidente de seguridad con su alcance, nivel de severidad, descripción, acciones de contención y hash de evidencia.

### Queries.ListSecurityIncidents
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Recupera una lista paginada de incidentes de seguridad ordenados de forma descendente por fecha de detección y por su identificador.

### Queries.GetSecurityIncident
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene un incidente de seguridad específico mediante su identificador único.

### Queries.UpdateSecurityIncident
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza la narrativa, campos de seguimiento y hash de evidencia de un incidente existente, manteniendo inalterada la fecha de detección.
