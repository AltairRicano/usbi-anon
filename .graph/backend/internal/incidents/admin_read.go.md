---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona la lectura paginada, consulta individual y edición de incidentes de seguridad para administradores. La seguridad e inmutabilidad se garantizan al recalcular dinámicamente el HMAC de evidencia en cada lectura para validar integridad, prohibiendo la eliminación física en BD mediante un trigger BEFORE DELETE y manteniendo un historial append-only de versiones previas en audit_log.

## Funciones

### Service.toResponse
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Convierte el registro de base de datos a IncidentResponse, recalculando el hash HMAC con [[backend/internal/crypto/hash.go.md#VerifyHMAC|VerifyHMAC]] sobre el payload para determinar y adjuntar la validez criptográfica de la evidencia en cada lectura.

### Service.List
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la lista paginada de incidentes con [[backend/internal/repository/security_incident_queries.go.md#Queries.ListSecurityIncidents|Queries.ListSecurityIncidents]], ordenados cronológicamente descendentemente mediante un cursor opaco con formato fecha-UUID, restringiendo el acceso exclusivamente al rol de administrador.

### Service.Get
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Recupera un incidente de seguridad por su identificador único con [[backend/internal/repository/security_incident_queries.go.md#Queries.GetSecurityIncident|Queries.GetSecurityIncident]] para usuarios con rol administrador, retornando la estructura con el estado de verificación HMAC recalculado.

### Service.Update
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza la narrativa y columnas de seguimiento de un incidente dentro de una transacción SQL con [[backend/internal/repository/security_incident_queries.go.md#Queries.UpdateSecurityIncident|Queries.UpdateSecurityIncident]], re-sellando el HMAC de evidencia con [[backend/internal/crypto/hash.go.md#GenerateHMAC|GenerateHMAC]], impidiendo modificar la fecha de detección y registrando la versión previa completa en `audit_log` mediante [[backend/internal/audit/audit.go.md#Log|audit.Log]].
