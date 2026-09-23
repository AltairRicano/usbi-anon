---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa la lógica de negocio para consultar el aviso de privacidad y registrar su aceptación. Garantiza la integridad del registro de aceptación mediante la generación de firmas HMAC calculadas a partir del ID de usuario, la fecha en UTC y una clave secreta.

## Funciones

### Service.CurrentNotice
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la versión vigente del aviso de privacidad con `legaltext.Current()` (construida por [[backend/legal/embed.go.md#mustBuildCurrent|mustBuildCurrent]], empotrada vía `go:embed`) y la transforma a [[backend/internal/legal/dto.go.md|PrivacyNoticeResponse]] para su exposición pública.

### Service.AcceptCurrent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Registra la aceptación del aviso vigente para un usuario autenticado, calculando un HMAC criptográfico con [[backend/internal/crypto/hash.go.md#GenerateHMAC|GenerateHMAC]] sobre [[backend/legal/embed.go.md#SealPayload|legaltext.SealPayload]] y guardando los datos con [[backend/internal/repository/account_queries.go.md#Queries.UpdatePrivacyAcceptance|Queries.UpdatePrivacyAcceptance]]. El mismo `legaltext.SealPayload` se usa en el registro inicial, en [[backend/internal/auth/service.go.md#Service.RegisterConfirm|Service.RegisterConfirm]].
