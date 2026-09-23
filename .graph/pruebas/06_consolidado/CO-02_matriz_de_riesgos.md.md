---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Sintetiza los hallazgos en 14 riesgos de nivel matriz (probabilidad × impacto), enfocándose en determinar qué elementos bloquean el despliegue en el servidor de la UV. Identifica 6 riesgos bloqueantes para la migración a la UV (credenciales por defecto, autenticación trust en Postgres, falta de roles acotados, secretos en texto plano, topología de 500 MB con fallo en cascada y licitud de menores autorreportados).

## Enlaces relacionados

Base de datos:
[[pruebas/01_base_datos/BD-04_configuracion_postgres.md|BD-04]] [[pruebas/01_base_datos/BD-05_credenciales_y_conexion.md|BD-05]] [[pruebas/01_base_datos/BD-06_drift_migraciones.md|BD-06]]

Backend:
[[pruebas/02_backend/BE-01_arquitectura_capas_modulos.md|BE-01]] [[pruebas/02_backend/BE-02_inventario_endpoints.md|BE-02]] [[pruebas/02_backend/BE-03_vulnerabilidades_y_casos_limite.md|BE-03]] [[pruebas/02_backend/BE-04_receptividad_multicliente_tauri.md|BE-04]] [[pruebas/02_backend/BE-05_arbol_y_referencias.md|BE-05]] [[pruebas/02_backend/BE-06_convenciones_y_comentarios.md|BE-06]]

Frontend:
[[pruebas/03_frontend/FE-01_arbol_nomenclatura_convenciones.md|FE-01]] [[pruebas/03_frontend/FE-02_dependencias_y_build.md|FE-02]] [[pruebas/03_frontend/FE-03_color_tokens_y_consistencia.md|FE-03]] [[pruebas/03_frontend/FE-04_accesibilidad_daltonismo_tts.md|FE-04]] [[pruebas/03_frontend/FE-05_privacidad_cliente_cookies_pii.md|FE-05]] [[pruebas/03_frontend/FE-06_superficie_de_ataque_cliente.md|FE-06]] [[pruebas/03_frontend/FE-07_viabilidad_tauri.md|FE-07]] [[pruebas/03_frontend/FE-08_flujos_de_usuario.md|FE-08]]

Legal:
[[pruebas/04_legal/LG-01_inventario_y_tratamiento_de_datos.md|LG-01]] [[pruebas/04_legal/LG-03_seleccion_aviso_privacidad_uv.md|LG-03]]

Consolidación:
[[pruebas/06_consolidado/CO-01_backlog_priorizado.md|CO-01]]
