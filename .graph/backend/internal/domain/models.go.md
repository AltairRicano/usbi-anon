---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define los tipos de datos, enums y DTOs principales del dominio para usuarios, JWT, sincronización offline y respuestas RFC 7807. Garantiza la privacidad del sistema al excluir cualquier dato personal (PII) o material criptográfico del DTO `User` y `SyncPayload`, e impone la invariante de seguridad de que el XP enviado por clientes offline no es confiable y debe recalcularse obligatoriamente en el servidor. `JWTClaims` es generado y validado por [[backend/internal/crypto/jwt.go.md#GenerateToken|GenerateToken]]/[[backend/internal/crypto/jwt.go.md#ValidateToken|ValidateToken]] y viaja en el contexto de cada petición bajo `ClaimsKey`, leído por los handlers de todo el backend (p. ej. [[backend/internal/auth/handler.go.md|auth/handler.go]], [[backend/internal/auditlog/handler.go.md|auditlog/handler.go]]); `User` es el DTO que arma [[backend/internal/auth/service.go.md#Service.issueSession|Service.issueSession]] en cada login; `ProblemDetails` es serializado por [[backend/internal/httpproblem/httpproblem.go.md#WriteProblem|WriteProblem]] en cada respuesta de error.
