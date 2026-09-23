---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define la estrategia de construcción multi-etapa en Docker para el frontend, separando la fase de compilación en Node 20 del servidor web final en Nginx Alpine. Esta arquitectura optimiza la seguridad e infraestructura reduciendo el peso final de la imagen al omitir el código fuente y dependencias, evitando picos de memoria en entornos de producción restringidos.

Usado en: [[docker-compose.yml.md|docker-compose.yml]] (servicio `web`). Referencia: [[frontend/deploy/nginx.conf.md|nginx.conf]] (configuración del servidor web).
