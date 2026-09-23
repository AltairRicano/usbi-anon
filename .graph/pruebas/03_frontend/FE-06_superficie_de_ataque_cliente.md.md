---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo evalúa la superficie de ataque del cliente web y el grado de confianza entre el frontend y el backend Go. Confirma que el cliente envía `completed: true` de forma incondicional sin importar si el jugador ganó o perdió, que la doble confirmación para purgar contenido es mera fricción de UI sin respaldo en la API, y verifica con evidencia que no existen conexiones directas a PostgreSQL desde el cliente.

## Enlaces relacionados

Configuración de base de datos:
[[pruebas/01_base_datos/BD-04_configuracion_postgres.md|BD-04]] [[pruebas/01_base_datos/BD-05_credenciales_y_conexion.md|BD-05]]

Endpoints y vulnerabilidades del backend:
[[pruebas/02_backend/BE-02_inventario_endpoints.md|BE-02]] [[pruebas/02_backend/BE-03_vulnerabilidades_y_casos_limite.md|BE-03]] [[pruebas/02_backend/BE-04_receptividad_multicliente_tauri.md|BE-04]]

Privacidad del cliente y viabilidad Tauri:
[[pruebas/03_frontend/FE-05_privacidad_cliente_cookies_pii.md|FE-05]] [[pruebas/03_frontend/FE-07_viabilidad_tauri.md|FE-07]]

Aspectos legales:
[[pruebas/04_legal/LG-01_inventario_y_tratamiento_de_datos.md|LG-01]] [[pruebas/04_legal/LG-02_leyes_a_investigar.md|LG-02]]
