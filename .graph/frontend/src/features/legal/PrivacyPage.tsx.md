---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Página pública que expone las versiones simplificada e integral del aviso de privacidad. Si el usuario cuenta con sesión activa, muestra la versión del aviso aceptada por su cuenta y notifica si existe una actualización pendiente.

## Funciones

### PrivacyPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene los datos del aviso de privacidad y el estado del usuario para estructurar la visualización pública de las políticas y el estado de cumplimiento individual.

## Relaciones

- Usa [[frontend/src/features/auth/useAuthStore.ts|useAuthStore]]
- Usa [[frontend/src/features/legal/usePrivacyNotice.ts|usePrivacyNotice]]
- Usa [[frontend/src/features/legal/useMyPrivacyStatus.ts|useMyPrivacyStatus]]
- Usa [[frontend/src/features/legal/NoticeSections.tsx|NoticeSections]]
