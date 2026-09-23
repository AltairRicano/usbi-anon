---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define la estructura y lógica de emisión y verificación de JSON Web Tokens (JWT) firmados mediante HMAC-SHA256 (HS256). Asigna los claims `domain.JWTClaims` (definidos en [[backend/internal/domain/models.go.md|models.go]]) y la expiración configurable para la autenticación de peticiones en el backend. Lo usa [[backend/internal/auth/service.go.md#Service.issueSession|Service.issueSession]] para emitir el access token en cada login/refresh.

## Funciones

### GenerateToken
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Crea y firma un JWT HS256 incluyendo los claims de usuario, rol y versión de token, además de las fechas de emisión y expiración especificadas en la configuración.

### ValidateToken
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Parsea y valida la firma y vigencia de un JWT, verificando explícitamente que utilice el algoritmo HMAC para prevenir vulnerabilidades de confusión de tipo de algoritmo y extrayendo los claims del dominio.
