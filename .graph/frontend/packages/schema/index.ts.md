---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define los esquemas de validación Zod y tipos TypeScript para las plantillas de contenido y el formato de exportación del editor local (maker). Es una versión curada que elimina intencionalmente esquemas de identidad y PII (registro, login, consentimiento de tutor) para alinearse con el modelo de anonimato basado en cuestionario de gustos, excluyendo también plantillas obsoletas no utilizadas.

Utilizado por: [[frontend/packages/engine/src/games/MemoryEngine.ts.md|MemoryEngine]], [[frontend/packages/engine/src/games/snakes/SnakeLadderEngine.ts.md|SnakeLadderEngine]] y otros motores.
Reexporta: [[frontend/packages/schema/snakes.ts.md|snakes.ts]] (esquema del juego de serpientes y escaleras).

Forma parte del paquete npm `@usbi/schema`.
