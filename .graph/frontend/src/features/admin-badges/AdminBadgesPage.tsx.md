---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Interfaz para gestionar el catálogo global de insignias, permitiendo crear, editar y eliminar insignias por ID. Si al menos una cuenta de usuario ya ganó una insignia, el backend rechaza su eliminación con HTTP 409 (badge-has-holder), lo cual se presenta como un mensaje de error en la UI.

## Funciones

### AdminBadgesPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente principal para la visualización y administración del catálogo de insignias.

### AdminBadgesPage.loadBadges
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Realiza la petición GET /admin/badges y valida la respuesta para poblar el estado de insignias.

### AdminBadgesPage.createBadge
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Procesa el formulario de creación y envía los datos de la nueva insignia vía POST /admin/badges.

### AdminBadgesPage.saveEdit
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Guarda las modificaciones realizadas a una insignia mediante PATCH /admin/badges/{id}.

### AdminBadgesPage.deleteBadge
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Solicita la eliminación de una insignia mediante DELETE /admin/badges/{id}, gestionando el error 409 cuando ya fue otorgada.

## Relaciones

- [[frontend/src/features/admin-badges/schemas.ts.md|schemas.ts]] — valida respuestas de API con `BadgesResponseSchema`
