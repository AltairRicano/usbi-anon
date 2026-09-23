---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Crea las tablas para gestionar la sección de enlaces de interés externos agrupados por categorías y el buzón de sugerencias anónimo. Aplica validaciones estrictas por CHECK en formato de URL (http/https) y color hexadecimal, mientras que la tabla de sugerencias desacopla completamente cualquier ID de usuario reteniendo solo instantáneas de progreso acumulado.

## Relaciones

- [[backend/migrations/0003_enlaces_interes_y_sugerencias.down.sql|0003 down]]: Migración reversa de este cambio
