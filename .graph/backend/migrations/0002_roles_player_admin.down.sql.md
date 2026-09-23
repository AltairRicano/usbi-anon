---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Restaura la restricción CHECK en la columna role de la tabla accounts para permitir nuevamente los valores operator y director, además de player y admin. No revierte las degradaciones de roles realizadas previamente en la migración de ida debido a la ausencia de un historial de cambios.

## Relaciones

- [[backend/migrations/0002_roles_player_admin.up.sql|0002 up]]: Migración forward de este cambio
