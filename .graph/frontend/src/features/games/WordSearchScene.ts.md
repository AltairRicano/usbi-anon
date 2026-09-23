---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Escena Phaser 3 que renderiza la cuadrícula de la sopa de letras. Permite la selección interactiva mediante arrastre del puntero (ratón o táctil) en 8 direcciones y ofrece soporte completo de accesibilidad con navegación por teclado (flechas, Enter/Espacio para marcar inicio y fin, y Escape para cancelar).

## Funciones

### WordSearchScene.init
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Guarda las referencias al `WordSearchEngine` y al callback `onFinish` proporcionados al inicializar la escena.

### WordSearchScene.create
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera la matriz de letras en pantalla, asigna manejadores de eventos táctiles/teclado y dibuja las palabras encontradas previamente.

### WordSearchScene.getGridCoord
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Mapea las coordenadas en píxeles del lienzo a las coordenadas de celda (columna, fila) en la cuadrícula si el puntero se encuentra dentro de sus límites.

### WordSearchScene.handlePointerDown
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicia la selección de una palabra al presionar sobre una celda válida de la cuadrícula.

### WordSearchScene.handlePointerMove
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Extiende la selección en progreso conforme el usuario mueve el puntero a través del tablero.

### WordSearchScene.handlePointerUp
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Concluye el arrastre del puntero y envía las celdas seleccionadas al motor para su validación.

### WordSearchScene.extendSelectionTo
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula la trayectoria recta entre la celda inicial de selección y la celda destino si forman una línea horizontal, vertical o diagonal válida.

### WordSearchScene.submitSelection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Envía las coordenadas de la selección al `WordSearchEngine` para validar la palabra y limpia las líneas gráficas de selección temporal.

### WordSearchScene.moveCursor
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Desplaza el cursor de teclado por las celdas de la cuadrícula y actualiza la línea de selección en caso de estar en modo de marcado.

### WordSearchScene.handleKeyDown
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Maneja los eventos de teclado para desplazar el cursor, iniciar o confirmar selecciones (Enter/Espacio) y cancelar la selección actual (Escape).

### WordSearchScene.drawCursor
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Dibuja en pantalla el marco visual de enfoque para la navegación accesible por teclado según las normas WCAG.

### WordSearchScene.drawSelection
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Dibuja el trazo gráfico translúcido que representa las celdas marcadas durante una selección activa.

### WordSearchScene.redrawFound
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Renderiza resaltados verdes permanentes sobre las palabras de la cuadrícula que han sido identificadas correctamente.

### WordSearchScene.findWordInGrid
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Busca la secuencia de coordenadas de una palabra dentro de la matriz 2D inspeccionando las ocho direcciones posibles.

### WordSearchScene.shutdown
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Limpia los escuchadores de eventos, se desubscribe del motor y destruye recursos de temporizadores al cerrar la escena.

## Relaciones

(Escena Phaser que usa @usbi/engine WordSearchEngine)
