---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Suite de pruebas unitarias para la lógica interna del servicio de sincronización. Valida que la construcción de la cadena canónica ([[backend/internal/sync/service.go.md#CanonicalSigningPayload|service.go#CanonicalSigningPayload]]) sea independiente del orden de los elementos del payload, verifica la generación y verificación de firmas HMAC-SHA256 mediante [[backend/internal/crypto/hash.go.md|crypto.GenerateHMAC/VerifyHMAC]], y comprueba que la función [[backend/internal/sync/service.go.md#xpForAttempt|service.go#xpForAttempt]] aplique la escala de XP esperada según la dificultad e intento.
