---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Provee la vista de administración para gestionar el banco de preguntas de registro y configurar el número de preguntas que se muestran a nuevos usuarios. Garantiza que existan al menos 4 preguntas activas capturando errores de conflicto (HTTP 409) para desplegar un modal de advertencia obligatoria.

## Funciones

### AdminQuizBankPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente reactivo principal que renderiza el panel de gestión del banco de preguntas de registro, formularios de creación/edición y ajuste de configuración.

### AdminQuizBankPage.loadQuestions
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Consulta y valida la lista de preguntas del banco de registro desde la API `/admin/registration-questions`.

### AdminQuizBankPage.loadCurrentMaxQuestionsShown
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la cantidad actual de preguntas mostradas invocando el endpoint público de registro `/auth/register/questions`.

### AdminQuizBankPage.createQuestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Envía la solicitud POST para registrar una nueva pregunta activa con el texto y orden especificados.

### AdminQuizBankPage.saveEdit
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza los datos de una pregunta existente y captura errores de conflicto HTTP 409 cuando la modificación invalida el mínimo de preguntas activas.

### AdminQuizBankPage.deleteQuestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Solicita la eliminación de una pregunta por su ID y activa la alerta modal si se incumple el mínimo de 4 preguntas activas (HTTP 409).

### AdminQuizBankPage.toggleActive
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Alterna el estado activo/inactivo de una pregunta controlando posibles respuestas de conflicto HTTP 409 del backend.

### AdminQuizBankPage.saveSettings
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida localmente que el número de preguntas a mostrar esté en el rango de 4 a 10 antes de enviar la actualización a `/admin/registration-settings`.

### isMinActiveQuestionsConflict
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Evalúa si una excepción producida durante las peticiones HTTP corresponde a una respuesta con código de estado 409.

## Relaciones

- [[frontend/src/features/admin-quiz-bank/schemas.ts.md|schemas.ts]] — valida respuestas de API con esquemas Zod
- [[frontend/src/features/admin-quiz-bank/MinQuestionsModal.tsx.md|MinQuestionsModal]] — componente modal de alerta para conflicto de preguntas mínimas
