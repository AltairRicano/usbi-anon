---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Página administrativa para inspeccionar la bitácora de auditoría y gestionar incidentes de seguridad. Preserva la privacidad sin exponer nicknames para actor_account_id y alerta visualmente si un incidente posee evidencia alterada (evidence_valid es falso).

## Funciones

### AdminSecurityPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente React principal para la administración de bitácoras de auditoría e incidentes de seguridad.

### AdminSecurityPage.loadAuditLog
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Consulta entradas de auditoría con filtros opcionales de acción y tipo de entidad, utilizando paginación por cursor vía GET /admin/audit-log.

### AdminSecurityPage.loadIncidents
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la lista paginada de incidentes de seguridad mediante GET /admin/security-incidents.

### AdminSecurityPage.openIncidentsTab
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Cambia la pestaña activa a incidentes y solicita su carga inicial si no se han cargado previamente.

### AdminSecurityPage.createIncident
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida la severidad y registra un nuevo incidente de seguridad con su alcance y contención mediante POST /admin/security-incidents.

### AdminSecurityPage.startEdit
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Copia los datos de un incidente al estado de edición local para permitir su modificación.

### AdminSecurityPage.saveEdit
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza el registro de un incidente de seguridad mediante PATCH /admin/security-incidents/{id}.

## Relaciones

- [[frontend/src/features/admin-security/schemas.ts.md|schemas.ts]] — valida bitácora e incidentes con esquemas Zod
