---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona el acceso a datos centralizado para la tabla de cuentas (`accounts`), gestionando la creación de identidad, credenciales y estados del jugador. Garantiza que las cuentas inactivas o borradas lógicamente (`deleted_at IS NULL`) no se consulten en logins o autenticaciones, e integra la lógica de generación aleatoria segura de alias y actualización del control de versiones de tokens para la invalidación de sesiones. Su principal consumidor es [[backend/internal/auth/service.go.md|auth/service.go]] (registro en 3 pasos, login, logout, cancelación); [[backend/internal/legal/service.go.md|legal/service.go]] también la usa para registrar la aceptación del aviso de privacidad vía `UpdatePrivacyAcceptance`, y [[backend/internal/transport/router.go.md#jwtAuthMiddleware|router.go#jwtAuthMiddleware]] llama `GetAccountByID` en cada petición autenticada para revalidar `token_version` y estado.

## Funciones

### RandomAlias
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera aleatoriamente mediante `crypto/rand` los identificadores de adjetivo, sustantivo y un número (0-999) para conformar el alias visible único al momento del registro.

### Queries.CreateAccount
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inserta la fila inicial de la cuenta con estado 'active' y versión de token por defecto, sirviendo como la única inserción durante el ciclo de vida de la cuenta.

### Queries.FindAccountByNickname
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Busca la cuenta mediante su nickname para el proceso de login, filtrando explícitamente cuentas borradas lógicamente (`deleted_at IS NULL`).

### Queries.GetAccountByID
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene los datos de la cuenta activa por su ID para revalidar el estado, rol y versión de token en cada solicitud autenticada en el middleware.

### Queries.GetAccountAlias
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Consulta la vista `account_aliases` para construir y retornar el alias legible formateado del usuario.

### Queries.TouchAccountLastLogin
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza la fecha de último inicio de sesión (`last_login_at`), permitiendo calcular de forma precisa la inactividad del jugador.

### Queries.IncrementAccountTokenVersion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Incrementa la versión del token de la cuenta, invalidando de inmediato todas las sesiones y tokens JWT emitidos previamente.

### Queries.IncrementAgeUpAttempts
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Incrementa el contador de intentos de la cuenta para certificar la mayoría de edad.

### Queries.MarkAccountAdult
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza la bandera `is_adult` a verdadero para reflejar el cambio de estado a usuario adulto.

### Queries.SetAccountPassword
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza el hash de la contraseña e incrementa atómicamente la versión del token para revocar las sesiones activas en el mismo cambio.

### Queries.UpdatePrivacyAcceptance
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Registra la versión, fecha y hash de aceptación del aviso de privacidad actualizado por el usuario.
