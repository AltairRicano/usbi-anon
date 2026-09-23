---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Guía operativa para el despliegue reproducible en producción del stack de 3 servicios (db, api, web) mediante docker-compose.yml. Define la secuencia obligatoria de arranque seguro utilizando usbictl para la generación e inicialización de secretos en .env (permisos 600), renderizado de configuración Postgres con SCRAM-SHA-256 en red interna, ejecución de migraciones con usbi_migrate y verificación de permisos con usbictl doctor.

Referencia:
- [[docker-compose.yml.md|docker-compose.yml]] — orquestación de los tres servicios
- [[plan/00_Plan_de_maduracion.md.md|Plan de maduración]] — contexto y fases (M3 — despliegue reproducible)
