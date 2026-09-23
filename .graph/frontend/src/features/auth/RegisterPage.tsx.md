---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Componente de registro anónimo por pasos que evita solicitar datos personales. Genera nicknames a partir de un cuestionario de gustos, exige la aceptación obligatoria del aviso de privacidad y muestra las credenciales emitidas por única vez para su resguardo.

## Funciones

### RegisterPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Orquesta la vista del flujo de registro en 3 pasos (cuestionario, elección de nickname y credenciales emitidas) controlando el estado del formulario.

### RegisterPage.handleAnswersSubmit
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida localmente que las respuestas no violen las reglas anti-inyección, verifica la aceptación del aviso de privacidad y envía los datos a `/auth/register/answers` para avanzar al paso 2.

### RegisterPage.handleConfirmSubmit
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Confirma la elección del nickname enviando el token temporal a `/auth/register/confirm` para recibir la contraseña definitiva e incrementar el paso al resumen de credenciales.

### RegisterPage.handleCopyCredentials
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Copia el nickname y contraseña generados al portapapeles utilizando la API del navegador.

### StepIndicator
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Renderiza la barra indicadora del progreso del registro marcando visualmente el paso activo y los completados.

## Relaciones

- [[frontend/src/features/auth/schemas.ts.md|schemas.ts]] — valida respuestas del registro en los tres pasos
- [[frontend/src/features/legal/usePrivacyNotice.ts.md|usePrivacyNotice]] — recupera aviso de privacidad actual
- [[frontend/src/features/legal/PrivacyNoticeInline.tsx.md|PrivacyNoticeInline]] — componente para mostrar el aviso
