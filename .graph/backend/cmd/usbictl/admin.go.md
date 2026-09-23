---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa la lógica del subcomando `admin create` en la herramienta `usbictl` para la creación e inicialización de cuentas con rol administrativo o de jugador. Al utilizar directamente el servicio `auth.Service`, asegura que las reglas de validación de nicknames y contraseñas sean idénticas a las aplicadas en la API HTTP.

## Funciones

### runAdmin
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Parsea las opciones de la CLI, establece conexión a la base de datos omitiendo la URL en errores para prevenir la fuga de credenciales, e invoca el servicio de autenticación representando a un actor administrativo sintético.

### adminBootstrapPassword
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Recupera la contraseña desde la variable de entorno `ADMIN_BOOTSTRAP_PASSWORD` o, si no está presente, la solicita interactivamente por consola ocultando los caracteres ingresados y verificando que no esté vacía.

## Relaciones

- [[backend/cmd/usbictl/main.go|usbictl]]: Subcomando montado desde main
- [[backend/internal/auth/service.go|auth.Service]]: Comparte CreateAdminAccount con la ruta HTTP
- [[backend/cmd/hash_password/main.go|hash_password]]: Herramienta anterior, ahora integrada en auth.Service
