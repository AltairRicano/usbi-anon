---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo analiza la configuración efectiva de `postgresql.conf` y `pg_hba.conf` dentro del contenedor `usbi-anon`. Documenta hallazgos clave como el uso del método `trust` sin contraseña en conexiones locales y loopback, la falta de ajuste de parámetros de memoria para el límite de 500 MB del contenedor, la ausencia de logging de consultas y la mitigación parcial de exposición de puerto por reglas de red.

## Enlaces relacionados

Estructura y credenciales documentadas:
[[pruebas/01_base_datos/BD-01_script_unificado.sql.md|BD-01]] [[pruebas/01_base_datos/BD-02_diccionario_de_datos.md|BD-02]] [[pruebas/01_base_datos/BD-05_credenciales_y_conexion.md|BD-05]] [[pruebas/evidencia/bd/BD-04_show_params.txt.md|Evidencia show_params]]

Análisis de seguridad que dependen de la configuración:
[[pruebas/02_backend/BE-03_vulnerabilidades_y_casos_limite.md|BE-03]] [[pruebas/02_backend/BE-06_convenciones_y_comentarios.md|BE-06]] [[pruebas/03_frontend/FE-02_dependencias_y_build.md|FE-02]] [[pruebas/06_consolidado/CO-03_plan_de_pruebas_derivado.md|CO-03]]
