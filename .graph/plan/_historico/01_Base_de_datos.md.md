---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Documento histórico de detalle para la base de datos bajo el esquema previo de dos bases de datos independientes. Especifica las estructuras de `identities` (credenciales con `pgcrypto`) y `accounts` (ancla anónima con alias compuesto de tres enteros para evitar PII), la partición de auditorías (`identity_audit_log` vs `admin_audit_log`) y el particionado anual de tablas de progreso. Define la matriz de permisos por roles PostgreSQL (`usbi_ident_app`, `usbi_main_app`), mitigaciones contra fugas en `devices.device_kind` y los criterios de validación de la fase F1.
