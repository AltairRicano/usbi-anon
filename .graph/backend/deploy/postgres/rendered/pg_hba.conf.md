---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Versión final renderizada de `pg_hba.conf` generada por la herramienta `usbictl pgconf render`. Implementa autenticación obligatoria `scram-sha-256` y restringe el acceso de red a la subred `172.28.0.0/16` utilizada internamente por Docker Compose.

## Relaciones

- [[backend/deploy/postgres/pg_hba.conf.tmpl|pg_hba.conf.tmpl]]: Plantilla origen renderizada por usbictl
