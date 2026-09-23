---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona funciones de validación y sanitización para los datos de entrada en los procesos de autenticación y registro. Garantiza la seguridad de las respuestas del cuestionario evitando inyecciones de código HTML, JSON o caracteres de control, y valida el formato de nicknames y credenciales de inicio de sesión. Lo consume [[backend/internal/auth/service.go.md|service.go]]: `validateAnswerText` en [[backend/internal/auth/service.go.md#Service.RegisterAnswers|Service.RegisterAnswers]] y `validateLogin`/`normalizeNickname` en [[backend/internal/auth/service.go.md#Service.Login|Service.Login]].

## Funciones

### validateAnswerText
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida que el texto de respuesta no esté vacío, no exceda 200 caracteres, no contenga caracteres de control ASCII (< 0x20), no inicie como JSON y no contenga etiquetas HTML.

### normalizeNickname
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Normaliza el nickname convirtiéndolo a minúsculas y eliminando espacios en blanco en los extremos.

### nicknameFormatValid
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica mediante expresión regular que el nickname contenga entre 6 y 20 caracteres alfanuméricos en minúsculas.

### validateLogin
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida que los campos de nickname y contraseña en la petición de inicio de sesión no estén vacíos.
