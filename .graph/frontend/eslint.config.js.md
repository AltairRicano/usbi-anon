---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Configuración de linter basada en ESLint para validar la calidad del código TypeScript y React en la aplicación frontend. Aplica reglas recomendadas e impone de manera estricta estándares de accesibilidad web (`jsx-a11y`) para asegurar elementos ARIA y descripciones textuales válidas en componentes de interfaz.

Ejecutada por: scripts en [[frontend/package.json.md|package.json]] (linting).
Valida: código TypeScript/React del frontend y workspaces ([[frontend/packages/engine/src/index.ts.md|@usbi/engine]], [[frontend/packages/schema/index.ts.md|@usbi/schema]]).
