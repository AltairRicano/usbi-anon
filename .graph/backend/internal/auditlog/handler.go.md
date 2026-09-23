---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Expone los puntos de entrada HTTP para la lectura del registro de auditoría en el panel de administración (`GET /api/v1/admin/audit-log`). Implementa control de acceso estricto reservado exclusivamente para administradores y parsea y valida los parámetros de consulta y paginación antes de delegar a [[backend/internal/auditlog/service.go.md#AdminService.List|AdminService.List]].

## Funciones

### Handler.List
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende solicitudes `GET /api/v1/admin/audit-log`. Valida el rol de administrador en los claims JWT (retornando 403 Forbidden mediante [[backend/internal/httpproblem/httpproblem.go.md#WriteProblem|WriteProblem]] si no es administrador), parsea y valida los parámetros de consulta opcionales (UUID del actor, timestamps RFC3339 de rango, acción, tipo de entidad, cursor y tamaño de página entre 1 y 50), delega en [[backend/internal/auditlog/service.go.md#AdminService.List|AdminService.List]] y retorna la respuesta paginada con [[backend/internal/httpproblem/httpproblem.go.md#WriteJSON|WriteJSON]].
