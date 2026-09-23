---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Controlador HTTP encargado de manejar las solicitudes del módulo legal. Permite la consulta pública del aviso de privacidad sin autenticación y gestiona el registro de aceptación exigiendo la presencia de claims JWT válidos en el contexto HTTP.

## Funciones

### Handler.GetPrivacyNotice
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Maneja las peticiones GET para obtener el aviso de privacidad vigente delegando en [[backend/internal/legal/service.go.md#Service.CurrentNotice|Service.CurrentNotice]], entregando la respuesta en formato JSON de forma pública y sin requerir sesión.

### Handler.Accept
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Maneja las peticiones POST para registrar la aceptación del aviso de privacidad, validando la autenticación del usuario mediante JWT en el contexto y delegando la persistencia a [[backend/internal/legal/service.go.md#Service.AcceptCurrent|Service.AcceptCurrent]].
