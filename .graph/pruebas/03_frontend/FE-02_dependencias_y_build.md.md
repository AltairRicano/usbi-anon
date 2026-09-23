---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo documenta la auditoría de dependencias, tamaño del bundle de producción y el análisis forense de dos core dumps en el frontend. Determina que el paso `tsc --noEmit` revienta por falta de memoria heap en Node/V8 dentro del contenedor de 500 MB y que el build de producción `vite build` se evita ejecutándolo en el host, además de registrar vulnerabilidades moderadas en `react-router-dom` y un bundle dominado en 55% por Phaser.

## Enlaces relacionados

Estructura de frontend:
[[pruebas/03_frontend/FE-01_arbol_nomenclatura_convenciones.md|FE-01]] [[pruebas/03_frontend/FE-07_viabilidad_tauri.md|FE-07]]

Planes de prueba derivados:
[[pruebas/06_consolidado/CO-03_plan_de_pruebas_derivado.md|CO-03]]
