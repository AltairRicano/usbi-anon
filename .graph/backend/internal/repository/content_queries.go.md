---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Contiene las consultas para la gestión de contenido educativo (secciones y niveles), seguimiento del progreso de jugadores, eventos de sincronización offline y logs de auditoría. Implementa reglas de integridad estrictas como la preservación de progreso retirado ante borrado de niveles, bloqueos consultivos de transacción (`pg_advisory_xact_lock`) para evitar intentos concurrentes duplicados, y filtrado seguro por estado de archivo/borrado. El CRUD administrativo de secciones y niveles lo consume [[backend/internal/levels/admin_service.go.md|levels/admin_service.go]]; el camino de juego (`LockLevelAttempt`, `CountLevelAttemptsByDate`, `UpsertPlayerProgressForAttempt`, `GetUserProgressTotals`) lo comparten [[backend/internal/levels/player_service.go.md|levels/player_service.go]] (online) y [[backend/internal/sync/service.go.md#Service.ProcessSync|sync/service.go#Service.ProcessSync]] (offline). `ListAuditLog` la expone [[backend/internal/auditlog/service.go.md|auditlog/service.go]].

## Funciones

### Queries.BeginTx
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicia una transacción SQL en el controlador de base de datos si la interfaz subyacente lo soporta.

### Queries.CreateSectionReturning
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inserta una nueva sección educativa y devuelve el registro creado.

### Queries.ListSections
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Recupera las secciones no eliminadas ni archivadas, con opción de incluir las despublicadas.

### Queries.UpdateSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Modifica el título, descripción y color de una sección no archivada ni eliminada.

### Queries.PublishSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Establece el estado de publicación de una sección en verdadero.

### Queries.UnpublishSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Cambia el estado de publicación de una sección a falso.

### Queries.ArchiveSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Archiva una sección registrando la fecha actual y despublicándola automáticamente.

### Queries.GetSectionByIDAny
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene una sección por su ID independientemente de si se encuentra archivada.

### Queries.ListArchivedSections
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Lista las secciones archivadas ordenadas descendentemente por fecha de archivo.

### Queries.UnarchiveSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Restaura una sección archivada eliminando la fecha de archivo.

### Queries.CountLevelsBySection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Cuenta la cantidad total de niveles asociados a una sección para validar su purga.

### Queries.PurgeSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina físicamente una sección si ya está archivada y no posee niveles vinculados.

### Queries.CreateLevelReturning
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inserta un nuevo nivel educativo con sus parámetros de configuración y contenido JSON.

### Queries.ListLevels
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene niveles no eliminados soportando paginación por cursor y filtros opcionales de sección y publicación.

### Queries.GetLevelByID
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene los datos detallados de un nivel activo por su ID.

### Queries.UpdateLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza los metadatos y contenido JSON de un nivel no eliminado.

### Queries.PublishLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Marca un nivel como publicado para los jugadores.

### Queries.UnpublishLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Retira la publicación de un nivel.

### Queries.ArchiveLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Realiza el borrado lógico de un nivel registrando la fecha y el administrador que ejecutó la acción.

### Queries.ArchiveLevelsBySection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Archiva en masa todos los niveles asociados a una sección específica.

### Queries.GetLevelByIDAny
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene un nivel por su ID incluyendo aquellos que han sido borrados lógicamente.

### Queries.ListArchivedLevels
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Lista los niveles en estado archivado/borrado lógico con filtro opcional por sección.

### Queries.UnarchiveLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Restaura un nivel archivado limpiando sus marcadores de borrado.

### Queries.AccumulateRetiredProgressForLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Transfiere y acumula el progreso de un nivel a la tabla de progreso retirado previo a su borrado físico para no perder contadores del jugador.

### Queries.PurgeLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Ejecuta el borrado físico de un nivel previamente archivado en la base de datos.

### Queries.LockLevelAttempt
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Solicita un bloqueo consultivo explícito en PostgreSQL basado en usuario, nivel y fecha para impedir registros duplicados concurrentes.

### Queries.CountLevelAttemptsByDate
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Retorna la cantidad de intentos de resolución registrados para un nivel por un usuario en una fecha específica.

### Queries.UpsertPlayerProgressForAttempt
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inserta o actualiza atómicamente el progreso acumulado del jugador (mejor puntaje, XP e intentos) tras completar un intento.

### Queries.InsertSyncEventWithPayload
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Registra un evento de sincronización fuera de línea con su payload, firmas de validación HMAC y estado inicial.

### Queries.UpdateSyncEventRejected
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Marca un evento de sincronización como rechazado especificando la razón de rechazo.

### Queries.GetUserProgressTotals
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula el XP total, niveles completados e intentos agregando los valores del progreso activo con el acumulado de niveles retirados.

### Queries.ListDailyStreakDates
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene las fechas de actividad reciente para calcular la racha diaria del jugador.

### Queries.ListUserProgressLevels
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Lista los niveles completados o intentados por un usuario vinculando la información del nivel activo.

### Queries.ListSyncEventsForUser
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene un listado paginado por fecha de recepción de los eventos de sincronización registrados por el usuario.

### Queries.ListAuditLog
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Consulta la bitácora de auditoría mediante filtros configurables por actor, acción, entidad y rango de fechas con paginación por cursor.
