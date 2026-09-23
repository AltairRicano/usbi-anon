---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Administra las operaciones relativas a insignias del sistema y su asignación a los jugadores según el XP acumulado. Incluye tanto el otorgamiento automático e inmutable de logros para los usuarios como la gestión CRUD administrativa para el catálogo de insignias. `AwardEligibleBadges` se invoca desde [[backend/internal/sync/service.go.md#Service.ProcessSync|sync/service.go#Service.ProcessSync]] al cierre de cada sincronización offline, y también desde el camino online en [[backend/internal/levels/player_service.go.md|levels/player_service.go]]; el CRUD del catálogo lo expone [[backend/internal/badges/admin_service.go.md|badges/admin_service.go]].

## Funciones

### Queries.AwardEligibleBadges
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inserta de forma atómica y sin duplicados las insignias cuyo umbral de XP sea alcanzado por el usuario, retornando las insignias otorgadas.

### Queries.ListUserBadges
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la lista de insignias otorgadas a un usuario ordenadas por el umbral de XP y la fecha de obtención.

### Queries.ListBadges
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene el catálogo completo de insignias ordenado de forma ascendente por el umbral de XP requerido.

### Queries.CreateBadge
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Crea una nueva insignia en el catálogo con su nombre, umbral de XP e icono.

### Queries.UpdateBadge
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza las propiedades de una insignia existente.

### Queries.CountUserBadgesByBadge
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Cuenta cuántos usuarios han ganado una insignia específica antes de permitir su eliminación en administración.

### Queries.DeleteBadge
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina físicamente una insignia de la base de datos por su ID.
