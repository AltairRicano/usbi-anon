---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-28
---

Formulario de configuración para el juego de Serpientes y Escaleras. Permite ajustar dimensiones del tablero, cantidad de elementos calculados según el tamaño disponible, nivel de la IA, semilla matemática y exige un mínimo de 8 preguntas con exactamente dos alternativas cada una para evitar repeticiones aceleradas en la partida.

## Funciones

### SnakeLadderForm
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente que administra la configuración del tablero de juego, parámetros de la IA y el listado de preguntas requeridas con validación de dos opciones.

### normalizeContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula y genera de forma determinista las conexiones de serpientes y escaleras sobre el tablero utilizando las dimensiones y la semilla matemática especificadas.

## Relaciones

- [[frontend/src/features/content/maker/registry.ts.md|registry.ts]] — formulario registrado en el registro de plantillas
