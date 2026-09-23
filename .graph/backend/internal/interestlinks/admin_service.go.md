---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa la lógica de negocio administrativa para categorías y enlaces de interés utilizando el pool de conexiones de moderador (`usbi_moderador`). Maneja transacciones en base de datos para operaciones de creación, actualización y eliminación, registrando de forma consistente cada acción con [[backend/internal/audit/audit.go.md#Log|audit.Log]]. Previene eliminaciones accidentales de categorías con enlaces activos consultando primero [[backend/internal/repository/interest_link_queries.go.md#Queries.CountInterestLinksByCategory|Queries.CountInterestLinksByCategory]] y traduce violaciones de integridad y unicidad de PostgreSQL (23505 y 23503) a errores de validación legibles. Las validaciones de campo (nombre, título, color, URL) que aplica antes de tocar la base de datos están en [[backend/internal/interestlinks/service.go.md|service.go]] (`validateCategoryInput`, `validateLinkInput`), y el DTO consumido/producido está definido en [[backend/internal/interestlinks/dto.go.md|dto.go]].

## Funciones

### AdminService.ListCategories
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene y retorna la lista completa de categorías de enlaces de interés mapeadas a DTOs de respuesta.

### AdminService.CreateCategory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida el nombre recibido, genera un UUID v7 e inserta la categoría con [[backend/internal/repository/interest_link_queries.go.md#Queries.CreateInterestLinkCategory|Queries.CreateInterestLinkCategory]] dentro de una transacción con registro de auditoría, capturando conflictos de unicidad.

### AdminService.UpdateCategory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida los datos y actualiza la categoría indicada dentro de una transacción con registro de auditoría, manejando casos de categoría inexistente o nombre duplicado.

### AdminService.DeleteCategory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica que la categoría no posea enlaces asociados con [[backend/internal/repository/interest_link_queries.go.md#Queries.CountInterestLinksByCategory|Queries.CountInterestLinksByCategory]], retornando `ErrCategoryHasLinks` antes de eliminarla con [[backend/internal/repository/interest_link_queries.go.md#Queries.DeleteInterestLinkCategory|Queries.DeleteInterestLinkCategory]] e insertar la entrada correspondiente en auditoría.

### AdminService.ListLinks
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Consulta y retorna el listado global de enlaces de interés convertidos a sus DTOs de respuesta.

### AdminService.CreateLink
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida la categoría de destino y los atributos del enlace (título, descripción, color hex, URL HTTP/HTTPS), creándolo en una transacción con auditoría.

### AdminService.UpdateLink
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida y actualiza los campos de un enlace de interés existente dentro de una transacción con auditoría, verificando la existencia de la categoría y del enlace.

### AdminService.DeleteLink
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina el enlace de interés especificado dentro de una transacción transaccional y registra el evento de eliminación en la tabla de auditoría.

### isUniqueViolation
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inspecciona el error de la base de datos para identificar si corresponde a una violación de restricción de unicidad de PostgreSQL (código 23505).

### isForeignKeyViolation
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inspecciona el error devuelto por la base de datos para determinar si se trata de una violación de clave foránea de PostgreSQL (código 23503).
