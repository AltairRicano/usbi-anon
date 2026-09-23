---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Provee las operaciones administrativas para crear, actualizar, publicar, archivar y purgar niveles y secciones. Permite gestionar contenido no publicado o archivado utilizando transacciones de base de datos, registrando auditorías de administración con [[backend/internal/audit/audit.go.md#Log|audit.Log]] y acumulando el progreso histórico de los jugadores al purgar niveles. Reutiliza `validateLevelInput` de [[backend/internal/levels/service.go.md|service.go]] antes de escribir en base de datos.

## Funciones

### AdminService.CreateLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Crea un nivel no publicado en estado borrador dentro de una transacción con auditoría de administración tras validar la estructura y contenido de la plantilla.

### AdminService.UpdateLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza título, color, tipo de plantilla, dificultad y contenido JSON de un nivel dentro de una transacción auditada.

### AdminService.PublishLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Cambia el estado de un nivel a publicado para permitir el acceso a los jugadores, registrando la acción en la auditoría.

### AdminService.UnpublishLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Revoca la publicación de un nivel devolviéndolo a estado borrador y registrando auditoría.

### AdminService.ArchiveLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Realiza el borrado lógico de un nivel registrando el ID del administrador que lo elimina y guardando una entrada de auditoría.

### AdminService.UnarchiveLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Restaura un nivel previamente archivado verificando que se encuentre en estado eliminado.

### AdminService.ListArchivedLevels
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Devuelve la lista de niveles archivados permitiendo un filtrado opcional por sección.

### AdminService.PurgeLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina permanentemente un nivel archivado con [[backend/internal/repository/content_queries.go.md#Queries.PurgeLevel|Queries.PurgeLevel]] tras acumular previamente el progreso de los jugadores en `account_retired_progress` con [[backend/internal/repository/content_queries.go.md#Queries.AccumulateRetiredProgressForLevel|Queries.AccumulateRetiredProgressForLevel]] dentro de una transacción auditada — aplica la regla de rotación de temporadas descrita en [[Estado_Proyecto/Decisiones.md#rotacion_niveles_tres_fk_distintas|Decisiones.md]].

### AdminService.CreateSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Crea una nueva sección no publicada verificando campos obligatorios y registrando la auditoría correspondiente.

### AdminService.ListSections
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene las secciones registradas permitiendo incluir o excluir secciones no publicadas.

### AdminService.UpdateSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza el título, descripción y color de una sección en una transacción auditada.

### AdminService.PublishSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Habilita la publicación de una sección para hacerla visible en el catálogo de jugadores.

### AdminService.UnpublishSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Desactiva la publicación de una sección registrando el evento de auditoría.

### AdminService.ArchiveSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Archiva de forma lógica una sección con [[backend/internal/repository/content_queries.go.md#Queries.ArchiveSection|Queries.ArchiveSection]] y archiva automáticamente todos los niveles vinculados a ella con [[backend/internal/repository/content_queries.go.md#Queries.ArchiveLevelsBySection|Queries.ArchiveLevelsBySection]] — `levels.section_id` es `RESTRICT` (ver [[Estado_Proyecto/Decisiones.md#rotacion_niveles_tres_fk_distintas|Decisiones.md]]), así que purgar una sección exige purgar antes sus niveles.

### AdminService.UnarchiveSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Restaura una sección previamente archivada previa verificación de su estado.

### AdminService.ListArchivedSections
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Lista todas las secciones que están archivadas en el sistema.

### AdminService.PurgeSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina físicamente una sección archivada verificando obligatoriamente que no contenga niveles asociados.

### AdminService.GetLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la información completa de un nivel por su ID permitiendo visualizar contenido no publicado.

### AdminService.ListLevels
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Consulta de forma paginada por cursor los niveles, con soporte para filtrar por sección e incluir borradores.
