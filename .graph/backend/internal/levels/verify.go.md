---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa la verificación en el servidor de las respuestas enviadas por los jugadores para evitar manipulaciones del cliente. Recalcula de forma transparente el resultado para las plantillas de trivia, noticias falsas, crucigramas, sopa de letras y rompecabezas, dejando sin verificar intencionalmente las plantillas de memoria y serpientes y escaleras.

`verifyAnswers` la invoca [[backend/internal/levels/player_service.go.md#PlayerService.CompleteLevel|PlayerService.CompleteLevel]] dentro de su transacción serializable; para memory y snakes_ladders, ese mismo método acepta el `completed`/`score` reportado por el cliente sin verificarlo en servidor.

## Funciones

### verifyAnswers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Determina y ejecuta la rutina de verificación correspondiente para el tipo de plantilla indicado; retorna `ok=false` si el tipo no es verificable en servidor.

### verifyTriviaAnswers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Compara las opciones elegidas contra el contenido del nivel y exige un 60% de aciertos para marcar el nivel como completado.

### verifyFakeNewsAnswers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Evalúa las predicciones enviadas sobre noticias verdaderas/falsas comparándolas contra la propiedad `isFake` original.

### verifyCrosswordAnswers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida la lista de palabras resueltas contra las respuestas normalizadas del crucigrama, exigiendo el 100% de aciertos para completarlo.

### verifyWordSearchAnswers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Comprueba las palabras encontradas contra el contenido esperado, realizando normalización insensible a diacríticos y convirtiendo la letra Ñ a N.

### verifyPuzzleAnswers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica que la secuencia de piezas enviada por el cliente coincida exactamente con la ordenación original consecutiva.

### countDistinctMatches
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Cuenta aciertos de respuestas enviadas sin repetir palabras duplicadas para prevenir inflación artificial del puntaje.

### normalizeWordSearchWord
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Normaliza una palabra a mayúsculas sin acentos y convierte 'Ñ' en 'N' para acoplarse al motor de la sopa de letras.
