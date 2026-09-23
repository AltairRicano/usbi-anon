---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Manifiesto de configuración de Node.js para el proyecto frontend, gestionando dependencias del sistema, paquetes monorepo (`workspaces`) y scripts de automatización. Incluye librerías para renderizado UI (React, Phaser, Framer Motion), ruteo, validación (Zod) y scripts para compilación con límites explícitos de memoria Node.js.

Workspaces (paquetes internos):
- [[frontend/packages/engine/package.json.md|@usbi/engine]] (motores de juego)
- [[frontend/packages/schema/package.json.md|@usbi/schema]] (esquemas de validación)

Scripts principales:
- `npm run build` — compilado en [[frontend/Dockerfile.md|Dockerfile frontend]]
- `npm run test` — pruebas del monorepo
