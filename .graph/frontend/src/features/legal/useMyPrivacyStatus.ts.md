---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Hook que consulta a `/auth/me` el estado de la versión del aviso de privacidad aceptada por el usuario. La petición únicamente se ejecuta si hay una sesión activa (`enabled`) y maneja cualquier error de forma silenciosa al ser una comprobación meramente informativa.

## Funciones

### useMyPrivacyStatus
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Consulta de forma condicional la información de privacidad del usuario autenticado y provee una función `refresh` para actualizar el estado tras una interacción.

## Relaciones

- Usa [[frontend/src/shared/apiClient.ts|apiClient]]
- Usa [[frontend/src/features/legal/schemas.ts|schemas]]
