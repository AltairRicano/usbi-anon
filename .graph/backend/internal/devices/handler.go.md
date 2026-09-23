---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Expone los endpoints HTTP para la gestión de dispositivos de usuario, delegando en [[backend/internal/devices/service.go.md|Service]]. Extrae y valida las credenciales de autenticación JWT del contexto, decodifica las peticiones JSON con [[backend/internal/httpjson/decode.go.md#DecodeStrict|DecodeStrict]] y mapea los resultados o errores del servicio a respuestas estandarizadas bajo el formato RFC 7807 (Problem Details) con [[backend/internal/httpproblem/httpproblem.go.md#WriteProblem|WriteProblem]].

## Funciones

### Handler.RegisterDevice
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida la autenticación JWT de la petición, decodifica el cuerpo JSON de forma estricta y delega el registro en [[backend/internal/devices/service.go.md#Service.RegisterDevice|Service.RegisterDevice]]. Retorna un código HTTP 201 Created si el dispositivo es nuevo o 200 OK si se actualizó el último acceso (touch), devolviendo 422 Unprocessable Entity ante errores de validación.

### Handler.ListDevices
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica las credenciales JWT del usuario en el contexto HTTP y consulta la lista de sus dispositivos asociados con [[backend/internal/devices/service.go.md#Service.ListDevices|Service.ListDevices]], retornando la colección con un estado HTTP 200 OK.

### Handler.RevokeDevice
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene y valida el identificador UUID del dispositivo desde la ruta HTTP e invoca [[backend/internal/devices/service.go.md#Service.RevokeDevice|Service.RevokeDevice]]. Mapea la respuesta a HTTP 204 No Content en caso de éxito, o a respuestas de error 404 Not Found, 422 Unprocessable Entity o 500 Internal Server Error según el fallo devuelto.
