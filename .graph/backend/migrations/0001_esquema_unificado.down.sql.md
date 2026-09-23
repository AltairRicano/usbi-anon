---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Revierte el esquema unificado de la base de datos eliminando disparadores, funciones, vistas y tablas en orden inverso de dependencia. No incluye la eliminación de la extensión pgcrypto al no haberse utilizado cifrado en el esquema baseline, destruyendo las tablas particionadas e índices asociados.

## Relaciones

- [[backend/migrations/0001_esquema_unificado.up.sql|0001 up]]: Migración forward de este cambio
