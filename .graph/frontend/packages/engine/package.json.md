---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define la configuración del paquete ESM privado `@usbi/engine`, estableciendo sus puntos de entrada tipados, la dependencia del esquema de datos `@usbi/schema` y los scripts de compilación (`tsc`) y pruebas (`vitest`).

Depende de: [[frontend/packages/schema/index.ts.md|@usbi/schema]].
Configuración de pruebas: [[frontend/packages/engine/vitest.config.ts.md|vitest.config.ts]].
Entrada: [[frontend/packages/engine/src/index.ts.md|src/index.ts]] (reexporta todos los motores).
