---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Archivo de evidencia que contiene la definición en PL/pgSQL del catálogo de funciones de base de datos. Su responsabilidad es establecer triggers para mantener registros acumulativos (ledgers) inmutables y controlar mutaciones por derechos ARCO o rotación de niveles sin comprometer los datos de experiencia ni auditoría.

[[pruebas/01_base_datos/BD-04_configuracion_postgres.md|Documento de análisis BD-04]]

## Funciones

### enforce_append_only_ledgers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Función trigger de PL/pgSQL que rechaza operaciones DELETE en tablas append-only y limita las operaciones UPDATE en experience_history y audit_log únicamente al borrado lógico/seudonimización de referencias (user_id, level_id, actor_account_id) para cumplir derechos ARCO o rotación de temporadas, garantizando que xp_gained y el resto de campos permanezcan inalterados.
