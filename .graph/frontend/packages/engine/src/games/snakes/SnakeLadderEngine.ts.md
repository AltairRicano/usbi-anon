---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa el motor de juego para Serpientes y Escaleras compitiendo contra la IA. Controla la máquina de estados de los turnos, la lógica de rebote al sobrepasar la casilla de llegada, la resolución de desplazamientos por serpientes/escaleras y la toma de decisiones probabilísticas de la IA según el nivel de dificultad.

Relacionado con: [[frontend/packages/engine/src/games/snakes/SnakeLadderEngine.test.ts.md|Test de SnakeLadderEngine]], [[frontend/packages/engine/src/games/snakes/WeightedRandom.ts.md|WeightedRandom]] (selección ponderada), [[frontend/packages/engine/src/games/snakes/BoardMap.ts.md|BoardMap]] (configuración del tablero), [[frontend/packages/engine/src/interfaces/GameResult.ts.md|GameResult]].

## Funciones

### SnakeLadderEngine.start
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicia la partida cambiando el estado del juego al turno del jugador, asegurando que solo pueda ejecutarse cuando se encuentra en reposo ('idle').

### SnakeLadderEngine.rollPlayer
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Simula el lanzamiento de dado del jugador, calcula la casilla de destino aplicando rebote si excede la meta y procesa la casilla final resultando en serpiente, escalera o victoria.

### SnakeLadderEngine.playAITurn
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Gestiona la jugada de la IA mediante un cálculo aleatorio ponderado basado en la dificultad configurada, pudiendo optar por tiros óptimos para ganar o tiradas con rebote.

### SnakeLadderEngine.getResult
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Retorna el resultado consolidado del juego reflejando si el jugador resultó victorioso y asignando la puntuación correspondiente.
