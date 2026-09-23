---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo analiza la arquitectura en capas, límites de módulos y el grafo de dependencias de los 20 paquetes internos en Go y `main.go`. Confirma la separación estricta entre transporte, servicios de negocio y persistencia con sqlc (salvo la excepción de DDL en `dbmaint`), la ausencia de dependencias circulares y la presencia de funciones de consulta sin uso heredadas.

## Enlaces relacionados

Fundamentos de datos:
[[pruebas/01_base_datos/BD-01_script_unificado.sql.md|BD-01]] [[pruebas/01_base_datos/BD-06_drift_migraciones.md|BD-06]]
