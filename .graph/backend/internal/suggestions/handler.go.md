---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Expone los endpoints HTTP para el envío de sugerencias y la gestión administrativa del buzón mediante Chi. Aplica control de acceso basado en roles exigiendo perfil de administrador para consultar y eliminar sugerencias, mientras permite el envío a cualquier usuario autenticado. [[backend/internal/transport/router.go.md#SetupRoutes|transport/router.go#SetupRoutes]] monta estas tres rutas (`POST /suggestions`, `GET /admin/suggestions`, `DELETE /admin/suggestions/{suggestion_id}`) bajo el grupo autenticado.

## Funciones

### Handler.Submit
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Maneja peticiones POST /suggestions decodificando de forma estricta la estructura de la sugerencia e inspeccionando las credenciales del contexto. Delega el registro a [[backend/internal/suggestions/player_service.go.md#PlayerService.Submit|player_service.go#PlayerService.Submit]] y retorna las respuestas estructuradas o errores según la especificación RFC 7807.

### Handler.List
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Procesa consultas GET /admin/suggestions comprobando que el usuario tenga rol de administrador. Valida y convierte los parámetros de consulta cursor y page_size (limitado entre 1 y 50) antes de obtener los datos mediante [[backend/internal/suggestions/admin_service.go.md#AdminService.List|admin_service.go#AdminService.List]].

### Handler.Delete
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Atiende solicitudes DELETE /admin/suggestions/{suggestion_id} restringidas a administradores. Valida el formato UUID del identificador proporcionado en la URL y ejecuta el borrado mediante [[backend/internal/suggestions/admin_service.go.md#AdminService.Delete|admin_service.go#AdminService.Delete]], respondiendo con un estado HTTP 204 No Content en caso de éxito.
