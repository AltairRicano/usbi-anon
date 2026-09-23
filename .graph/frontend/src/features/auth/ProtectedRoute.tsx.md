---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Componente de enrutamiento que restringe el acceso a rutas privadas verificando la autenticación y los roles del usuario. Preserva la ruta de origen en el estado de navegación al redirigir hacia `/login` en caso de no contar con sesión activa.

## Funciones

### ProtectedRoute
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inspecciona el estado de autenticación y el rol de usuario en `useAuthStore`, redirigiendo a `/login` si no está autenticado o a `/unauthorized` si el rol no coincide con los autorizados.

## Relaciones

- [[frontend/src/features/auth/useAuthStore.ts.md|useAuthStore]] — consulta estado de autenticación y rol del usuario
