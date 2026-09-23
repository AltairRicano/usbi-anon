---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Establece la matriz centralizada y versionada de permisos GRANT/REVOKE para los roles usbi_app, usbi_moderador y usbi_dbmaint siguiendo el principio de menor privilegio. Restringe a la aplicación HTTP la escritura en catálogos y lectura de registros sensibles, otorga permisos administrativos selectivos a moderación (incluyendo restricciones a nivel de columna y prohibición total de borrado de incidentes) y limita a dbmaint al uso exclusivo de su función mantenible.

## Relaciones

- [[backend/migrations/0008_matriz_permisos.down.sql|0008 down]]: Migración reversa de este cambio
