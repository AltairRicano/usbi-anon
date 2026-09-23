---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Formulario de edición para niveles de tipo Crucigrama. Garantiza que no existan palabras respuestas duplicadas mediante normalización y utiliza la función de motor `canBuildConnectedCrossword` para validar en tiempo real que el conjunto de palabras pueda cruzarse formando una cuadrícula válida e interconectada.

## Funciones

### CrosswordForm
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente que administra la lista de palabras y pistas del crucigrama, desplegando advertencias de validación cuando las palabras se duplican o no pueden conectarse entre sí.

## Relaciones

- [[frontend/src/features/content/maker/registry.ts.md|registry.ts]] — formulario registrado en el registro de plantillas
