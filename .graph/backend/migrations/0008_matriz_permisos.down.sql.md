---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Revierte la matriz de permisos otorgada en la migración 0008 a un estado neutro sin privilegios sobre las tablas para los roles usbi_app, usbi_moderador y usbi_dbmaint. Restablece los permisos globales por defecto sobre el esquema public, dejando la base de datos sin permisos asignados previamente a la migración.

## Relaciones

- [[backend/migrations/0008_matriz_permisos.up.sql|0008 up]]: Migración forward de este cambio
