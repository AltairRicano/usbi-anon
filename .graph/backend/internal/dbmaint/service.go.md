---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo gestiona el mantenimiento preventivo de particiones por rango de años para las tablas level_attempts y daily_streak. Opera mediante un pool dedicado autenticado como usbi_dbmaint con permisos restringidos únicamente a EXECUTE sobre la función SECURITY DEFINER ensure_yearly_partition, manteniendo la separación DML/DDL y evitando que la partición DEFAULT absorba tráfico ordinario. Es un asunto estructural de la base de datos, independiente de los trabajos de retención legal/privacidad de [[backend/internal/maintenance/service.go.md|maintenance/service.go]] — cuyo `StartScheduler` sigue el mismo patrón de tick no fatal que el de este archivo.

## Funciones

### yearlyPartitionTargets
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula la lista pura de destinos (tabla, año) para el año actual y los dos años posteriores para cada tabla particionada configurada, sin interactuar con la base de datos.

### Service.EnsurePartitions
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Ejecuta la función de base de datos ensure_yearly_partition para cada tabla y año objetivo, garantizando la creación idempotente de particiones a través del pool usbi_dbmaint.

### StartScheduler
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicia en segundo plano un temporizador recurrente que invoca EnsurePartitions a intervalos regulares, capturando y registrando los errores de forma no fatal para reintentar en el siguiente ciclo.
