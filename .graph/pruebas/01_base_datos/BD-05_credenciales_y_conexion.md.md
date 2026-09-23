---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo cataloga las credenciales activas, secretos del backend (`HMAC_SECRET`, `JWT_SECRET`) y datos del administrador sembrado en el contenedor `usbi-anon`. Identifica riesgos de seguridad como contraseñas de base de datos por defecto (`usbi/usbi`), falta de autenticación con contraseña en superusuario debido al modo `trust`, y exposición de llaves criptográficas en variables de entorno legibles por proceso.

## Enlaces relacionados

Configuración relacionada:
[[pruebas/01_base_datos/BD-01_script_unificado.sql.md|BD-01]] [[pruebas/01_base_datos/BD-04_configuracion_postgres.md|BD-04]]

Análisis de seguridad e infraestructura que dependen de alcance de red:
[[pruebas/02_backend/BE-03_vulnerabilidades_y_casos_limite.md|BE-03]] [[pruebas/02_backend/BE-04_receptividad_multicliente_tauri.md|BE-04]] [[pruebas/05_documentacion_tecnica/DT-02_comunicacion_entre_componentes.md|DT-02]]
