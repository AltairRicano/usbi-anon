---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Archivo de configuración de autenticación de clientes a nivel de host para PostgreSQL correspondiente a la evidencia BD-04. Su responsabilidad es restringir el acceso a la base de datos permitiendo únicamente conexiones locales por sockets Unix e interfaces de loopback (127.0.0.1/32 y ::1/128) bajo el método trust, garantizando que no se expongan puertos de autenticación hacia redes externas.

## Enlace al reporte principal

[[pruebas/01_base_datos/BD-04_configuracion_postgres.md|BD-04 — Configuración Postgres]]
