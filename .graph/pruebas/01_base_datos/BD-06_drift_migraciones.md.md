---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo documenta las divergencias entre el esquema vivo de `usbi_anon_db` y los scripts de migración en `backend/migrations/`. Revela que el mecanismo de entrada no es incremental y se ejecuta solo sobre bases vacías, dejando la base viva congelada sin aplicar la migración 0002 (`accounts_role_check`) ni crear los roles acotados de aplicación (`usbi_app`/`usbi_migrate`).

## Enlaces relacionados

Migraciones y estructura:
[[pruebas/01_base_datos/BD-01_script_unificado.sql.md|BD-01]] [[pruebas/01_base_datos/BD-04_configuracion_postgres.md|BD-04]] [[pruebas/evidencia/bd/pg_dump_schema_only.sql.md|Evidencia pg_dump]]

Hallazgos consolidados:
[[pruebas/02_backend/BE-02_inventario_endpoints.md|BE-02]] [[pruebas/02_backend/BE-03_vulnerabilidades_y_casos_limite.md|BE-03]] [[pruebas/06_consolidado/CO-01_backlog_priorizado.md|CO-01]] [[pruebas/06_consolidado/CO-02_matriz_de_riesgos.md|CO-02]]
