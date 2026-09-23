---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Herramienta cliente para la creación y edición local de niveles (Maker). Permite previsualizar el contenido en tiempo real, validar la estructura según esquemas Zod por tipo de plantilla, guardar en `localStorage` o exportar las definiciones como archivos JSON mediante descargas Blob.

## Funciones

### generateId
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera un UUID mediante la API `crypto.randomUUID` o utiliza una rutina aleatoria de reserva.

### MakerPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente principal del editor Maker que coordina la edición de metadatos, selección de plantilla, estado de validación Zod y previsualización.

### MakerPage.handleTemplateChange
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza el tipo de plantilla del nivel e inicializa la estructura de contenido predeterminada desde el registro de plantillas.

### MakerPage.saveLocal
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida la ausencia de errores en el formulario y guarda o actualiza la estructura del nivel dentro del arreglo `usbi_local_levels` en `localStorage`.

### MakerPage.onExport
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Crea un objeto Blob con la estructura JSON del nivel y desencadena la descarga de un archivo local en el navegador.

## Relaciones

- Usa [[frontend/src/shared/components/ui/Button.tsx|Button]]
- Usa [[frontend/src/features/content/maker/registry.ts.md|registry]]
- Usa [[frontend/src/features/content/types.ts.md|types]]
