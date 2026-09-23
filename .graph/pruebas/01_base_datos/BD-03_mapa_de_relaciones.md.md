---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo presenta el diagrama Entidad-Relación y la narrativa detallada de los ciclos de vida de borrado en cuatro escenarios (retiro, archivo y purga de contenido, y cancelación de cuenta). Documenta la asimetría deliberada de borrado hacia `levels` (CASCADE en intentos/progreso y SET NULL en `experience_history`), garantizando que la rotación de temporadas libere almacenamiento sin despojar al jugador de su experiencia acumulada.

## Enlaces relacionados

Estructura relacionada:
[[pruebas/01_base_datos/BD-02_diccionario_de_datos.md|BD-02]] [[pruebas/01_base_datos/BD-04_configuracion_postgres.md|BD-04]] [[pruebas/01_base_datos/BD-05_credenciales_y_conexion.md|BD-05]] [[pruebas/01_base_datos/BD-06_drift_migraciones.md|BD-06]] [[pruebas/evidencia/bd/pg_dump_schema_only.sql.md|Evidencia pg_dump]]

Análisis que depende de la topología:
[[pruebas/02_backend/BE-01_arquitectura_capas_modulos.md|BE-01]] [[pruebas/02_backend/BE-06_convenciones_y_comentarios.md|BE-06]] [[pruebas/06_consolidado/CO-03_plan_de_pruebas_derivado.md|CO-03]]
