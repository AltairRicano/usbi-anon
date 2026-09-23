---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Cliente HTTP centralizado basado en Axios para comunicarse con la API REST `/api/v1`. Inyecta tokens JWT en las solicitudes y gestiona la renovación transparente tras respuestas 401 (RFC 7807) mediante la validación de esquemas Zod, evitando el cierre de sesión no deseado en respuestas 403.

## Funciones

### apiClient.interceptors.request
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inyecta la cabecera `Authorization: Bearer <token>` en las peticiones salientes cuando existe un token de sesión activo.

### apiClient.interceptors.response
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Captura respuestas con estado 401 e intenta renovar el token mediante `/auth/refresh` reintentando la solicitud original; parsea la respuesta con Zod (`AuthResponseSchema`) y dispara el evento `auth:unauthorized` si la renovación falla.

## Relaciones

- Usa [[frontend/src/features/auth/useAuthStore.ts|useAuthStore]]
- Usa [[frontend/src/shared/schemas.ts|schemas]]
