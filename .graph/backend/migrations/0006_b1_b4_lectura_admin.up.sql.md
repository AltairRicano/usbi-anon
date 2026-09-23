---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Añade soporte de esquema para consultas administrativas y de historial, imponiendo unicidad en nombres de insignias e índices optimizados para paginación de sincronizaciones e incidentes. Incorpora una regla de seguridad inmutable a nivel de base de datos mediante la función y trigger forbid_security_incident_delete, la cual prohíbe cualquier operación DELETE sobre la tabla security_incidents.

## Relaciones

- [[backend/migrations/0006_b1_b4_lectura_admin.down.sql|0006 down]]: Migración reversa de este cambio
