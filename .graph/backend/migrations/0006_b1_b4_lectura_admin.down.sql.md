---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Script de reversión de la migración 0006 que elimina los índices de ordenación de incidentes de seguridad y eventos de sincronización. Remueve el disparador y la función de esquema forbid_security_incident_delete que prohibía el borrado de incidentes, así como la restricción de unicidad sobre los nombres de las insignias.

## Relaciones

- [[backend/migrations/0006_b1_b4_lectura_admin.up.sql|0006 up]]: Migración forward de este cambio
