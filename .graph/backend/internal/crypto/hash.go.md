---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona utilidades de hashing seguro de contraseñas usando Argon2id configurado según los estándares OWASP 2023 con un solo hilo (para evitar sobrecargar el servidor monohilo de producción) y funciones de generación/verificación HMAC-SHA256 en tiempo constante, además de la creación de índices ciegos (BlindIndexHMAC) para búsquedas deterministas seguras. `HashPassword`/`VerifyPassword` respaldan el registro y login de [[backend/internal/auth/service.go.md|auth/service.go]] (incluida la verificación con hash ficticio contra enumeración de usuarios); `GenerateHMAC`/`VerifyHMAC` firman y validan el token intermedio de registro en [[backend/internal/auth/registration_token.go.md|auth/registration_token.go]].

## Funciones

### HashPassword
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera un hash Argon2id codificado en formato PHC utilizando una sal de 16 bytes generada de forma aleatoria y segura con `crypto/rand`.

### VerifyPassword
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Parsea la sal y parámetros del hash PHC, verifica la versión compatible de Argon2id y compara el hash resultante con el almacenado usando tiempo constante (`subtle.ConstantTimeCompare`).

### GenerateHMAC
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula un código de autenticación de mensajes basado en hash (HMAC-SHA256) sobre un conjunto de datos utilizando la clave secreta provista.

### VerifyHMAC
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Compara una firma HMAC recibida con el HMAC calculado a partir del payload y la clave secreta mediante `hmac.Equal` en tiempo constante.

### BlindIndexHMAC
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera un HMAC-SHA256 claveado sobre datos de coincidencia exacta (como correo o teléfono) para permitir búsquedas deterministas previniendo ataques de extensión de longitud y diccionario.
