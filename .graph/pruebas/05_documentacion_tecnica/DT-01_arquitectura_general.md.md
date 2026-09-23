---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Sintetiza la arquitectura general de los cinco componentes del sistema USBI-Anon (web, backend, base de datos construidos; escritorio y móvil Tauri evaluados). Describe la convivencia de los tres componentes construidos dentro de un único contenedor Docker bajo un límite estricto de 500 MB de RAM y explica la viabilidad de empaquetar el frontend como aplicación nativa Tauri v2 sin reescritura de código.

## Enlaces relacionados

Estructura de datos:
[[pruebas/01_base_datos/BD-01_script_unificado.sql.md|BD-01]] [[pruebas/01_base_datos/BD-02_diccionario_de_datos.md|BD-02]] [[pruebas/01_base_datos/BD-03_mapa_de_relaciones.md|BD-03]] [[pruebas/01_base_datos/BD-05_credenciales_y_conexion.md|BD-05]]

Análisis de backend:
[[pruebas/02_backend/BE-01_arquitectura_capas_modulos.md|BE-01]] [[pruebas/02_backend/BE-04_receptividad_multicliente_tauri.md|BE-04]] [[pruebas/02_backend/BE-05_arbol_y_referencias.md|BE-05]]

Análisis de frontend:
[[pruebas/03_frontend/FE-01_arbol_nomenclatura_convenciones.md|FE-01]] [[pruebas/03_frontend/FE-07_viabilidad_tauri.md|FE-07]]

Comunicación entre componentes:
[[pruebas/05_documentacion_tecnica/DT-02_comunicacion_entre_componentes.md|DT-02]] [[pruebas/05_documentacion_tecnica/DT-03_contrato_api_v1.md|DT-03]] [[pruebas/05_documentacion_tecnica/DT-04_guia_de_onboarding.md|DT-04]]
