---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona utilidades para la generación, saneamiento y diseño dinámico de tarjetas en el juego de memoria. Garantiza asignaciones de color únicas para cada pareja mediante distribución HSL, sanea entradas de datos arbitrarias y calcula ratios de contraste WCAG para asegurar la legibilidad del texto sobre cualquier fondo.

Utilizado por: [[frontend/packages/engine/src/games/MemoryEngine.ts.md|MemoryEngine]]. Importa de: [[frontend/packages/schema/index.ts.md|@usbi/schema]] (tipos de datos).

## Funciones

### createMemoryPairs
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera una lista de parejas de memoria iniciales con IDs secuenciales, contenidos vacíos y colores calculados mediante distribución HSL.

### createMemoryColorOptions
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Crea un listado de códigos de color en formato hexadecimal para ser utilizados como opciones de paleta visual.

### normalizeMemoryPairs
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Sanea y valida estructuras no confiables de parejas de memoria, asignando identificadores por defecto y resolviendo colisiones para asegurar colores hexadecimales únicos por pareja.

### filterPlayableMemoryPairs
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Filtra y retorna únicamente las parejas de memoria que poseen contenido textual no vacío en ambas tarjetas.

### normalizeMemoryBackColor
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida si un valor de color de fondo es un código hexadecimal válido y lo retorna normalizado, utilizando un color por defecto en caso de invalidez.

### getMemoryReadableTextColor
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Evalúa el contraste WCAG de un color de fondo frente a blanco y texto oscuro, retornando el color de texto con mayor legibilidad.

### getMemoryBackCardStyle
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera el objeto de estilo CSS para la parte posterior de la tarjeta integrando gradientes decorativos y el color de texto accesible.

### getMemoryFrontCardStyle
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Construye la definición de estilos CSS para la cara frontal de la tarjeta aplicando un degradado lineal y color de texto legible.
