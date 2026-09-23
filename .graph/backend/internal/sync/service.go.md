---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Contiene la lógica central de negocio para el procesamiento transaccional de eventos de sincronización offline. Es responsable de verificar la firma HMAC sobre datos canónicos usando [[backend/internal/crypto/hash.go.md|crypto.VerifyHMAC/GenerateHMAC]], recalcular el XP en el servidor ignorando el valor reportado por el cliente, garantizar idempotencia ante reexpediciones del mismo evento, prevenir carreras mediante bloqueos en base de datos y actualizar rachas diarias y medallas. Invocado desde [[backend/internal/sync/handler.go.md#Handler.SyncData|handler.go#Handler.SyncData]].

## Funciones

### Service.ProcessSync
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Ejecuta el flujo de sincronización bajo una transacción con aislamiento `Serializable`. Verifica la firma HMAC del payload canónico; si la firma o los datos son inválidos, registra el evento rechazado y retorna el error correspondiente. Garantiza idempotencia ante un `sync_event_id` duplicado respondiendo `already_processed`. Bloquea intentos previos mediante [[backend/internal/repository/content_queries.go.md#Queries.LockLevelAttempt|Queries.LockLevelAttempt]] para recalcular el XP servidor según la dificultad del nivel, actualiza el progreso ([[backend/internal/repository/content_queries.go.md#Queries.UpsertPlayerProgressForAttempt|Queries.UpsertPlayerProgressForAttempt]]), registra la racha diaria y otorga las medallas obtenidas mediante [[backend/internal/repository/badge_queries.go.md#Queries.AwardEligibleBadges|Queries.AwardEligibleBadges]].

### Service.validateAndPrepare
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida y prepara los intentos de nivel y fechas de racha contenidos en la solicitud. Comprueba que las fechas no pertenezcan al futuro, que los puntajes sean no negativos y que cada nivel exista y esté publicado; ante cualquier inconsistencia retorna un error de validación específico.

### CanonicalSigningPayload
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Construye la cadena canónica determinista a partir del `SyncEventRequest` para la validación HMAC. Ordena lexicográficamente las listas de intentos de nivel, fechas de racha y medallas antes de unirlas con delimitadores, asegurando que el resultado sea independiente del orden en que el cliente envíe los arreglos.

### xpForAttempt
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula la cantidad de XP otorgada por completar un nivel según su dificultad y el número de intento. Otorga el 100% de la XP base (`4 * dificultad`) en el primer intento, 50% en los intentos 2 y 3, y 0% a partir del cuarto intento o si el nivel no se completó.
