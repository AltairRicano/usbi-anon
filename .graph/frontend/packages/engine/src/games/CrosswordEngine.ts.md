---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Gestiona la generación dinámica y el estado interactivo de crucigramas interconectados, calculando automáticamente la disposición de palabras en una cuadrícula 2D, gestionando la navegación de usuario, el bloqueo de celdas correctas y la exportación de evidencias para validación externa.

Relacionado con: [[frontend/packages/engine/src/games/CrosswordEngine.test.ts.md|Test de CrosswordEngine]], [[frontend/packages/engine/src/interfaces/GameResult.ts.md|GameResult]].

## Funciones

### CrosswordEngine.inputChar
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Procesa la entrada de una letra en la celda seleccionada previo control de bloqueo e inmutabilidad, aplica normalización de caracteres y desplaza automáticamente el cursor a la celda adyacente según la orientación actual.

### CrosswordEngine.selectCell
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Establece la celda activa en las coordenadas especificadas, alternando la orientación entre horizontal y vertical si se reselecciona la misma celda o infiriéndola según la disponibilidad de celdas vecinas.

### CrosswordEngine.navigate
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Desplaza la selección de celda activa aplicando los desplazamientos `dx` y `dy` indicados si la celda de destino pertenece al tablero.

### CrosswordEngine.validate
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Compara el contenido ingresado contra la solución ignorando acentos, bloquea las celdas coincidentes contra futuras modificaciones, actualiza la puntuación y marca el juego como finalizado si todas las celdas son correctas.

### CrosswordEngine.reset
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Restablece el estado del motor a sus valores iniciales, limpiando las entradas del usuario, el puntaje, las celdas bloqueadas y la selección activa.

### CrosswordEngine.getSolvedWords
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Filtra y devuelve las palabras colocadas cuyas celdas constituyentes están totalmente consolidadas en el conjunto de celdas bloqueadas.

### CrosswordEngine.getResult
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera la estructura `GameResult` calculando el progreso del jugador, el puntaje obtenido y el listado de palabras resueltas para auditoría del servidor.

### normalizeCrosswordAnswer
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Sanitiza y normaliza cadenas eliminando diacríticos, convirtiendo a mayúsculas y descartando caracteres fuera del rango alfabético A-Z y Ñ.

### canBuildConnectedCrossword
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Evalúa si un conjunto de palabras puede organizarse completamente en una cuadrícula interconectada válida sin dejar elementos huérfanos.

### buildCrosswordLayout
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Construye iterativamente y selecciona la mejor disposición espacial en cuadrícula para una lista de palabras basándose en cruces de letras e intenciones glotonas de maquetado.
