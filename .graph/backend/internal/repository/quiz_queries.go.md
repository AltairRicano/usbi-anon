---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona el acceso a datos para las preguntas de registro, su configuración y las respuestas congeladas. Garantiza mediante bloqueos explícitos `FOR UPDATE` en subconsultas la regla de negocio que impide reducir el banco a menos de 4 preguntas activas durante transacciones concurrentes. Las consultas de lectura pública (`GetRegistrationQuestionByID`, `GetRegistrationSettings`, `ListActiveRegistrationQuestions`) las usa [[backend/internal/quiz/player_service.go.md|quiz/player_service.go]] durante el registro; el resto (CRUD del banco, con los bloqueos `FOR UPDATE`) lo expone [[backend/internal/quiz/admin_service.go.md|quiz/admin_service.go]] en `/admin/registration-questions`. `InsertAccountQuizAnswer` y `ListAccountQuizAnswers` las usa [[backend/internal/auth/service.go.md|auth/service.go]] al congelar y auditar las respuestas del registro.

## Funciones

### Queries.ListRegistrationQuestions
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Recupera todas las preguntas del banco de registro, tanto activas como inactivas, para su administración.

### Queries.ListActiveRegistrationQuestions
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene las preguntas de registro activas ordenadas para el muestreo aleatorio presentado al usuario durante el registro.

### Queries.CountActiveRegistrationQuestions
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Cuenta las preguntas activas aplicando un bloqueo `FOR UPDATE` en subconsulta para evitar condiciones de carrera al validar el mínimo requerido.

### Queries.CreateRegistrationQuestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inserta una nueva pregunta en el banco de preguntas de registro con su texto, estado y orden de presentación.

### Queries.UpdateRegistrationQuestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza el contenido, estado de activación y orden de una pregunta de registro existente.

### Queries.GetRegistrationQuestionForUpdate
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Lee una pregunta de registro bajo un bloqueo `FOR UPDATE` para verificar su estado en transacciones de modificación o borrado.

### Queries.DeleteRegistrationQuestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina una pregunta de registro del sistema por su identificador.

### Queries.GetRegistrationQuestionByID
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene los datos de una pregunta de registro por su identificador sin aplicar bloqueos para lectura durante el flujo de registro.

### Queries.GetRegistrationSettings
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la configuración global del registro de usuarios, incluyendo el límite de preguntas mostradas.

### Queries.UpdateRegistrationSettings
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza el número máximo de preguntas que se presentan al usuario en el proceso de registro.

### Queries.InsertAccountQuizAnswer
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inserta la respuesta dada a una pregunta durante el registro congelando una copia del texto de la pregunta dentro de la transacción de creación de cuenta.

### Queries.ListAccountQuizAnswers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene las respuestas a cuestionarios congeladas asociadas a una cuenta para su consulta auditada en el panel de administración.
