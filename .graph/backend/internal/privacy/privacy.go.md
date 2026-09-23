---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo gestiona la cancelación definitiva y autoservicio de cuentas de usuario mediante seudonimización y eliminación de datos. Ejecuta la baja en una transacción serializable que purga datos temporales y de progreso, anonimiza bitácoras append-only, revoca tokens y desactiva la cuenta conservando la integridad referencial.

## Funciones

### CancelAccount
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Ejecuta dentro de una transacción serializable la desvinculación de bitácoras append-only con [[backend/internal/repository/privacy_queries.go.md#Queries.NullUserInPseudonymizableLedgers|Queries.NullUserInPseudonymizableLedgers]], la purga de progreso con [[backend/internal/repository/privacy_queries.go.md#Queries.PurgeUserProgressData|Queries.PurgeUserProgressData]] y de respuestas del cuestionario con [[backend/internal/repository/privacy_queries.go.md#Queries.PurgeAccountQuizAnswers|Queries.PurgeAccountQuizAnswers]], el marcado de dispositivos para wipe local con [[backend/internal/repository/device_queries.go.md#Queries.MarkUserDevicesForWipe|Queries.MarkUserDevicesForWipe]], la revocación de tokens de refresco con [[backend/internal/repository/auth_queries.go.md#Queries.RevokeRefreshTokensForUser|Queries.RevokeRefreshTokensForUser]] y la desactivación de la cuenta con [[backend/internal/repository/privacy_queries.go.md#Queries.DeactivateAccount|Queries.DeactivateAccount]], conservando la fila seudonimizada.

La invocan tanto maintenance.Service.cancelAccount (retención legal automática) como `auth.Service.DeleteAccount` y `auth.Service.CancelSelf` (baja autoservicio en `DELETE /api/v1/auth/me`).

### randomNicknameFill
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera mediante `crypto/rand` una secuencia alfanumérica aleatoria de 20 caracteres utilizada para sobrescribir el nickname de la cuenta al ser desactivada, evitando colisiones y garantizando seudonimización irreversible.
