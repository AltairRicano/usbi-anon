---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-28
---

Vista pública del catálogo de niveles pertenecientes a una sección específica. Se encarga de solicitar al servidor únicamente las secciones y niveles oficiales que se encuentran en estado publicado para presentarlos al jugador.

## Funciones

### SectionLevelsPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente que obtiene de la API los datos de la sección elegida y sus niveles publicados, desplegándolos en tarjetas con enlace directo a la vista de juego.

## Relaciones

- [[frontend/src/features/content/schemas.ts.md|schemas.ts]] — valida secciones y niveles con `SectionsResponseSchema` y `LevelsPageDTOSchema`
- [[frontend/src/features/content/types.ts.md|types.ts]] — define tipos como `SectionDTO`, `LevelSummaryDTO` y `templateTypeLabel`
