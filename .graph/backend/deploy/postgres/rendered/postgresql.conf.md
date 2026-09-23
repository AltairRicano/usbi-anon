---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Versión final renderizada de `postgresql.conf` ajustada para un servidor VPS de 1024 MB de RAM compartido entre los servicios web, API y base de datos. Establece límites de uso de memoria, un máximo de 20 conexiones simultáneas y umbral de bitácora para sentencias de más de 250 ms.

## Relaciones

- [[backend/deploy/postgres/postgresql.conf.tmpl|postgresql.conf.tmpl]]: Plantilla origen renderizada por usbictl
