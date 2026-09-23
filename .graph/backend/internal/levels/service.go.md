---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Agrupa utilidades de dominio, constantes de error y funciones de validación compartidas. Incluye el cálculo de puntos de experiencia (XP) ajustado por número de intento y los validadores de esquema/contenido específicos para los 7 tipos de plantillas de juego. `validateLevelInput` la invocan tanto [[backend/internal/levels/admin_service.go.md#AdminService.CreateLevel|AdminService.CreateLevel]] como `AdminService.UpdateLevel`; `CalculateXP` y `calculateCurrentStreak` las invoca [[backend/internal/levels/player_service.go.md#PlayerService.CompleteLevel|PlayerService.CompleteLevel]].

## Funciones

### CalculateXP
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula la experiencia otorgada según la dificultad del nivel y el número de intento (1er intento: base completa, 2do y 3er intento: 50%, 4to o superior: 0).

### validateLevelInput
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida presencia de campos básicos, límites de dificultad (1 a 10), límites de tamaño del JSON (hasta 5 MB) y ejecuta el validador específico según la plantilla.

### validateTriviaContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida que el contenido de trivia tenga al menos una pregunta con 2 a 4 opciones no vacías e índice de respuesta correcta válido.

### validateMemoryContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Exige al menos 4 parejas con identificador y contenidos de texto válidos.

### validateFakeNewsContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida la estructura de noticias falsas requiriendo título, contenido, referencia, indicador booleano `isFake` y URL válida si se incluye imagen.

### validateWordSearchContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica que haya al menos 2 palabras de longitud adecuada y que las dimensiones de la cuadrícula estén entre 5 y 24.

### validatePuzzleContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Comprueba que la frase no esté vacía y que el número de piezas se encuentre en el rango permitido (3 a 20).

### validateCrosswordContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida el contenido del crucigrama limitando la lista a máximo 30 palabras únicas con pistas y comprobando que puedan interconectarse en un tablero.

### canBuildConnectedCrossword
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Algoritmo que evalúa si un conjunto de candidatos de palabras puede entrecruzarse de forma válida para formar un crucigrama conexo.

### validateSnakesContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida tableros de serpientes y escaleras asegurando rangos de casillas, configuración de dificultad de la IA y un banco mínimo de 8 preguntas con 2 opciones.

### calculateCurrentStreak
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula la cantidad de días consecutivos de actividad del jugador evaluando las fechas almacenadas desde la fecha actual hacia atrás.
