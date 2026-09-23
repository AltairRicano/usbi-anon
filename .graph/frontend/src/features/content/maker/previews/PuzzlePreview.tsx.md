---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Componente de vista previa para el minijuego de rompecabezas. Utiliza carga diferida (lazy loading) del componente `PuzzleGame` para brindar una previsualización interactiva en tiempo real si se ha definido una frase válida.

## Funciones

### PuzzlePreview
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Carga asíncronamente el juego de rompecabezas pasándole la frase, número de piezas y semilla configurados, o muestra una indicación si la frase no ha sido ingresada.

## Relaciones

- [[frontend/src/features/content/maker/registry.ts.md|registry.ts]] — previsualizador registrado en el registro de plantillas
