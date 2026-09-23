---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Ofrece una función de utilidad probabilística para seleccionar elementos de una lista según una distribución de pesos. Aplica validaciones de seguridad para prevenir arreglos vacíos, discrepancias entre opciones y pesos, pesos negativos o suma total de pesos nula.

Utilizado por: [[frontend/packages/engine/src/games/snakes/SnakeLadderEngine.ts.md|SnakeLadderEngine]] (decisiones de IA).

## Funciones

### weightedRandom
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Selecciona y devuelve un elemento del arreglo `options` en función de los pesos relativos especificados y un generador de números aleatorios.
