---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Elimina las tablas suggestions, interest_links e interest_link_categories junto con sus respectivos índices en orden inverso a sus relaciones de clave foránea. Sirve como reversión para el módulo de enlaces externos y buzón de sugerencias de la sección Más.

## Relaciones

- [[backend/migrations/0003_enlaces_interes_y_sugerencias.up.sql|0003 up]]: Migración forward de este cambio
