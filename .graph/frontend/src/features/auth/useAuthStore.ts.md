---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Store global de Zustand para administrar el estado de autenticación y los tokens de usuario. Utiliza `sessionStorage` como mecanismo de persistencia para evitar que los JWT queden expuestos permanentemente ante posibles ataques XSS.

## Funciones

### useAuthStore.login
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Establece en el estado global los datos del usuario, el token de acceso y el token de refresco, marcando la sesión como autenticada.

### useAuthStore.updateUser
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza la información del perfil del usuario en el estado sin modificar los tokens de autenticación.

### useAuthStore.logout
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Limpia la información del usuario, restablece los tokens a nulo y marca el estado de autenticación como falso.

## Relaciones

- [[frontend/src/features/auth/schemas.ts.md|schemas.ts]] — define el tipo `User` que se almacena en el estado
