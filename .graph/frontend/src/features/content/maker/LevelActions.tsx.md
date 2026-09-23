---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Componente de la interfaz del creador de niveles que renderiza los controles principales de acción para guardar o cancelar la operación. Mantiene la invariante de deshabilitar el botón de guardado si el formulario contiene errores de validación o si existe un proceso de envío en ejecución.

## Funciones

### LevelActions
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente de UI que renderiza los botones de guardado y cancelación adaptando sus etiquetas y estado deshabilitado según los props de validación y edición.

## Relaciones

- [[frontend/src/features/content/maker/LevelMakerForm.tsx.md|LevelMakerForm]] — componente padre que consume este componente
