---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo contiene el diccionario de datos completo de las 29 relaciones y la vista del esquema `public` en `usbi_anon_db`. Evalúa la clasificación de privacidad de cada columna (datos seudonimizados vs. PII directa) e identifica riesgos estructurales como columnas de texto libre no controladas en solicitudes ARCO y la permanencia inmutable de la IP de auditoría.

## Enlaces relacionados

Fuentes de estructura y configuración:
[[pruebas/01_base_datos/BD-04_configuracion_postgres.md|BD-04]] [[pruebas/01_base_datos/BD-05_credenciales_y_conexion.md|BD-05]] [[pruebas/01_base_datos/BD-06_drift_migraciones.md|BD-06]] [[pruebas/evidencia/bd/pg_dump_schema_only.sql.md|Evidencia pg_dump]]

Análisis de seguridad y legal que dependen de este inventario:
[[pruebas/02_backend/BE-03_vulnerabilidades_y_casos_limite.md|BE-03]] [[pruebas/02_backend/BE-06_convenciones_y_comentarios.md|BE-06]] [[pruebas/04_legal/LG-01_inventario_y_tratamiento_de_datos.md|LG-01]]
