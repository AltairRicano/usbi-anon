---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Administra la mecánica del juego de identificación de noticias falsas, evaluando en orden secuencial las respuestas del usuario sobre la veracidad de cada noticia, registrando el historial de elecciones y determinando la aprobación según un umbral de aciertos del 60%.

Relacionado con: [[frontend/packages/engine/src/games/FakeNewsEngine.test.ts.md|Test de FakeNewsEngine]], [[frontend/packages/engine/src/interfaces/GameResult.ts.md|GameResult]].

## Funciones

### FakeNewsEngine.answer
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Evalúa la respuesta brindada para la noticia actual, actualiza la puntuación en caso de acierto, almacena la elección en el historial de evidencia y avanza el índice a la siguiente noticia.

### FakeNewsEngine.getResult
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula la métrica final de rendimiento del jugador, determinando si se alcanzó la razón de aprobación del 60% y emitiendo el objeto `GameResult` con la secuencia completa de elecciones.
