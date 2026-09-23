---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define el esquema Zod `SnakesSchema` para el juego de serpientes y escaleras, validando la dimensión del tablero y refinando mediante reglas de negocio que las serpientes desciendan, las escaleras asciendan y no compartan casillas de origen. Asimismo, exige una banca mínima de 8 preguntas con opciones no vacías e índices válidos para evitar la repetición acelerada de preguntas durante la partida.

Utilizado por: [[frontend/packages/engine/src/games/snakes/SnakeLadderEngine.ts.md|SnakeLadderEngine]] para validación de configuración.
Reexportado por: [[frontend/packages/schema/index.ts.md|@usbi/schema]].
