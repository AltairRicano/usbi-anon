---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Revierte la restricción CHECK de la tabla experience_history al estado original que solo aceptaba el valor online_direct para fuentes en línea. Debido a los triggers de solo-anexar (append-only), esta reversión únicamente puede ejecutarse limpiamente si no existen registros que utilicen los nuevos métodos de verificación introducidos en la migración 0007.

## Relaciones

- [[backend/migrations/0007_verificacion_resultado_nivel.up.sql|0007 up]]: Migración forward de este cambio
