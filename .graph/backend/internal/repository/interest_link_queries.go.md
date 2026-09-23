---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona acceso a datos para categorías de enlaces de interés, enlaces y sugerencias. Implementa un guard previo al borrado de categorías para evitar violaciones de clave foránea y gestiona sugerencias anónimas omitiendo `RETURNING` en consultas para no requerir privilegios `SELECT` bajo el rol `usbi_app`. Las consultas de enlaces las consumen [[backend/internal/interestlinks/player_service.go.md|interestlinks/player_service.go]] (lectura pública) y [[backend/internal/interestlinks/admin_service.go.md|interestlinks/admin_service.go]] (CRUD); las de sugerencias las consumen [[backend/internal/suggestions/player_service.go.md#PlayerService.Submit|suggestions/player_service.go#PlayerService.Submit]] (`CreateSuggestion`) y [[backend/internal/suggestions/admin_service.go.md|suggestions/admin_service.go]] (`ListSuggestions`, `DeleteSuggestion`).

## Funciones

### Queries.ListInterestLinkCategories
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene las categorías de enlaces de interés ordenadas por preferencia de despliegue y fecha de creación para el carrusel de la interfaz.

### Queries.CreateInterestLinkCategory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inserta una nueva categoría de enlaces de interés con su identificador, nombre y orden visual.

### Queries.UpdateInterestLinkCategory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza el nombre y orden de despliegue de una categoría de enlaces de interés, actualizando la fecha de modificación.

### Queries.CountInterestLinksByCategory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Cuenta los enlaces asociados a una categoría como validación previa a su eliminación para retornar un error de dominio controlado.

### Queries.DeleteInterestLinkCategory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina una categoría de enlaces de interés por su identificador único.

### Queries.ListInterestLinksByCategory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Lista los enlaces pertenecientes a una categoría específica ordenados cronológicamente por su fecha de creación.

### Queries.ListInterestLinks
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Recupera la totalidad de los enlaces de interés sin filtrar para la vista plana de administración.

### Queries.CreateInterestLink
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Crea un nuevo enlace de interés con su categoría, título, descripción, color de tarjeta y URL objetivo.

### Queries.UpdateInterestLink
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Modifica la información de un enlace de interés existente y actualiza la fecha de modificación.

### Queries.DeleteInterestLink
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina un enlace de interés específico por su identificador único.

### Queries.CreateSuggestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inserta una sugerencia anónima registrando una foto del avance del usuario sin vincular identificadores de cuenta y evitando la cláusula `RETURNING` para cumplir con las restricciones de permisos del rol `usbi_app`.

### Queries.ListSuggestions
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Lista sugerencias utilizando el identificador UUIDv7 como cursor descendente para la paginación.

### Queries.DeleteSuggestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina un registro de sugerencia por su identificador único.
