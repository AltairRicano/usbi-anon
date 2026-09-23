---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Plantilla parametrizada para generar el archivo `pg_hba.conf` de PostgreSQL forzando la autenticación mediante `scram-sha-256`. Restringe la superficie de red aceptando únicamente conexiones locales, localhost y la subred interna CIDR del contenedor Docker Compose, evitando exponer el puerto públicamente.

## Relaciones

- [[backend/deploy/postgres/embed.go|embed.go]]: Empotra esta plantilla
- [[backend/cmd/usbictl/pgconf.go|pgconf.go]]: La renderiza
- [[backend/deploy/postgres/rendered/pg_hba.conf|pg_hba.conf]]: Versión renderizada
