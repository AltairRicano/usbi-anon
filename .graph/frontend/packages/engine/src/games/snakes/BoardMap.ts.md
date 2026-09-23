---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Calcula la conversión de un índice unidimensional (1-based) en coordenadas cartesianas `(x, y)` dentro del tablero de serpientes y escaleras. Aplica un patrón de recorrido serpenteante (boustrophedon) que comienza desde la parte inferior y alterna la dirección horizontal en filas impares, validando límites de rango.

Utilizado por: [[frontend/packages/engine/src/games/snakes/SnakeLadderEngine.ts.md|SnakeLadderEngine]].

## Funciones

### mapToGrid
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Transforma la posición numérica de una casilla dentro del tablero a sus coordenadas bidimensionales `(x, y)`, arrojando una excepción si el índice está fuera de rango.
