---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Suite de pruebas de integración end-to-end para el flujo de sincronización de eventos contra una base de datos PostgreSQL real. Cubre el ciclo completo mediante servicios reales ([[backend/internal/auth/service.go.md|auth/service.go]] para registro en 3 pasos y autenticación, [[backend/internal/devices/service.go.md|devices/service.go]] para el registro de dispositivo y [[backend/internal/sync/service.go.md#Service.ProcessSync|sync/service.go#Service.ProcessSync]] como sujeto bajo prueba), verificando el recálculo estricto de XP en el servidor, la idempotencia ante solicitudes duplicadas, el rechazo de firmas HMAC alteradas y la invalidez de intentos con fechas futuras. La base de datos aislada la provee [[backend/internal/testdb/testdb.go.md#Setup|testdb/testdb.go#Setup]].
