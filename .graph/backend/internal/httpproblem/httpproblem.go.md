---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo centraliza la emisión de respuestas HTTP estandarizadas del backend, tanto en formato JSON para respuestas exitosas como en RFC 7807 (`application/problem+json`) para errores. Su responsabilidad es garantizar la envoltura uniforme de errores entre paquetes y abstraer la clasificación de fallos al decodificar peticiones HTTP.

## Funciones

### WriteProblem
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Emite una respuesta de error siguiendo la especificación RFC 7807 (`application/problem+json`). Configura la cabecera `Content-Type`, establece el estado HTTP y serializa un objeto [[backend/internal/domain/models.go.md|domain.ProblemDetails]] donde el campo `type` se construye concatenando el `slug` a la URI base de errores, e `instance` registra la ruta de la petición. Es el escritor de error que usan todos los handlers del backend (p. ej. [[backend/internal/auth/handler.go.md|auth/handler.go]], [[backend/internal/devices/handler.go.md|devices/handler.go]], [[backend/internal/badges/handler.go.md|badges/handler.go]], [[backend/internal/auditlog/handler.go.md|auditlog/handler.go]]).

### WriteJSON
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Serializa cualquier estructura o valor Go a JSON en el `http.ResponseWriter`, estableciendo la cabecera `Content-Type` como `application/json` y el estado HTTP recibido.

### WriteDecodeProblem
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Mapea los errores ocurridos durante la decodificación del cuerpo JSON de una petición (típicamente los que devuelve [[backend/internal/httpjson/decode.go.md#DecodeStrict|DecodeStrict]]) a respuestas RFC 7807. Si el error inspeccionado es de tipo `*http.MaxBytesError` (el cuerpo excede el límite configurado), emite un error 413 con slug `payload-too-large`; de lo contrario, responde con error 400 y slug `bad-request`.
