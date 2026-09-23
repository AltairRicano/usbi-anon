---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo inventaría los mecanismos de almacenamiento local y verifica la ausencia de fugas de PII o huellas de dispositivo en el cliente web. Confirma que la aplicación no usa cookies ni recursos de terceros, transportando la sesión mediante Bearer JWT en `sessionStorage` (expuesto a XSS durante la sesión) y almacenando preferencias y niveles locales sin datos identificables en `localStorage`.

## Enlaces relacionados

Datos en la base de datos:
[[pruebas/01_base_datos/BD-02_diccionario_de_datos.md|BD-02]]

Endpoints del backend:
[[pruebas/02_backend/BE-02_inventario_endpoints.md|BE-02]] [[pruebas/02_backend/BE-03_vulnerabilidades_y_casos_limite.md|BE-03]] [[pruebas/02_backend/BE-04_receptividad_multicliente_tauri.md|BE-04]]

Superficie de ataque y viabilidad Tauri:
[[pruebas/03_frontend/FE-06_superficie_de_ataque_cliente.md|FE-06]] [[pruebas/03_frontend/FE-07_viabilidad_tauri.md|FE-07]]

Aspectos legales:
[[pruebas/04_legal/LG-01_inventario_y_tratamiento_de_datos.md|LG-01]]
