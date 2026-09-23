---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa el subcomando `pgconf render` para generar archivos de configuración ajustados (`postgresql.conf` y `pg_hba.conf`) a partir de plantillas embebidas. Distribuye la memoria RAM total disponible entre los servicios para optimizar los parámetros de rendimiento de PostgreSQL según los recursos del servidor.

## Funciones

### runPgconf
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Procesa los flags de configuración como la RAM total y la red CIDR, valida los límites mínimos de memoria y renderiza las plantillas de configuración en el directorio de salida.

### computePgconfValues
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula la proporción de memoria asignada a Postgres (~35% del total) y determina valores óptimos para `shared_buffers`, `effective_cache_size`, `work_mem` y `maintenance_work_mem`.

### renderTemplate
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Lee la plantilla especificada del sistema de archivos embebido, aplica los valores calculados y escribe el resultado en el archivo de destino.

## Relaciones

- [[backend/cmd/usbictl/main.go|usbictl]]: Subcomando montado desde main
- [[backend/deploy/postgres/embed.go|pgconf]]: Empotra las plantillas `*.tmpl`
- [[backend/deploy/postgres/postgresql.conf.tmpl|postgresql.conf.tmpl]]: Plantilla renderizada
- [[backend/deploy/postgres/pg_hba.conf.tmpl|pg_hba.conf.tmpl]]: Plantilla renderizada
