---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Script de arranque para el contenedor de stress-test que orquesta PostgreSQL 15, el backend en Go y el frontend en Vite/Phaser en un entorno limitado a 500 MB de RAM. Aplica migraciones SQL iniciales, otorga privilegios explícitos al usuario usbi sobre el esquema creado por postgres, recompila componentes y gestiona la limpieza de procesos mediante trampas de señales (SIGTERM/SIGINT).

## Enlace al reporte principal

[[pruebas/01_base_datos/BD-06_drift_migraciones.md|BD-06 — Drift migraciones]]
