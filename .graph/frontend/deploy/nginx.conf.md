---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Configuración del servidor Nginx que actúa como servidor web estático de producción y proxy inverso para el frontend. Incluye reglas de compresión gzip, cabeceras de seguridad HTTP (`nosniff`, `DENY` en marcos, `strict-origin`), reescritura de rutas SPA para HTML5 (`try_files`), políticas de caché agresivas e inmutables para assets compilados y reenvío de peticiones `/api/` hacia el backend.

Referenciada por: [[frontend/Dockerfile.md|Dockerfile frontend]] (copiada en el contenedor). Usado en: [[docker-compose.yml.md|docker-compose.yml]] (servicio `web`).
