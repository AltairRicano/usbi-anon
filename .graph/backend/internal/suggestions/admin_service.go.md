---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona la lógica del servicio de administración para listar y eliminar sugerencias anónimas operando sobre el pool de base de datos de moderación (usbi_moderador). Destaca por omitir el contenido de la sugerencia en los registros de auditoría durante la eliminación para no frustrar la depuración del buzón, y por utilizar paginación por cursores mediante UUIDv7. Expuesto vía HTTP por [[backend/internal/suggestions/handler.go.md#Handler.List|handler.go#Handler.List]] y [[backend/internal/suggestions/handler.go.md#Handler.Delete|handler.go#Handler.Delete]], montados en transport/router.go bajo `/admin/suggestions`.

## Funciones

### AdminService.List
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene una página de sugerencias mediante [[backend/internal/repository/interest_link_queries.go.md#Queries.ListSuggestions|Queries.ListSuggestions]], utilizando un cursor UUIDv7 y ajustando el tamaño de página entre 20 y 50 elementos. Solicita un registro adicional a la base de datos para verificar si existen más resultados y construye la respuesta paginada con el siguiente cursor.

### AdminService.Delete
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina una sugerencia por su identificador dentro de una transacción SQL ([[backend/internal/repository/interest_link_queries.go.md#Queries.DeleteSuggestion|Queries.DeleteSuggestion]]) y registra la acción mediante [[backend/internal/audit/audit.go.md|audit/audit.go]]. Para proteger la privacidad, la entrada de auditoría registra únicamente el actor y el ID de la entidad, omitiendo el texto de la sugerencia eliminada.
