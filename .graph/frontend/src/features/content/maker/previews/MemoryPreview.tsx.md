---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Vista previa interactiva para la plantilla de memorama (juego de memoria). Normaliza y renderiza tanto la muestra del dorso común como los pares de tarjetas de contenido, calculando dinámicamente colores de texto legibles según el fondo de cada tarjeta.

## Funciones

### MemoryPreview
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Renderiza la vista previa del memorama, procesando la normalización de pares y del color del reverso de las tarjetas.

### MemoryBackSample
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Renderiza la tarjeta de muestra del dorso común utilizando los estilos de color normalizados.

### MemoryFrontSample
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Renderiza el anverso de una tarjeta de memoria determinando un color de texto legible adecuado al color de fondo recibido.

## Relaciones

- [[frontend/src/features/content/maker/registry.ts.md|registry.ts]] — previsualizador registrado en el registro de plantillas
