---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Punto de entrada principal HTML5 para la aplicación de página única (SPA) cliente. Define los metadatos básicos del documento, el contenedor `div#root` donde se monta la aplicación de React y el script inicial modulado `src/main.tsx`.

Compilada por: [[frontend/Dockerfile.md|Dockerfile frontend]] (ejecuta `npm run build` en Node, resultando en `dist/`).
Servida por: [[frontend/deploy/nginx.conf.md|nginx.conf]] (rewrite SPA con `try_files`).
