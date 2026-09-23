---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define el plan maestro de la fase de planeación de pruebas de integración para USBI-Anon, dividiendo el trabajo en 6 fases (P0 a P6) y asignando entregables a 5 integrantes del equipo. Establece como invariantes que solo se producirán reportes de lectura sin ejecutar pruebas activas sobre el contenedor vivo, que la base de datos viva es la fuente de verdad del esquema sobre las migraciones (BD-06), y que las claves foráneas hacia levels(id) tienen comportamientos de borrado asimétricos deliberados.

## Enlaces relacionados

Documentos de preparación:
[[pruebas/00_plan/01_Delegaciones_a_agentes.md|01_Delegaciones]] [[pruebas/00_plan/02_Plantillas_y_convenciones.md|02_Plantillas]] [[pruebas/00_plan/03_Hipotesis_de_riesgo.md|03_Hipotesis]]

Entregables de consolidación final:
[[pruebas/06_consolidado/CO-03_plan_de_pruebas_derivado.md|CO-03]]
