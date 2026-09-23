---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Elimina la función almacenada ensure_yearly_partition, utilizada por las tareas de mantenimiento de la base de datos para la creación automatizada de particiones anuales. Reasigna la gestión de particionamiento a la creación manual de tablas.

## Relaciones

- [[backend/migrations/0005_particion_dbmaint.up.sql|0005 up]]: Migración forward de este cambio
