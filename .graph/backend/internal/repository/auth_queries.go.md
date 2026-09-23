---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Maneja las operaciones de persistencia para tokens de refresco (`refresh_tokens`), permitiendo la emisión, revalidación, revocación y purga de sesiones. Garantiza la seguridad validando en la base de datos la fecha de expiración, el estado de revocación del token y la validez de la cuenta asociada antes de renovar los JWT. [[backend/internal/auth/service.go.md|auth/service.go]] es el consumidor principal (emisión y rotación de refresh tokens en login/refresh/logout); [[backend/internal/maintenance/service.go.md|maintenance/service.go]] purga tokens expirados en su tarea periódica, y [[backend/internal/privacy/privacy.go.md|privacy/privacy.go]] revoca las sesiones activas de una cuenta al ejercer derechos ARCO.

## Funciones

### Queries.InsertRefreshToken
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Almacena un nuevo token de refresco con su hash y fecha de expiración.

### Queries.GetRefreshTokenAccount
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la proyección mínima de la cuenta asociada a un hash de token activo, verificando que el token no esté revocado ni expirado y que la cuenta no esté eliminada.

### Queries.RevokeRefreshToken
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Revoca un token de refresco específico marcando su fecha de revocación (`revoked_at`).

### Queries.RevokeRefreshTokensForUser
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Revoca masivamente todos los tokens de refresco activos pertenecientes a un usuario.

### Queries.PurgeExpiredRefreshTokens
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina físicamente los tokens expirados o revocados hace más de 7 días para evitar el crecimiento desmedido de la tabla.
