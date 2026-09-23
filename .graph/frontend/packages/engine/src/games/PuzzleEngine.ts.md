---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Modela un juego de rompecabezas de frases divididas en piezas, asegurando la fragmentación equilibrada del texto, la mezcla determinista basada en semillas generadoras y la reordenación de piezas con penalización progresiva en el puntaje según la cantidad de movimientos realizados.

Relacionado con: [[frontend/packages/engine/src/games/PuzzleEngine.test.ts.md|Test de PuzzleEngine]], [[frontend/packages/engine/src/interfaces/GameResult.ts.md|GameResult]].

## Funciones

### PuzzleEngine.reorderPieces
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Mueve una pieza de una posición a otra dentro del arreglo de piezas, incrementa el contador de movimientos realizados y evalúa si la secuencia actual coincide con la solución original.

### PuzzleEngine.reset
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Restablece el estado del rompecabezas, reiniciando los movimientos y el puntaje, y regenerando el desordenamiento inicial de las piezas mediante el algoritmo pseudoaleatorio determinista.

### PuzzleEngine.getResult
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Construye la respuesta `GameResult` reportando las piezas correctamente colocadas, el estado de resolución total y la secuencia ordenada final de los índices de las piezas.
