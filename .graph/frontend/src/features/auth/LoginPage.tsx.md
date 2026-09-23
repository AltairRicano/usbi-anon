---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Componente de vista para el inicio de sesión de usuarios mediante nickname y contraseña. Valida la respuesta del servidor contra un esquema Zod, registra el dispositivo en almacenamiento local para soporte offline y actualiza la tienda de autenticación antes de redirigir al usuario.

## Funciones

### LoginPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Renderiza la interfaz de inicio de sesión gestionando los estados del formulario, la visibilidad de la contraseña y los mensajes de error.

### LoginPage.handleSubmit
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Envía las credenciales a `/auth/login`, valida el token y los datos devueltos mediante `AuthResponseSchema`, registra la identidad del dispositivo en almacenamiento offline y redirige a la ruta previa o principal.

## Relaciones

- [[frontend/src/features/auth/useAuthStore.ts.md|useAuthStore]] — store de estado para tokens y datos del usuario
- [[frontend/src/features/auth/schemas.ts.md|schemas.ts]] — validación de respuesta de login mediante `AuthResponseSchema`
