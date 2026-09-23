---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Componente de formulario para los metadatos generales de un nivel (sección, título, dificultad, color y tipo de plantilla). Implementa la regla de negocio de heredar automáticamente el color de la sección seleccionada y restringe la modificación de la sección y del tipo de plantilla al editar un nivel existente.

## Funciones

### LevelMetadataForm
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Renderiza los campos de entrada de metadatos, controlando la actualización del color según la sección elegida y bloqueando campos estructurales en modo de edición.

## Relaciones

- [[frontend/src/features/content/types.ts.md|types.ts]] — define tipos `SectionDTO`, `TemplateType` y `TEMPLATE_TYPE_LABELS`
- [[frontend/src/features/content/maker/LevelMakerForm.tsx.md|LevelMakerForm]] — componente padre que consume este formulario
