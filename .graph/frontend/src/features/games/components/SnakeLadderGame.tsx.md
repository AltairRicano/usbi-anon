---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Componente contenedor React para el juego de Serpientes y Escaleras. Combina la escena del tablero en Phaser (`SnakeLadderScene`) con un sistema de preguntas condicionales de opción múltiple; acertar una pregunta permite tirar el dado, mientras que fallarla coloca la pregunta al final de la cola y cede el turno a la IA.

## Funciones

### SnakeLadderGame
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Coordina la lógica principal del juego de Serpientes y Escaleras, el modal de preguntas condicionales y la integración con la escena Phaser.

### SnakeLadderGame.handleGameReady
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Almacena la referencia a la escena `SnakeLadderScene` en el estado local una vez que el juego Phaser ha completado su arranque.

### SnakeLadderGame.handleRollClick
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Muestra la siguiente pregunta condicional en la cola si existen preguntas configuradas, o ejecuta directamente el lanzamiento del dado si no las hay.

### SnakeLadderGame.executeRoll
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Emite el evento `ROLL_DICE` hacia la escena Phaser si el jugador dispone del turno y la escena no está realizando animaciones.

### SnakeLadderGame.handleAnswer
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Evalúa la respuesta del modal de pregunta; si es correcta activa el tiro de dado, y si es incorrecta reacola la pregunta y cede el turno a la IA.

### shuffledIndices
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Función utilitaria que genera un arreglo de índices enteros mezclados de forma aleatoria utilizando el algoritmo de Fisher-Yates.

## Relaciones

- Usa [[frontend/src/shared/PhaserGame.tsx|PhaserGame]]
- Usa [[frontend/src/shared/components/ui/Button.tsx|Button]]
- Usa [[frontend/src/features/games/phaser/SnakeLadderScene.ts|SnakeLadderScene]]
