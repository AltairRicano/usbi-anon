---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Formulario React para la creación y edición de sopas de letras en el maker de contenido. Gestiona las dimensiones del tablero (5 a 24), la generación aleatoria de semillas y la administración de palabras, aplicando una regla de sanitización que fuerza la conversión a mayúsculas y la eliminación de caracteres que no sean letras de la A a la Z.

## Funciones

### WordSearchForm
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente de formulario que maneja la configuración del tablero y la lista de palabras, aplicando sanitización a mayúsculas A-Z en cada edición de palabra.

## Relaciones

- [[frontend/src/features/content/maker/registry.ts.md|registry.ts]] — formulario registrado en el registro de plantillas
