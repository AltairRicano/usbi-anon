---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-28
---

Módulo de utilidades generales para la gestión y concatenación de clases CSS. Ofrece soporte para la combinación limpia de Tailwind CSS y la generación unificada de estilos para botones y enlaces.

## Funciones

### cn
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Unifica listas de clases CSS resolviendo duplicados y conflictos de estilos de Tailwind mediante `clsx` y `tailwind-merge`.

### buttonClasses
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera la lista concatenada de clases de Tailwind según las combinaciones de variantes visuales y tamaños definidos para botones y enlaces.

## Relaciones

- Usa clsx y tailwind-merge para procesamiento de clases CSS
- Utilizado por [[frontend/src/shared/components/ui/Button.tsx|Button]], [[frontend/src/shared/components/ui/LinkButton.tsx|LinkButton]], [[frontend/src/shared/components/ui/Card.tsx|Card]], [[frontend/src/shared/components/ui/Input.tsx|Input]], [[frontend/src/shared/components/ui/Brand.tsx|Brand]]
