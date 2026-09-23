---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define la estructura baseline de la base de datos unificada sin PII, implementando identidades anónimas con alias estructurados por vocabulario cerrado, cuestionario de gustos y soporte offline. Garantiza la conservación de la experiencia acumulada tras el retiro de niveles mediante agregación en account_retired_progress y referencias nulas en experience_history, además de aplicar la invariante append-only sobre bitácoras y registros de experiencia vía trigger. Implementa las tres direcciones de FK de la regla de rotación de niveles sin pérdida de XP — ver [[Estado_Proyecto/Decisiones.md#rotacion_niveles_tres_fk_distintas|Decisiones]].

## Relaciones

- [[backend/migrations/0001_esquema_unificado.down.sql|0001 down]]: Migración reversa de este cambio
