---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Representa la capa de presentación HTTP para el paquete de niveles. Inspecciona las credenciales JWT en el contexto de la solicitud para validar roles (`canManageContent`, `canArchiveContent`) y despacha la ejecución hacia [[backend/internal/levels/admin_service.go.md|AdminService]] o [[backend/internal/levels/player_service.go.md|PlayerService]] según corresponda, devolviendo respuestas JSON o errores en formato Problem Details.

## Funciones

### Handler.CreateLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida permisos de gestor de contenido, decodifica el cuerpo HTTP estrictamente y delega la creación del nivel a `AdminService`.

### Handler.ListLevels
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Parsea parámetros de paginación por cursor y sección, despachando la consulta a `AdminService` o `PlayerService` según el rol JWT del usuario.

### Handler.GetLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene el ID del nivel desde la ruta y llama a `AdminService` o `PlayerService` en función de los permisos del cliente.

### Handler.UpdateLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Comprueba los permisos de gestión de contenido y actualiza un nivel existente mediante `AdminService`.

### Handler.PublishLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica autorización y publica un nivel a través del servicio de administración.

### Handler.UnpublishLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida permisos y retira la publicación de un nivel.

### Handler.ArchiveLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica permisos de archivado y realiza el borrado lógico del nivel indicado.

### Handler.ListArchivedLevels
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Retorna el listado de niveles archivados filtrados opcionalmente por sección para usuarios autorizados.

### Handler.UnarchiveLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida rol administrativo y restaura un nivel previamente archivado.

### Handler.PurgeLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Ejecuta el purgado físico e irreversible de un nivel archivado tras validar permisos, delegando en [[backend/internal/levels/admin_service.go.md#AdminService.PurgeLevel|AdminService.PurgeLevel]].

### Handler.CompleteLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Extrae el ID del usuario del JWT y procesa el intento de juego del nivel mediante [[backend/internal/levels/player_service.go.md#PlayerService.CompleteLevel|PlayerService.CompleteLevel]].

### Handler.GetProfileProgress
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Consulta y devuelve las estadísticas de progreso, racha e insignias del perfil del jugador autenticado.

### Handler.CreateSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida permisos de gestión y crea una nueva sección vía `AdminService`.

### Handler.ListSections
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene las secciones despachando a `AdminService` o `PlayerService` según las credenciales del solicitante.

### Handler.UpdateSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza metadatos de una sección existente previa validación de rol.

### Handler.PublishSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica autorización y publica la sección especificada.

### Handler.UnpublishSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Desactiva la publicación de una sección si el usuario tiene rol de gestor.

### Handler.ArchiveSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Archiva una sección y sus niveles tras validar permisos administrativos.

### Handler.ListArchivedSections
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Devuelve las secciones archivadas para los gestores de contenido autorizados.

### Handler.UnarchiveSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Restaura una sección archivada previa verificación de permisos.

### Handler.PurgeSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Procesa la eliminación permanente de una sección archivada asegurando los requisitos de autorización.
