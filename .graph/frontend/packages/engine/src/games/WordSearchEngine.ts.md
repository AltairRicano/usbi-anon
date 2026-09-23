---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Mantiene el estado y la lógica de juego de la sopa de letras, incluyendo la generación determinista del tablero mediante un PRNG lineal personalizado. Se encarga de normalizar las palabras eliminando caracteres especiales y acentos, posicionarlas en direcciones válidas, comprobar aciertos en sentido directo e inverso y notificar a los suscriptores.

Relacionado con: [[frontend/packages/engine/src/games/WordSearchEngine.test.ts.md|Test de WordSearchEngine]], [[frontend/packages/engine/src/interfaces/GameResult.ts.md|GameResult]].

## Funciones

### WordSearchEngine.checkWord
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica si la secuencia de letras seleccionada coincide con alguna palabra pendiente (en orden normal o invertido), incrementando el puntaje y marcando el juego como completado al encontrar todas las palabras.

### WordSearchEngine.reset
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Restablece el progreso del juego reiniciando las palabras encontradas, el puntaje acumulado y el estado de finalización, notificando la actualización a los oyentes.

### WordSearchEngine.getResult
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula y retorna la estructura del resultado final del juego indicando si fue completado, las palabras encontradas y la puntuación obtenida.

### WordSearchEngine.subscribe
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Suscibe una función observadora a los cambios de estado del motor, ejecutándola inmediatamente con el estado actual y devolviendo una función de cancelación.
