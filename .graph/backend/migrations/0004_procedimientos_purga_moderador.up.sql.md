---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define procedimientos con privilegio SECURITY DEFINER pertenecientes al rol usbi_moderador para permitir la eliminación de respuestas del cuestionario y la seudonimización de bitácoras durante la cancelación de cuenta. Otorga permiso de ejecución al rol usbi_app, previniendo que dicho rol requiera permisos directos de UPDATE o DELETE sobre tablas sensibles.

## Relaciones

- [[backend/migrations/0004_procedimientos_purga_moderador.down.sql|0004 down]]: Migración reversa de este cambio
