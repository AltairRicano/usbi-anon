---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo evalúa la consistencia de tokens de color, el sistema de modo oscuro y el cumplimiento del contraste WCAG AA en la interfaz web. Revela que las utilidades `dark:` de Tailwind fallan al no sincronizarse con el interruptor de tema de la app por falta de `@custom-variant dark`, identifica 41 literales hexadecimales fuera de `index.css` y calcula fallos de contraste en colores por defecto de tarjetas y botones de cerrar sesión.

## Enlaces relacionados

Accesibilidad relacionada:
[[pruebas/03_frontend/FE-04_accesibilidad_daltonismo_tts.md|FE-04]]

Hallazgos consolidados:
[[pruebas/06_consolidado/CO-01_backlog_priorizado.md|CO-01]]
