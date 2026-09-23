---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa el servicio encargado del envío de sugerencias por parte de los jugadores utilizando el pool de base de datos usbi_app. Preserva el anonimato desvinculando la identidad del usuario tras consultar su progreso, almacenando un snapshot estático de nivel y experiencia sin crear claves foráneas ni registros de auditoría correlacionables. Invocado por [[backend/internal/suggestions/handler.go.md#Handler.Submit|handler.go#Handler.Submit]].

## Funciones

### PlayerService.Submit
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida que el texto de la sugerencia sea no vacío y no exceda los 1000 caracteres, recupera el resumen de progreso del jugador mediante [[backend/internal/repository/content_queries.go.md#Queries.GetUserProgressTotals|Queries.GetUserProgressTotals]] y genera la entrada con un ID UUIDv7 vía [[backend/internal/repository/interest_link_queries.go.md#Queries.CreateSuggestion|Queries.CreateSuggestion]]. Retorna la confirmación mínima sin guardar ninguna referencia a la cuenta del autor.
