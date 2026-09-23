---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Expone los puntos de entrada HTTP para la gestión administrativa de insignias, delegando en [[backend/internal/badges/admin_service.go.md|AdminService]]. Se encarga de verificar que el usuario tenga rol de administrador (`RoleAdmin`), decodificar peticiones JSON de forma estricta con [[backend/internal/httpjson/decode.go.md#DecodeStrict|DecodeStrict]], validar parámetros de ruta UUID y traducir los errores del servicio en respuestas Problem Details (RFC 7807) con [[backend/internal/httpproblem/httpproblem.go.md#WriteProblem|WriteProblem]].

## Funciones

### Handler.List
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica el rol de administrador en el contexto de la petición y retorna el listado general de insignias delegando en [[backend/internal/badges/admin_service.go.md#AdminService.List|AdminService.List]].

### Handler.Create
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida permisos de administrador, decodifica el cuerpo JSON estrictamente y delega la creación de la insignia en [[backend/internal/badges/admin_service.go.md#AdminService.Create|AdminService.Create]], respondiendo con el código 201 Created.

### Handler.Update
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida permisos de administración, extrae el UUID de la insignia desde los parámetros de la URL, decodifica el cuerpo JSON y delega la actualización en [[backend/internal/badges/admin_service.go.md#AdminService.Update|AdminService.Update]].

### Handler.Delete
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida el rol de administrador, obtiene el UUID del parámetro de ruta y delega el borrado en [[backend/internal/badges/admin_service.go.md#AdminService.Delete|AdminService.Delete]], retornando 204 No Content o un conflicto HTTP si la insignia tiene titulares.
