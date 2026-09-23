---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Plantilla de overrides para `postgresql.conf` que se monta e incluye en la configuración principal mediante `include_dir`. Ajusta parámetros de memoria (`shared_buffers`, `work_mem`) y límites de conexión calculados según la RAM del VPS, mantiene SSL desactivado a nivel de base de datos (terminado en Nginx) y habilita registros para depurar consultas lentas.

## Relaciones

- [[backend/deploy/postgres/embed.go|embed.go]]: Empotra esta plantilla
- [[backend/cmd/usbictl/pgconf.go|pgconf.go]]: La renderiza
- [[backend/deploy/postgres/rendered/postgresql.conf|postgresql.conf]]: Versión renderizada
