---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Módulo de utilidades para el cálculo de geometría y distribución del tablero de Serpientes y Escaleras. Encargado de la generación pseudoaleatoria determinista de enlaces mediante PRNG por semilla y del mapeo de coordenadas en cuadrícula serpentina (boustrophedon).

## Funciones

### clampBoardSize
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Limita las dimensiones del tablero a un entero en el rango de 2 a 10.

### maxFeatureCount
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula el límite máximo permitido de serpientes o escaleras según el total de celdas del tablero.

### generateSnakeLadderLinks
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera listas de serpientes y escaleras asegurando el uso de una semilla PRNG y evitando colisiones de celdas de origen.

### cellAt
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Determina el número de celda (1 a N) según la fila y columna en una cuadrícula con distribución alternada por fila.

### cellCenterPercent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la posición porcentual (X, Y) del centro de una celda para el renderizado de gráficos sobre el tablero.

## Relaciones

- [[frontend/src/features/content/maker/registry.ts.md|registry.ts]] — utilizado en `snakeLadderDefaults` para generar configuración inicial del juego
