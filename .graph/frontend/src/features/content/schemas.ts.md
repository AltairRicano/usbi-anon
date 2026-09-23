---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Archivo de definición de esquemas Zod que replican las estructuras DTO del backend en Go para secciones, niveles, paginación, insignias y progreso del perfil. Sirve para validar la integridad de las respuestas API recibidas en la aplicación.

## Relaciones

- OfficialLevelPage — valida nivel con `LevelDTOSchema` y respuesta con `CompleteLevelResponseSchema`
- SectionLevelsPage — valida con `SectionsResponseSchema` y `LevelsPageDTOSchema`
- AdminContentPage — valida secciones, niveles y estados archivados
- DashboardPage — valida secciones y niveles
