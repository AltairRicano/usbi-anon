---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Controla la lógica del juego de memoria (memorama), gestionando la mezcla de cartas mediante el algoritmo Fisher-Yates, la restricción de volteo simultáneo a un máximo de dos cartas, la verificación de coincidencias y la garantía de finalización sin condición de derrota.

Relacionado con: [[frontend/packages/engine/src/games/MemoryEngine.test.ts.md|Test de MemoryEngine]], [[frontend/packages/engine/src/games/memory/MemoryPalette.ts.md|MemoryPalette]] (normalización de pares), [[frontend/packages/engine/src/interfaces/GameResult.ts.md|GameResult]], [[frontend/packages/schema/index.ts.md|@usbi/schema]].

## Funciones

### MemoryEngine.flipCard
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida y ejecuta el volteo de una carta según su posición en la baraja, impidiendo selecciones sobre cartas previamente reveladas, emparejadas o cuando ya se han seleccionado dos cartas.

### MemoryEngine.checkMatch
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica si la pareja de cartas seleccionadas comparte el mismo identificador, consolidando el emparejamiento o volteándolas nuevamente de no coincidir, y evalúa el fin del juego.

### MemoryEngine.getResult
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Retorna el resultado `GameResult` indicando la finalización del juego únicamente al emparejar la totalidad del mazo de cartas.
