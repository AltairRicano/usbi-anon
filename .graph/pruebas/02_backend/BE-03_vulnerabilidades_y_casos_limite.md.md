---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo analiza la postura de seguridad y casos límite del backend frente a un modelo de cliente hostil como aplicaciones Tauri modificadas. Identifica vulnerabilidades críticas como la confianza ciega en la bandera de completitud enviada por el cliente (`completed: true`), la ineficiencia del rate limit por IP tras Cloudflare Tunnel sin cabeceras confiables, la falta de bloqueo de cuenta en login y listados no paginados.

## Enlaces relacionados

Estructura y configuración de datos:
[[pruebas/01_base_datos/BD-01_script_unificado.sql.md|BD-01]] [[pruebas/01_base_datos/BD-04_configuracion_postgres.md|BD-04]] [[pruebas/01_base_datos/BD-05_credenciales_y_conexion.md|BD-05]]

Endpoints e infraestructura de backend:
[[pruebas/02_backend/BE-02_inventario_endpoints.md|BE-02]] [[pruebas/02_backend/BE-04_receptividad_multicliente_tauri.md|BE-04]]

Vulnerabilidades en frontend y legal:
[[pruebas/03_frontend/FE-02_dependencias_y_build.md|FE-02]] [[pruebas/03_frontend/FE-05_privacidad_cliente_cookies_pii.md|FE-05]] [[pruebas/03_frontend/FE-06_superficie_de_ataque_cliente.md|FE-06]] [[pruebas/03_frontend/FE-07_viabilidad_tauri.md|FE-07]] [[pruebas/04_legal/LG-01_inventario_y_tratamiento_de_datos.md|LG-01]]

Planes de prueba derivados:
[[pruebas/06_consolidado/CO-03_plan_de_pruebas_derivado.md|CO-03]]
