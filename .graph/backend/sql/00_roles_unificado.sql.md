---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Script de definición SQL a nivel de clúster que crea la base de datos `usbi_anon_db` y aplica el principio de mínimo privilegio. Establece cuatro roles distintos con propósitos aislados: `usbi_app` (operaciones de jugador), `usbi_moderador` (administración), `usbi_migrate` (aplicación de esquema DDL) y `usbi_dbmaint` (mantenimiento de particiones anuales).

## Relaciones

- [[backend/cmd/usbictl/doctor.go|doctor]]: Verifica esta matriz de roles
- [[backend/cmd/usbictl/migrate.go|migrate]]: Usa el rol `usbi_migrate` definido aquí
- [[backend/migrations/0008_matriz_permisos.up.sql|0008 up]]: Otorga los GRANT según esta estructura
