---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Contiene los manejadores HTTP del módulo de autenticación y administración de cuentas, delegando cada endpoint al [[backend/internal/auth/service.go.md|Service]] homónimo. Se encarga de la decodificación estricta de peticiones JSON con [[backend/internal/httpjson/decode.go.md#DecodeStrict|DecodeStrict]] y el mapeo de errores del dominio a respuestas RFC 7807 Problem Details con [[backend/internal/httpproblem/httpproblem.go.md#WriteProblem|WriteProblem]]/[[backend/internal/httpproblem/httpproblem.go.md#WriteJSON|WriteJSON]]. Aplica decisiones de seguridad como unificar respuestas de credenciales inválidas para evitar enumeración de usuarios.

## Funciones

### Handler.RegisterQuestions
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende `GET /auth/register/questions`, obteniendo de [[backend/internal/auth/service.go.md#Service.RegisterQuestions|Service.RegisterQuestions]] el conjunto de preguntas activas para iniciar el registro.

### Handler.RegisterAnswers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende `POST /auth/register/answers`, decodificando las respuestas enviadas y delegando a [[backend/internal/auth/service.go.md#Service.RegisterAnswers|Service.RegisterAnswers]], gestionando errores de validación, versión de privacidad obsoleta (409) o saturación del servicio (429).

### Handler.RegisterConfirm
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende `POST /auth/register/confirm`, delegando a [[backend/internal/auth/service.go.md#Service.RegisterConfirm|Service.RegisterConfirm]] y manejando errores de token expirado (410), token inválido (400) o conflicto de nickname (409).

### Handler.Login
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende `POST /auth/login`, delegando a [[backend/internal/auth/service.go.md#Service.Login|Service.Login]] y retornando un mensaje de error 401 unificado tanto para usuario no encontrado como contraseña incorrecta para mitigar la enumeración de usuarios.

### Handler.Refresh
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende `POST /auth/refresh`, delegando a [[backend/internal/auth/service.go.md#Service.Refresh|Service.Refresh]] para renovar la sesión con un nuevo token de acceso a partir de un token de refresco válido.

### Handler.Logout
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende `POST /auth/logout`, delegando a [[backend/internal/auth/service.go.md#Service.Logout|Service.Logout]] para revocar los tokens de refresco y descartar las sesiones activas del usuario autenticado.

### Handler.Me
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende `GET /auth/me`, delegando a [[backend/internal/auth/service.go.md#Service.Me|Service.Me]] para devolver el estado y la versión del aviso de privacidad de la cuenta autenticada.

### Handler.AgeUp
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende `POST /auth/age-up`, delegando a [[backend/internal/auth/service.go.md#Service.AgeUp|Service.AgeUp]] enviando la dirección IP y User-Agent del cliente.

### Handler.CancelSelf
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende `DELETE /auth/me`, delegando a [[backend/internal/auth/service.go.md#Service.CancelSelf|Service.CancelSelf]] para ejecutar la cancelación autoservicio de la cuenta del usuario autenticado.

### Handler.CreateAdminAccount
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende `POST /admin/accounts`, delegando a [[backend/internal/auth/service.go.md#Service.CreateAdminAccount|Service.CreateAdminAccount]] para permitir a administradores crear cuentas de staff con rol y credenciales explícitas.

### Handler.DeleteAdminAccount
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende `DELETE /admin/accounts/{account_id}`, delegando a [[backend/internal/auth/service.go.md#Service.DeleteAccount|Service.DeleteAccount]], que impide estrictamente que un administrador elimine a otro administrador (403).

### Handler.GetAccountQuizAnswers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende `GET /admin/accounts/{account_id}/quiz-answers`, delegando a [[backend/internal/auth/service.go.md#Service.GetAccountQuizAnswers|Service.GetAccountQuizAnswers]] para que administradores consulten las respuestas registradas por un usuario.

### Handler.ResetAccountPassword
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende `POST /admin/accounts/{account_id}/reset-password`, delegando a [[backend/internal/auth/service.go.md#Service.ResetAccountPassword|Service.ResetAccountPassword]] para el reseteo administrativo de la contraseña de una cuenta objetivo.
