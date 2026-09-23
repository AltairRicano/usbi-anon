---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa la lógica central de negocio para autenticación, registro en 3 pasos, ciclo de vida de cuentas y administración. Utiliza un semáforo de concurrencia para limitar operaciones pesadas de hash Argon2id y mitiga ataques de canal lateral en el login mediante la verificación de hashes ficticios cuando el usuario no existe.

## Funciones

### NewService
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicializa el servicio verificando que los secretos HMAC y JWT no estén vacíos, y configurando el canal semáforo para limitar el hashing concurrente de contraseñas. Recibe un [[backend/internal/quiz/player_service.go.md|quiz.PlayerService]] — nunca el AdminService del banco de preguntas — porque auth solo necesita muestrear preguntas activas durante el registro.

### Service.RegisterQuestions
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene el conjunto aleatorio de preguntas activas para el inicio del registro delegando la selección a [[backend/internal/quiz/player_service.go.md#PlayerService.SelectQuestionsForRegistration|PlayerService.SelectQuestionsForRegistration]]. La invoca [[backend/internal/auth/handler.go.md#Handler.RegisterQuestions|Handler.RegisterQuestions]].

### Service.RegisterAnswers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida las respuestas enviadas contra [[backend/internal/auth/validation.go.md#validateAnswerText|validateAnswerText]], comprueba con [[backend/legal/embed.go.md#VerifyVersion|legaltext.VerifyVersion]] que la versión del aviso de privacidad sea la vigente, resuelve cada pregunta con [[backend/internal/quiz/player_service.go.md#PlayerService.GetActiveQuestionByID|PlayerService.GetActiveQuestionByID]], genera 4 candidatos a nickname con [[backend/internal/quiz/credentials.go.md#GenerateNicknameCandidates|quiz.GenerateNicknameCandidates]] y emite un token de registro firmado de 10 minutos con [[backend/internal/auth/registration_token.go.md#Service.signRegistrationToken|Service.signRegistrationToken]]. La invoca [[backend/internal/auth/handler.go.md#Handler.RegisterAnswers|Handler.RegisterAnswers]].

### Service.RegisterConfirm
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida el token de registro con [[backend/internal/auth/registration_token.go.md#Service.verifyRegistrationToken|Service.verifyRegistrationToken]] y el nickname elegido, verifica la ausencia de colisiones, genera la contraseña con [[backend/internal/quiz/credentials.go.md#GeneratePassword|quiz.GeneratePassword]] y la hashea con [[backend/internal/crypto/hash.go.md#HashPassword|HashPassword]], y persiste la cuenta con [[backend/internal/repository/account_queries.go.md#Queries.CreateAccount|Queries.CreateAccount]], las respuestas y el alias dentro de una transacción serializable, dejando rastro con [[backend/internal/audit/audit.go.md#Log|audit.Log]]. La invoca [[backend/internal/auth/handler.go.md#Handler.RegisterConfirm|Handler.RegisterConfirm]].

### Service.Login
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Autentica credenciales buscando la cuenta con [[backend/internal/repository/account_queries.go.md#Queries.FindAccountByNickname|Queries.FindAccountByNickname]] y verificando el password con [[backend/internal/crypto/hash.go.md#VerifyPassword|VerifyPassword]], ejecutando una verificación con hash ficticio en caso de usuario inexistente para mantener constante el tiempo de cómputo de Argon2id y prevenir enumeración. La invoca [[backend/internal/auth/handler.go.md#Handler.Login|Handler.Login]].

### Service.Refresh
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida un token de refresco via HMAC hash contra [[backend/internal/repository/auth_queries.go.md#Queries.GetRefreshTokenAccount|Queries.GetRefreshTokenAccount]], revoca el token utilizado con [[backend/internal/repository/auth_queries.go.md#Queries.RevokeRefreshToken|Queries.RevokeRefreshToken]] y emite un nuevo par de tokens de acceso y refresco si la cuenta permanece activa. La invoca [[backend/internal/auth/handler.go.md#Handler.Refresh|Handler.Refresh]].

### Service.Logout
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Incrementa la versión del token de la cuenta con [[backend/internal/repository/account_queries.go.md#Queries.IncrementAccountTokenVersion|Queries.IncrementAccountTokenVersion]] e invalida todos sus tokens de refresco activos con [[backend/internal/repository/auth_queries.go.md#Queries.RevokeRefreshTokensForUser|Queries.RevokeRefreshTokensForUser]]. La invoca [[backend/internal/auth/handler.go.md#Handler.Logout|Handler.Logout]].

### Service.Me
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Recupera los detalles de la cuenta con [[backend/internal/repository/account_queries.go.md#Queries.GetAccountByID|Queries.GetAccountByID]] y expone la versión del aviso de privacidad aceptado junto con `legaltext.CurrentVersion` ([[backend/legal/embed.go.md|embed.go]]). La invoca [[backend/internal/auth/handler.go.md#Handler.Me|Handler.Me]].

### Service.AgeUp
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Registra el intento de mayoría de edad (máximo 3) con [[backend/internal/repository/account_queries.go.md#Queries.IncrementAgeUpAttempts|Queries.IncrementAgeUpAttempts]] y actualiza el estado de la cuenta con [[backend/internal/repository/account_queries.go.md#Queries.MarkAccountAdult|Queries.MarkAccountAdult]] en una transacción auditada con [[backend/internal/audit/audit.go.md#Log|audit.Log]]. La invoca [[backend/internal/auth/handler.go.md#Handler.AgeUp|Handler.AgeUp]].

### Service.CancelSelf
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Ejecuta la cancelación autoservicio e inhabilitación inmediata de la cuenta solicitante delegando a [[backend/internal/privacy/privacy.go.md#CancelAccount|privacy.CancelAccount]]. La invoca [[backend/internal/auth/handler.go.md#Handler.CancelSelf|Handler.CancelSelf]].

### Service.CreateAdminAccount
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Permite a un administrador crear cuentas de staff con rol y credenciales explícitas sin pasar por el cuestionario de registro, persistiendo con [[backend/internal/repository/account_queries.go.md#Queries.CreateAccount|Queries.CreateAccount]]. La invoca [[backend/internal/auth/handler.go.md#Handler.CreateAdminAccount|Handler.CreateAdminAccount]].

### Service.DeleteAccount
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Ejecuta la cancelación administrativa de una cuenta delegando en [[backend/internal/privacy/privacy.go.md#CancelAccount|privacy.CancelAccount]], impidiendo estrictamente que un administrador elimine a otro administrador. La invoca [[backend/internal/auth/handler.go.md#Handler.DeleteAdminAccount|Handler.DeleteAdminAccount]].

### Service.GetAccountQuizAnswers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene las respuestas del cuestionario de una cuenta con `Queries.ListAccountQuizAnswers` para revisión administrativa, registrando el acceso en la bitácora de auditoría con [[backend/internal/audit/audit.go.md#Log|audit.Log]]. La invoca [[backend/internal/auth/handler.go.md#Handler.GetAccountQuizAnswers|Handler.GetAccountQuizAnswers]].

### Service.ResetAccountPassword
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera una nueva contraseña aleatoria de 16 caracteres para una cuenta objetivo, la hashea con [[backend/internal/crypto/hash.go.md#HashPassword|HashPassword]] y actualiza su hash con [[backend/internal/repository/account_queries.go.md#Queries.SetAccountPassword|Queries.SetAccountPassword]]. La invoca [[backend/internal/auth/handler.go.md#Handler.ResetAccountPassword|Handler.ResetAccountPassword]].

### Service.acquirePasswordHashSlot
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Adquiere un cupo en el canal semáforo para limitar las operaciones concurrentes de hash de contraseña, devolviendo error de servicio ocupado si el límite se excede.

### Service.issueSession
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera el par de tokens de acceso y refresco con [[backend/internal/crypto/jwt.go.md#GenerateToken|GenerateToken]], registra la fecha del último inicio de sesión con [[backend/internal/repository/account_queries.go.md#Queries.TouchAccountLastLogin|Queries.TouchAccountLastLogin]] y construye el `domain.User` de la respuesta (definido en [[backend/internal/domain/models.go.md|models.go]]). Lo comparten [[backend/internal/auth/service.go.md#Service.Login|Service.Login]] y [[backend/internal/auth/service.go.md#Service.Refresh|Service.Refresh]].
