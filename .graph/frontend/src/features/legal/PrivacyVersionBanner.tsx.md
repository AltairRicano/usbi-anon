---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Banner global informativo que alerta a los usuarios autenticados si su cuenta aceptó una versión previa del aviso de privacidad. No interrumpe la navegación y permite registrar la nueva aceptación contra el servidor o descartar el aviso temporalmente.

## Funciones

### PrivacyVersionBanner
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Compara la versión del aviso de privacidad aceptada por el usuario contra la vigente para condicionar la renderización de la barra de aviso.

### PrivacyVersionBanner.handleAccept
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Envía la confirmación de aceptación a `/legal/accept`, refresca el estado del usuario de forma asíncrona y oculta el banner.

## Relaciones

- Usa [[frontend/src/shared/components/ui/Button.tsx|Button]]
- Usa [[frontend/src/shared/apiClient.ts|apiClient]]
- Usa [[frontend/src/features/auth/useAuthStore.ts|useAuthStore]]
- Usa [[frontend/src/features/legal/useMyPrivacyStatus.ts|useMyPrivacyStatus]]
