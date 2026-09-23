---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Configuración del empaquetador Vite para el entorno de desarrollo, pruebas y previsualización. Configura plugins para React y Tailwind CSS, proxies para redireccionar peticiones de desarrollo `/api` al backend, exclusión de workspaces internos en Vitest y restricción de `allowedHosts` para peticiones de previsualización recibidas a través de túneles o dominios permitidos.

Referenciada por: [[frontend/Dockerfile.md|Dockerfile frontend]] (ejecuta `npm run build` basado en esta configuración).
Forma parte de la compilación que incluye: [[frontend/packages/engine/src/index.ts.md|@usbi/engine]], [[frontend/packages/schema/index.ts.md|@usbi/schema]].
