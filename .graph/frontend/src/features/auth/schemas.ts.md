---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define los esquemas de validación Zod para las peticiones y respuestas de autenticación y registro. Implementa en `AnswerTextSchema` reglas estrictas contra caracteres de control, estructuras tipo JSON y etiquetas HTML para prevenir ataques de inyección en el cuestionario.

## Relaciones

- [[frontend/src/features/auth/LoginPage.tsx.md|LoginPage]] — valida respuesta de login con `AuthResponseSchema`
- [[frontend/src/features/auth/RegisterPage.tsx.md|RegisterPage]] — valida los tres pasos del registro
- [[frontend/src/features/admin-accounts/AdminAccountsPage.tsx.md|AdminAccountsPage]] — reutiliza `NicknameSchema` para validación en administración de cuentas
