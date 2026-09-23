---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Ofrece la consulta pública de solo lectura del carrusel de enlaces de interés para la vista de jugador utilizando el pool de conexiones `usbi_app`. Agrupa las categorías activas con sus enlaces ejecutando dos consultas secuenciales en lugar de un JOIN para optimizar el rendimiento y evitar deduplicaciones en memoria sobre contenido curado.

## Funciones

### PlayerService.ListGrouped
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene las categorías con [[backend/internal/repository/interest_link_queries.go.md#Queries.ListInterestLinkCategories|Queries.ListInterestLinkCategories]] e itera sobre ellas para recuperar sus enlaces asociados con [[backend/internal/repository/interest_link_queries.go.md#Queries.ListInterestLinksByCategory|Queries.ListInterestLinksByCategory]], estructurando la respuesta agrupada ([[backend/internal/interestlinks/dto.go.md|dto.go]]) que consume [[backend/internal/interestlinks/handler.go.md#Handler.ListInterestLinks|Handler.ListInterestLinks]] para el jugador.
