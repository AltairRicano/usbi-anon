---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona utilidades para la deserialización estricta de peticiones HTTP en formato JSON. Su responsabilidad es garantizar la integridad y seguridad de los datos de entrada prohibiendo campos no reconocidos y asegurando que la solicitud contenga exactamente un único cuerpo JSON válido. Los errores que produce se traducen a RFC 7807 con [[backend/internal/httpproblem/httpproblem.go.md#WriteDecodeProblem|WriteDecodeProblem]]; lo invocan los handlers de todo el backend (p. ej. [[backend/internal/auth/handler.go.md|auth/handler.go]], [[backend/internal/devices/handler.go.md|devices/handler.go]], [[backend/internal/badges/handler.go.md|badges/handler.go]]) al decodificar el cuerpo de cada petición.

## Funciones

### DecodeStrict
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Decodifica un cuerpo JSON desde un `io.Reader` rechazando campos desconocidos mediante `DisallowUnknownFields` y verificando que no existan valores JSON adicionales tras el primer objeto decodificado.
