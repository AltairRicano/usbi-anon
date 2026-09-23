---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Script Bash de inicialización ejecutado únicamente al arrancar un contenedor PostgreSQL con un volumen de datos vacío. Lee las variables de entorno de contraseñas y las transfiere sin comillas redundantes a `psql` para ejecutar `00_roles_unificado.sql.src` y aprovisionar los roles del clúster de forma segura.

## Relaciones

- [[backend/sql/00_roles_unificado.sql|00_roles_unificado.sql]]: Aplicado después en el bootstrap
