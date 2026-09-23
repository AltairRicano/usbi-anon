---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo contiene el DDL completo para reconstruir desde cero el esquema vivo de `usbi_anon_db` en PostgreSQL 15. Define tablas, particiones, claves foráneas con reglas estrictas de borrado (CASCADE/SET NULL), restricciones CHECK e índices parciales, además de la función e invariante `enforce_append_only_ledgers()` que protege la inmutabilidad de `audit_log` y `experience_history`. Documenta fielmente el estado real en producción reproduciendo defectos de drift sin corregirlos.

## Enlaces relacionados

Análisis del esquema:
[[pruebas/01_base_datos/BD-02_diccionario_de_datos.md|BD-02]] [[pruebas/01_base_datos/BD-03_mapa_de_relaciones.md|BD-03]] [[pruebas/01_base_datos/BD-04_configuracion_postgres.md|BD-04]] [[pruebas/01_base_datos/BD-05_credenciales_y_conexion.md|BD-05]] [[pruebas/01_base_datos/BD-06_drift_migraciones.md|BD-06]] [[pruebas/evidencia/bd/grants_funciones.txt.md|Evidencia grants_funciones]]

Análisis de backend:
[[pruebas/02_backend/BE-01_arquitectura_capas_modulos.md|BE-01]] [[pruebas/02_backend/BE-05_arbol_y_referencias.md|BE-05]]

Documentación técnica:
[[pruebas/05_documentacion_tecnica/DT-04_guia_de_onboarding.md|DT-04]]
