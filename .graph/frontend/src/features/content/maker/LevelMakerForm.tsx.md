---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-28
---

Formulario principal del creador de niveles que coordina la definición de metadatos, la selección de plantillas, la edición de contenido y la previsualización en vivo. Contiene errores de renderizado dentro de un `FormErrorBoundary` para evitar la caída del panel administrativo y valida los esquemas Zod en tiempo real antes de permitir el guardado.

## Funciones

### LevelMakerForm
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente envoltorio principal que incluye la frontera de captura de errores para garantizar el aislamiento de fallos en el creador de niveles.

### LevelMakerFormInner
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Gerencia el estado integral del formulario de creación y edición, ejecutando validaciones Zod automáticas según la plantilla seleccionada y enviando la carga útil al servidor.

### FormErrorBoundary.getDerivedStateFromError
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Método estático de ciclo de vida que intercepta excepciones en los formularios hijo para capturar el mensaje de error y mostrar una interfaz de contingencia local.

## Relaciones

- [[frontend/src/features/content/maker/registry.ts.md|registry.ts]] — proporciona el registro de plantillas con esquemas, formularios y previsualizadores
- [[frontend/src/features/content/maker/LevelMetadataForm.tsx.md|LevelMetadataForm]] — componente para editar metadatos del nivel
- [[frontend/src/features/content/maker/LevelActions.tsx.md|LevelActions]] — componente para botones de guardar/cancelar
- [[frontend/src/features/content/types.ts.md|types.ts]] — define tipos `SectionDTO` y `TemplateType`
- [[frontend/src/features/content/AdminContentPage.tsx.md|AdminContentPage]] — consumidor del formulario
