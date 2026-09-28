---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-28
---

Página de administración dividida en pestañas para gestionar las categorías y tarjetas del carrusel de enlaces de interés, así como el buzón de sugerencias anónimas. Garantiza que las categorías con enlaces no se eliminen directamente (error 409 category-has-links) y soporta paginación por cursor en el buzón.

## Funciones

### AdminCommunityPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente vista principal para la administración de enlaces de interés y sugerencias anónimas.

### AdminCommunityPage.loadCategories
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Carga el listado de categorías mediante GET /admin/interest-link-categories.

### AdminCommunityPage.loadLinks
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene las tarjetas de enlaces de interés registradas llamando a GET /admin/interest-links.

### AdminCommunityPage.createCategory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Envía una solicitud POST /admin/interest-link-categories para registrar una nueva categoría con su nombre y orden.

### AdminCommunityPage.saveCategory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza el nombre y orden de una categoría existente vía PATCH /admin/interest-link-categories/{id}.

### AdminCommunityPage.deleteCategory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina una categoría vía DELETE /admin/interest-link-categories/{id}, capturando conflictos si tiene enlaces asociados.

### AdminCommunityPage.createLink
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Registra una nueva tarjeta de enlace con título, descripción, color, categoría y URL vía POST /admin/interest-links.

### AdminCommunityPage.saveLink
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Edita una tarjeta de enlace existente mediante PATCH /admin/interest-links/{id}.

### AdminCommunityPage.deleteLink
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina una tarjeta de enlace del carrusel mediante DELETE /admin/interest-links/{id}.

### AdminCommunityPage.categoryName
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Función auxiliar que retorna el nombre visible de una categoría basándose en su ID.

### AdminCommunityPage.loadSuggestions
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene las sugerencias anónimas del buzón de manera paginada utilizando un cursor mediante GET /admin/suggestions.

### AdminCommunityPage.openSuggestionsTab
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Activa la pestaña del buzón de sugerencias e inicia la carga de datos si no se habían recuperado antes.

### AdminCommunityPage.deleteSuggestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina una sugerencia del buzón vía DELETE /admin/suggestions/{id}.

## Relaciones

- [[frontend/src/features/admin-community/schemas.ts.md|schemas.ts]] — valida categorías, enlaces y sugerencias con esquemas Zod
