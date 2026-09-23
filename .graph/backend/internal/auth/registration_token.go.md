---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa la gestión del estado intermedio del registro mediante tokens firmados con HMAC y TTL de 10 minutos sin persistencia intermedia en base de datos. Garantiza la integridad del payload firmado sin cifrarlo, verificando la firma antes de examinar la expiración para prevenir ataques de tiempo.

## Funciones

### Service.signRegistrationToken
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Serializa el payload con los candidatos a nickname y respuestas del registro, generando un token codificado en Base64 URL-safe firmado con [[backend/internal/crypto/hash.go.md#GenerateHMAC|GenerateHMAC]]. Lo invoca [[backend/internal/auth/service.go.md#Service.RegisterAnswers|Service.RegisterAnswers]] al cerrar el segundo paso del registro.

### Service.verifyRegistrationToken
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida la firma HMAC del token de registro con [[backend/internal/crypto/hash.go.md#VerifyHMAC|VerifyHMAC]] antes de decodificar y evaluar la expiración del payload para evitar manipulación y ataques de tiempo. Lo invoca [[backend/internal/auth/service.go.md#Service.RegisterConfirm|Service.RegisterConfirm]] al recibir el nickname elegido.
