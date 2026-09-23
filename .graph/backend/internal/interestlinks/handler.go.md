---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define el controlador HTTP que expone las rutas para consultar y administrar los enlaces de interés. Restringe el acceso a los endpoints administrativos únicamente a usuarios autenticados con rol `RoleAdmin` y traduce las excepciones de [[backend/internal/interestlinks/admin_service.go.md|admin_service.go]] y de [[backend/internal/interestlinks/service.go.md|service.go]] (`ErrValidation`, `ErrNotFound`, `ErrCategoryHasLinks`) a respuestas de error estructuradas bajo la especificación RFC 7807 Problem Details.

## Funciones

### canManageInterestLinks
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Comprueba si el rol del usuario extraído del contexto posee permisos de administración (`RoleAdmin`).

### Handler.ListInterestLinks
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende `GET /interest-links` delegando en [[backend/internal/interestlinks/player_service.go.md#PlayerService.ListGrouped|PlayerService.ListGrouped]], respondiendo con el carrusel de categorías y enlaces agrupados para cualquier usuario autenticado sin importar su rol.

### Handler.ListCategories
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida la autorización de administración y recupera el listado completo de categorías de enlaces de interés.

### Handler.CreateCategory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica el rol de administrador, decodifica estrictamente el cuerpo JSON y delega la creación de la categoría a [[backend/internal/interestlinks/admin_service.go.md#AdminService.CreateCategory|AdminService.CreateCategory]].

### Handler.UpdateCategory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida permisos administrativos, extrae el UUID de la ruta, decodifica la solicitud y ejecuta la actualización de la categoría.

### Handler.DeleteCategory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Comprueba privilegios de administrador, parsea el UUID de la categoría y solicita su eliminación, respondiendo con estado 204 No Content.

### Handler.ListLinks
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica el rol de administrador y consulta la lista global de todos los enlaces de interés registrados.

### Handler.CreateLink
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida permisos de administrador, decodifica los datos del cuerpo JSON y procesa la creación del nuevo enlace de interés.

### Handler.UpdateLink
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida autorización administrativa, obtiene el `link_id` de la URL, decodifica la carga JSON y ejecuta la actualización del enlace.

### Handler.DeleteLink
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Comprueba privilegios administrativos, parsea el identificador UUID del enlace y delega su eliminación retornando 204 No Content.

### parseURLUUID
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Extrae y valida un parámetro UUID desde la URL mediante Chi, escribiendo una respuesta 400 Bad Request si el formato es inválido.

### writeServiceError
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Mapea los errores de dominio (`ErrValidation`, `ErrNotFound`, `ErrCategoryHasLinks`) a sus correspondientes códigos HTTP (422, 404, 409, 500) y respuestas Problem Details.
