---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Empotra las plantillas de configuración de PostgreSQL (`*.tmpl`) en el binario compilado de Go mediante la directiva `embed.FS`. Permite a las herramientas de despliegue como `usbictl` renderizar configuraciones dinámicas sin empaquetar copias literales atadas a versiones fijas del motor de base de datos.

## Relaciones

- [[backend/cmd/usbictl/pgconf.go|pgconf.go]]: Renderiza las plantillas embebidas
- [[backend/deploy/postgres/postgresql.conf.tmpl|postgresql.conf.tmpl]]: Plantilla embebida
- [[backend/deploy/postgres/pg_hba.conf.tmpl|pg_hba.conf.tmpl]]: Plantilla embebida
