---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Archivo DDL que define la estructura completa del esquema de PostgreSQL, incluyendo tablas, particiones por año, vistas e índices. Implementa decisiones clave de privacidad (ausencia de PII directos) y la función de trigger enforce_append_only_ledgers() que garantiza la inmutabilidad de los registros de auditoría y experiencia, permitiendo únicamente seudonimización ARCO.

Documento de análisis BD-04
