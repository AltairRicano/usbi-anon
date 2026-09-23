---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa la lógica de lectura empleada durante el proceso de registro del usuario ejecutándose con el rol de base de datos de menor privilegio (`usbi_app`). Se encarga de seleccionar aleatoriamente las preguntas activas a mostrar y de validar y recuperar preguntas individuales para guardar su captura de texto al registrar las respuestas. El único consumidor es [[backend/internal/auth/service.go.md|auth/service.go]], que llama estos dos métodos directamente durante el registro en 3 pasos, sin pasar por `quiz.Handler`.

## Funciones

### PlayerService.GetActiveQuestionByID
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Busca una pregunta por su identificador único mediante [[backend/internal/repository/quiz_queries.go.md#Queries.GetRegistrationQuestionByID|Queries.GetRegistrationQuestionByID]] y verifica que permanezca activa, retornando un error de no encontrado si la pregunta no existe o está desactivada. Alimenta `POST /auth/register/answers`, manejado por [[backend/internal/auth/handler.go.md|auth/handler.go]].

### PlayerService.SelectQuestionsForRegistration
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la configuración de preguntas a mostrar ([[backend/internal/repository/quiz_queries.go.md#Queries.GetRegistrationSettings|Queries.GetRegistrationSettings]]) junto con las preguntas activas ([[backend/internal/repository/quiz_queries.go.md#Queries.ListActiveRegistrationQuestions|Queries.ListActiveRegistrationQuestions]]), ejecutando un muestreo aleatorio no ponderado para proyectar únicamente los campos públicos del cuestionario.
