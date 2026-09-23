---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Simplifica el modelo de autorización del sistema degradando todas las cuentas registradas con roles operator o director a admin. Posteriormente, restringe la restricción CHECK de la tabla accounts a únicamente dos roles válidos: player y admin.

## Relaciones

- [[backend/migrations/0002_roles_player_admin.down.sql|0002 down]]: Migración reversa de este cambio
