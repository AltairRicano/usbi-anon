---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define la configuración del paquete privado npm `@usbi/schema` dentro del monorepo del proyecto, especificando su punto de entrada en `index.ts` y declarando `zod` como su única dependencia para la definición y validación de esquemas.

Entrada: [[frontend/packages/schema/index.ts.md|src/index.ts]] (reexporta esquemas).
Utilizado por: [[frontend/packages/engine/package.json.md|@usbi/engine]] (dependencia).
