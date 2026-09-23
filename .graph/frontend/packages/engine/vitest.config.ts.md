---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Configuración del entorno de pruebas unitarias Vitest para el motor de juegos. Especifica el entorno `jsdom` para simular APIs globales de navegador requeridas por temporizadores y excluye directorios de distribución y dependencias.

Referenciada por: [[frontend/packages/engine/package.json.md|@usbi/engine]] (script `test`).
Ejecuta tests: [[frontend/packages/engine/src/games/MemoryEngine.test.ts.md|MemoryEngine.test.ts]], [[frontend/packages/engine/src/games/snakes/SnakeLadderEngine.test.ts.md|SnakeLadderEngine.test.ts]], y otros.
