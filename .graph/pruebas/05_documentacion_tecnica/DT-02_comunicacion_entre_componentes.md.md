---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Detalla los protocolos de comunicación, formatos de mensaje, mecanismos de autenticación y diagramas de secuencia entre componentes. Documenta que la sesión usa exclusivamente JWT Bearer y tokens de refresco opacos sin cookies, evidenciando la vulnerabilidad donde el frontend envía completed: true de forma incondicional al servidor sin que este revalide si el nivel se resolvió realmente.

## Enlaces relacionados

Configuración de red y credenciales:
[[pruebas/01_base_datos/BD-05_credenciales_y_conexion.md|BD-05]]

Análisis de endpoints y vulnerabilidades:
[[pruebas/02_backend/BE-02_inventario_endpoints.md|BE-02]] [[pruebas/02_backend/BE-03_vulnerabilidades_y_casos_limite.md|BE-03]] [[pruebas/02_backend/BE-04_receptividad_multicliente_tauri.md|BE-04]]

Privacidad del cliente y flujos:
[[pruebas/03_frontend/FE-05_privacidad_cliente_cookies_pii.md|FE-05]] [[pruebas/03_frontend/FE-06_superficie_de_ataque_cliente.md|FE-06]] [[pruebas/03_frontend/FE-08_flujos_de_usuario.md|FE-08]]

Contrato de API y onboarding:
[[pruebas/05_documentacion_tecnica/DT-03_contrato_api_v1.md|DT-03]] [[pruebas/05_documentacion_tecnica/DT-04_guia_de_onboarding.md|DT-04]]
