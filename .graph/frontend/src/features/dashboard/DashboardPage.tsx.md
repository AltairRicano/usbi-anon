---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-28
---

Página principal del panel de usuario que articula el acceso a secciones oficiales, niveles locales y pestañas adicionales. Incluye un menú lateral accesible mediante atajo Escape y un control de importación de niveles locales con validación de seguridad de tamaño máximo de 5MB por archivo JSON.

## Funciones

### LocalContentSection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Renderiza los niveles locales almacenados en el navegador y proporciona opciones para crearlos, jugarlos, eliminarlos o importarlos.

### LocalContentSection.processImportedJSON
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida que el archivo JSON importado no supere el límite de 5MB de seguridad, comprueba la presencia de metadatos de nivel y actualiza el almacenamiento local.

### LocalContentSection.handleDelete
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Solicita confirmación al usuario y remueve la entrada del nivel seleccionado en `localStorage`.

### TemplateIcon
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Retorna el icono SVG correspondiente al tipo de plantilla del nivel (`trivia`, `puzzle`, `word_search`, `crossword`, `memory`, `snakes_ladders` o `fake_news`).

### SectionAccordionItem
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente de acordeón que realiza la carga perezosa de niveles de una sección desde `/levels` y genera una presentación visual en zigzag.

### DashboardPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente contenedor principal que gestiona las pestañas de navegación, la apertura/cierre del menú lateral por teclado y el cierre de sesión.

### DashboardPage.handleLogout
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Notifica el cierre de sesión al endpoint `/auth/logout` y limpia el estado global de autenticación del cliente.

## Relaciones

- [[frontend/src/features/auth/useAuthStore.ts.md|useAuthStore]] — consulta y modifica estado de autenticación
- [[frontend/src/features/content/types.ts.md|types.ts]] — define tipos `SectionDTO`, `LevelsPageDTO`, `TemplateType` y función `templateTypeLabel`
- [[frontend/src/features/content/schemas.ts.md|schemas.ts]] — valida secciones y niveles con esquemas Zod
