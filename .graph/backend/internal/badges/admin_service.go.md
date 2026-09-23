---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona la lógica de negocio administrativa para el catálogo de insignias operando sobre el pool de datos de moderación (`usbi_moderador`). Administra transacciones SQL con registro automático de auditoría vía [[backend/internal/audit/audit.go.md#Log|audit.Log]], genera identificadores UUID v7, mapea violaciones de unicidad de Postgres y evita la eliminación de insignias que ya hayan sido otorgadas a usuarios. Es el destino de cada endpoint de [[backend/internal/badges/handler.go.md|handler.go]] y valida la entrada con [[backend/internal/badges/service.go.md#validateBadgeInput|validateBadgeInput]].

## Funciones

### AdminService.List
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la lista completa de insignias con [[backend/internal/repository/badge_queries.go.md#Queries.ListBadges|Queries.ListBadges]] y las mapea al formato de respuesta con `toResponse` ([[backend/internal/badges/dto.go.md|dto.go]]).

### AdminService.Create
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Sanea y valida los datos recibidos, genera un UUID v7 y crea una nueva insignia con [[backend/internal/repository/badge_queries.go.md#Queries.CreateBadge|Queries.CreateBadge]] en una transacción SQL, registrando la acción con [[backend/internal/audit/audit.go.md#Log|audit.Log]] y traduciendo errores de duplicidad de nombre.

### AdminService.Update
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida la entrada y actualiza los campos de una insignia existente con [[backend/internal/repository/badge_queries.go.md#Queries.UpdateBadge|Queries.UpdateBadge]] en una transacción SQL, registrando la auditoría del cambio con [[backend/internal/audit/audit.go.md#Log|audit.Log]] y controlando errores de inexistencia o duplicidad.

### AdminService.Delete
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica con [[backend/internal/repository/badge_queries.go.md#Queries.CountUserBadgesByBadge|Queries.CountUserBadgesByBadge]] si la insignia tiene usuarios asignados antes de intentar borrarla con [[backend/internal/repository/badge_queries.go.md#Queries.DeleteBadge|Queries.DeleteBadge]], rechazando la operación con error de dominio (`ErrBadgeHasHolder`) si tiene titulares para evitar errores de base de datos y mantener la intangibilidad de logros obtenidos, registrando la auditoría con [[backend/internal/audit/audit.go.md#Log|audit.Log]] en caso exitoso.
