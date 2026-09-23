---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo cataloga el inventario completo de los 49 endpoints HTTP expuestos por el backend (46 bajo `/api/v1` y 3 de salud sin versionar). Documenta los parámetros, middlewares de autenticación JWT, verificaciones internas de rol por paquete, respuestas de error estandarizadas bajo RFC 7807 y clasifica los riesgos por exposición de datos o funciones de administración.

## Enlaces relacionados

Estructura de datos y configuración:
[[pruebas/01_base_datos/BD-01_script_unificado.sql.md|BD-01]] [[pruebas/01_base_datos/BD-02_diccionario_de_datos.md|BD-02]] [[pruebas/01_base_datos/BD-03_mapa_de_relaciones.md|BD-03]]

Seguridad de backend y frontend:
[[pruebas/02_backend/BE-03_vulnerabilidades_y_casos_limite.md|BE-03]] [[pruebas/02_backend/BE-04_receptividad_multicliente_tauri.md|BE-04]] [[pruebas/02_backend/BE-06_convenciones_y_comentarios.md|BE-06]] [[pruebas/03_frontend/FE-08_flujos_de_usuario.md|FE-08]]

Aspectos legales:
[[pruebas/04_legal/LG-01_inventario_y_tratamiento_de_datos.md|LG-01]]
