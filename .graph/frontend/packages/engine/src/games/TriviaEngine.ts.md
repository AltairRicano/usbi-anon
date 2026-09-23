---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Gobierna la mecánica de un cuestionario de opción múltiple con límite de tiempo por pregunta, aplicando bonificaciones de puntaje en función de la velocidad de respuesta, administrando temporizadores automáticos y evaluando la aprobación global frente a un umbral del 60%.

Relacionado con: [[frontend/packages/engine/src/games/TriviaEngine.test.ts.md|Test de TriviaEngine]], [[frontend/packages/engine/src/interfaces/GameResult.ts.md|GameResult]].

## Funciones

### TriviaEngine.startTimer
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicia una cuenta regresiva por segundo para la pregunta actual, enviando automáticamente una respuesta nula por expiración de tiempo al agotar los segundos configurados.

### TriviaEngine.stopTimer
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Detiene y destruye la ejecución del temporizador de la pregunta activa si existe uno en ejecución.

### TriviaEngine.submitAnswer
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Registra la opción seleccionada por el jugador, acredita puntos base más una bonificación proporcional al tiempo restante en caso de acierto, y agenda la transición hacia la siguiente pregunta.

### TriviaEngine.reset
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Limpia el estado del juego, detiene temporizadores en curso, restablece el contador de aciertos y puntaje, y recomienza la trivia desde la primera pregunta.

### TriviaEngine.getResult
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera la estructura `GameResult` determinando el cumplimiento del porcentaje de aprobación del 60% y adjuntando el historial de índices seleccionados por pregunta.
