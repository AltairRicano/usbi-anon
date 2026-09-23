---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Crea la función ensure_yearly_partition con privilegio SECURITY DEFINER para permitir al rol usbi_dbmaint la creación dinámica de particiones anuales en las tablas level_attempts y daily_streak. Garantiza el principio de menor privilegio al evitar otorgar permisos DDL directos a las aplicaciones HTTP, restringiendo la función a una lista blanca de tablas y a un rango de años válido (2000-2100).

## Relaciones

- [[backend/migrations/0005_particion_dbmaint.down.sql|0005 down]]: Migración reversa de este cambio
