---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona la interfaz de administración para crear cuentas de staff, consultar el historial del cuestionario de registro por UUID, restablecer contraseñas y eliminar cuentas. En la gestión por UUID no existe un listado previo de cuentas, y las cuentas administrativas no pueden eliminarse (el backend devuelve 403 ErrCannotDeleteAdmin, reflejado directamente en la UI).

## Funciones

### AdminAccountsPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente principal para el alta de staff y la gestión individual de cuentas por UUID.

### AdminAccountsPage.handleCreate
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida el nickname (6-20 caracteres) y la longitud mínima de contraseña (8 caracteres) antes de enviar la solicitud POST /admin/accounts para crear cuentas de staff.

### AdminAccountsPage.resetManagePanels
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Restablece los estados locales de error, respuestas de cuestionario, contraseñas temporales y confirmaciones de borrado en el panel de gestión.

### AdminAccountsPage.handleViewQuizAnswers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Consulta GET /admin/accounts/{targetID}/quiz-answers para obtener las respuestas guardadas del cuestionario de registro de la cuenta especificada.

### AdminAccountsPage.handleResetPassword
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Solicita POST /admin/accounts/{targetID}/reset-password y despliega la nueva contraseña generada por única vez.

### AdminAccountsPage.handleDelete
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Ejecuta DELETE /admin/accounts/{targetID} para seudonimizar y purgar el progreso de una cuenta no administradora.

## Relaciones

- [[frontend/src/features/auth/schemas.ts.md|auth/schemas.ts]] — reutiliza `NicknameSchema` para validación de nickname
- [[frontend/src/features/admin-accounts/schemas.ts.md|admin-accounts/schemas.ts]] — valida respuestas de API para cuentas
