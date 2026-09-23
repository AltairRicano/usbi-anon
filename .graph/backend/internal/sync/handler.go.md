---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

El archivo expone los endpoints HTTP de la API REST para la sincronización offline (`POST /api/v1/sync` y `GET /api/v1/sync/events`). Su responsabilidad es validar peticiones, comprobar que el `user_id` del payload coincida estrictamente con las claims del JWT autenticado y transformar las respuestas o errores del servicio al formato Problem Details. [[backend/internal/transport/router.go.md#SetupRoutes|transport/router.go#SetupRoutes]] monta ambas rutas bajo el grupo autenticado.

## Funciones

### Handler.SyncData
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Maneja las peticiones `POST /api/v1/sync`. Lee el cuerpo respetando límites de tamaño, decodifica el JSON e impone que el `user_id` del payload coincida con el usuario del token JWT; de lo contrario rechaza la petición con 422 Unprocessable Entity. Delega la validación HMAC y la aplicación del payload a [[backend/internal/sync/service.go.md#Service.ProcessSync|service.go#Service.ProcessSync]], traduciendo fallos de firma a 401 Unauthorized y errores de validación a 422.

### Handler.ListMyHistory
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Maneja la consulta del historial de sincronización del propio usuario (`GET /api/v1/sync/events`). Valida el token JWT y los parámetros opcionales de consulta (`device_id` en formato UUID, `cursor` RFC3339 y `page_size` entre 1 y 50), retornando una respuesta paginada mediante [[backend/internal/sync/history.go.md#Service.ListMySyncEvents|history.go#Service.ListMySyncEvents]].
