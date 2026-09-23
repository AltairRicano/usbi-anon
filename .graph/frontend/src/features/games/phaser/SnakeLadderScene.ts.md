---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Escena de Phaser 3 encargada de renderizar gráficamente el tablero de Serpientes y Escaleras. Controla la animación del movimiento de las fichas en las casillas, dibuja serpientes con curvas Bézier y escaleras, e implementa la animación del dado rodando con síntesis de sonido mediante la API WebAudio.

## Funciones

### SnakeLadderScene.init
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Recupera la configuración y el motor `SnakeLadderEngine` almacenados en el registro de la instancia de Phaser.

### SnakeLadderScene.create
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Dibuja el HUD de información, el tablero, las serpientes, escaleras y fichas de juego, registrando los escuchadores para eventos de lanzamiento de dado y jugada de la IA.

### SnakeLadderScene.playAITurnAnimated
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Ejecuta el turno de la IA en el motor, activa la animación del dado y desplaza la ficha de la IA en el tablero.

### SnakeLadderScene.playDiceRollAnimation
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Muestra una animación de ~1 segundo alternando caras aleatorias del dado con efectos de sonido antes de fijar el valor definitivo e invocar la animación de movimiento.

### SnakeLadderScene.drawDieFace
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Dibuja en el panel HUD la representación gráfica de un dado con sus respectivos puntos (pips) según el valor numérico provisto.

### SnakeLadderScene.ensureAudioContext
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Crea o reanuda un contexto `AudioContext` de la WebAudio API para la generación de sonidos sintéticos sin requerir archivos de audio en disco.

### SnakeLadderScene.playDiceTick
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera un efecto de sonido sintético breve de giro o impacto de dado usando osciladores de WebAudio, comprobando la preferencia de silencio de sonido del usuario.

### SnakeLadderScene.updateState
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza el texto descriptivo del HUD y activa la animación de desplazamiento de la ficha correspondiente ('player' o 'ai').

### SnakeLadderScene.animateToken
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Coordina una secuencia de animaciones (`tweens.chain`) para trasladar casilla por casilla una ficha desde su posición inicial a la final, manejando rebotes al final del tablero y saltos por serpientes/escaleras.

### SnakeLadderScene.calculateLayout
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Determina las dimensiones de celda y los márgenes necesarios para adaptar y centrar el tablero dentro de los límites del canvas.

### SnakeLadderScene.drawHud
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Crea los elementos gráficos y textos del panel superior (HUD) para mostrar los mensajes de turno y resultados del dado.

### SnakeLadderScene.drawBoard
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Dibuja la matriz de casillas del tablero destacando numéricamente cada posición e identificando la casilla de inicio y la meta.

### SnakeLadderScene.drawLinks
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Recorre los elementos de serpientes y escaleras configurados para llamar a las funciones encargadas de su representación gráfica.

### SnakeLadderScene.drawSnake
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Dibuja una curva de Bézier cuadrática de color rojo que representa una serpiente uniendo su casilla de inicio y fin.

### SnakeLadderScene.drawLadder
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Dibuja las líneas paralelas y peldaños de una escalera de color verde entre dos casillas del tablero.

### SnakeLadderScene.drawTokens
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Instancia los círculos de color azul y violeta que representan las fichas del jugador y la IA en el tablero.

### SnakeLadderScene.cellCenter
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula el punto central (x, y) en píxeles de una casilla del tablero a partir de su número ordinal.

### SnakeLadderScene.tokenPoint
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula la posición de renderizado de la ficha aplicando un desplazamiento lateral respecto al centro de la casilla para evitar solapamientos entre fichas.

### SnakeLadderScene.shutdown
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Detiene los temporizadores activos de tirada de dado y cierra el contexto de audio WebAudio para liberar recursos del navegador.

## Relaciones

- Usa [[frontend/src/features/settings/useSettingsStore.ts|useSettingsStore]]
