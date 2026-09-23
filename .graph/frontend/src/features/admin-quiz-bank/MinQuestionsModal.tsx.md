---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Componente modal de alerta de accesibilidad (WCAG 2.4.3) que constituye la única excepción deliberada a la regla sin modales del proyecto. Su función es bloquear la interfaz cuando una acción administrativa intentaría dejar menos de 4 preguntas activas, protegiendo la regla de negocio que permite el registro de usuarios.

## Funciones

### MinQuestionsModal
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Renderiza el modal de alerta cuando la propiedad `open` es verdadera, gestionando el foco y escuchando la tecla Escape para permitir el cierre accesible.

## Relaciones

- [[frontend/src/features/admin-quiz-bank/AdminQuizBankPage.tsx.md|AdminQuizBankPage]] — componente padre que lo consume
