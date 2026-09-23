---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Página principal de administración de contenido donde los administradores gestionan secciones y niveles oficiales (creación, edición, publicación, ocultamiento, archivado y purga). Incorpora mecanismos de seguridad como doble confirmación tipeando el título exacto para la purga irreversible de entidades y adapta el formato JSON para la importación y exportación de niveles desde el creador local.

## Funciones

### AdminContentPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente principal de administración que gestiona la carga de contenido publicado y archivado, la creación/edición de secciones y la exportación e importación de niveles mediante archivos JSON.

### AdminContentPage.exportLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera y descarga un archivo JSON formateado para el maker local, mapeando la fecha de creación del servidor y descartando campos inexistentes en base de datos como el autor.

### AdminContentPage.onFileSelected
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Procesa la carga de un archivo JSON exportado por el creador local, validando su estructura de metadatos/contenido e informando sobre el descarte de autoría antes de pasar la información al formulario.

### ArchivedRow
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente para elementos archivados que exige escribir el título exacto de la sección o nivel como barrera de seguridad para habilitar la purga definitiva e irreversible.

## Relaciones

- [[frontend/src/features/content/schemas.ts.md|schemas.ts]] — valida respuestas de API con esquemas para secciones, niveles, archivados
- [[frontend/src/features/content/types.ts.md|types.ts]] — define tipos DTO y función `templateTypeLabel`
- [[frontend/src/features/content/maker/LevelMakerForm.tsx.md|LevelMakerForm]] — componente de formulario para crear/editar niveles
