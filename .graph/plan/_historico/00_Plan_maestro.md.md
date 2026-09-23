---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Archivo histórico que documenta el diseño original de migración de USBI a USBI-Anon basado en un modelo de dos bases de datos separadas (`usbi_ident_db` con email cifrado y `usbi_anon_db` de progreso) unidas por UUID. Explica la justificación de la tabla ancla local `accounts`, las fases F0–F4 y la saga ARCO de cancelación reanudable con checkpoints de estado. Contiene advertencias explícitas de obsolescencia respecto a la posterior unificación en una sola base sin correo electrónico tras las decisiones de la fase F5.
