---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Actualiza la restricción CHECK en la tabla experience_history para diferenciar entre la experiencia verificada por el servidor (online_verified) y la experiencia reportada pero no recalculable (online_reported). Reemplaza el identificador genérico online_direct permitiendo auditabilidad según la mecánica y capacidades de verificación de cada tipo de nivel.

## Relaciones

- [[backend/migrations/0007_verificacion_resultado_nivel.down.sql|0007 down]]: Migración reversa de este cambio
