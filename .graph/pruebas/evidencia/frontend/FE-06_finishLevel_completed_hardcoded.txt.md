---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Fragmento de código del frontend (OfficialLevelPage.tsx) que ilustra cómo los minijuegos (Memorama, Fake News, Serpientes y Escaleras) reportan su puntuación al completar una partida. Muestra la decisión de diseño donde el parámetro completed: true se envía fijado en duro (hardcoded) al servidor al finalizar cualquier juego.

[[pruebas/03_frontend/FE-06_superficie_de_ataque_cliente.md|Documento de análisis FE-06]]

## Funciones

### OfficialLevelPage.finishLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Función de retorno de llamada asíncrona enviada a los minijuegos que recibe la puntuación obtenida y reporta al backend el nivel como completado de manera incondicional (completed: true) junto con la marca de tiempo actual del cliente.
