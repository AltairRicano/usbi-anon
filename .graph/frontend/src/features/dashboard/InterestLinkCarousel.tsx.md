---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Renderiza un carrusel circular en abanico para explorar tarjetas de enlaces de interés por categoría. Garantiza el cumplimiento de accesibilidad de contraste (WCAG 2.2) al calcular dinámicamente el color del texto y bordes según la luminancia del color de fondo configurado por el administrador.

## Funciones

### readableTextColor
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula la luminancia relativa de un color hexadecimal en espacio de color sRGB para seleccionar texto negro o blanco con óptimo contraste.

### contrastBorderColor
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Determina el color semitransparente del borde de una tarjeta según su luminancia para asegurar la distinción visual sobre el fondo.

### InterestLinkCarousel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Gestiona la navegación circular entre tarjetas de enlaces usando operaciones de módulo para avanzar o retroceder indefinidamente.

## Relaciones

- Usa [[frontend/src/features/dashboard/interestLinksSchemas.ts|interestLinksSchemas]]
