---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Archivo de orquestación Docker para desplegar los tres servicios de producción (db con Postgres 15, api con backend Go, y web con frontend Nginx/React). Aísla la base de datos y la API en una red interna privada, impone límites de memoria de 250MB por contenedor, inyecta contraseñas obligatorias leídas de .env y configura comprobaciones de salud (healthcheck) para asegurar un arranque ordenado.

Referencia:
- Dockerfile del backend: [[backend/Dockerfile.md|Dockerfile backend]]
- Dockerfile del frontend: [[frontend/Dockerfile.md|Dockerfile frontend]]
- Configuración de Nginx: [[frontend/deploy/nginx.conf.md|nginx.conf]]

Documentado en: [[plan/00_Plan_de_maduracion.md|Plan de maduración]] (M3 — despliegue reproducible).
