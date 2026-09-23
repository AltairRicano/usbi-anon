---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Componente React para el juego de rompecabezas de palabras ("Mensaje Secreto"). Utiliza la biblioteca `framer-motion` (`Reorder`) para permitir al usuario arrastrar y reordenar visualmente fragmentos de texto en un eje horizontal hasta recomponer la frase original.

## Funciones

### PuzzleGame
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicializa el `PuzzleEngine`, suscribe el componente a sus cambios de estado y renderiza la interfaz contenedora con la lista de fichas reordenables.

### PuzzleGame.handleReorder
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Compara la posición previa y la nueva de los elementos reordenados mediante arrastre para notificar al `PuzzleEngine` sobre el intercambio de fichas realizado.

## Relaciones

- Usa [[frontend/src/shared/components/ui/Card.tsx|Card]]
