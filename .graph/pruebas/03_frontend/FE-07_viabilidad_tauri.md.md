---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo analiza la viabilidad técnica de empaquetar el frontend con Tauri v2 para escritorio y dispositivos móviles. Determina que el proyecto es empaquetable mediante reempaquetado sin reescritura, pero señala que el origen `tauri://` rompe las rutas relativas `/api/v1` de `apiClient`, que la sesión en `sessionStorage` requiere migrar a memoria y que la exportación de archivos mediante Blob/HTMLAnchor no funciona en WebViews móviles.

## Enlaces relacionados

Receptividad multicliente del backend:
[[pruebas/02_backend/BE-04_receptividad_multicliente_tauri.md|BE-04]]
