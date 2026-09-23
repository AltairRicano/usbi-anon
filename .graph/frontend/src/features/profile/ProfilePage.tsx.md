---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Despliega el perfil y progreso oficial del usuario (XP, racha, niveles e insignias ganadas). Permite confirmar mayoría de edad (Ley 251, limitado a 3 veces por el backend) y ofrece la opción de eliminación inmediata e irreversible de la cuenta mediante DELETE /auth/me.

## Funciones

### ProfilePage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente principal para la visualización de perfil, métricas de juego y opciones de cuenta.

### ProfilePage.handleAgeUp
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Envía una solicitud POST /auth/age-up para actualizar el estatus de mayoría de edad de la cuenta en el backend y el estado local.

### ProfilePage.handleCancelAccount
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Solicita la eliminación de la cuenta propia vía DELETE /auth/me, cierra la sesión y redirige al usuario a la pantalla de login.

### Metric
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente secundario de interfaz para renderizar tarjetas de métricas numéricas del progreso del usuario.

## Relaciones

- Usa [[frontend/src/shared/components/ui/Button.tsx|Button]]
- Usa [[frontend/src/shared/components/ui/HomeButton.tsx|HomeButton]]
- Usa [[frontend/src/shared/components/ui/LinkButton.tsx|LinkButton]]
- Usa [[frontend/src/shared/apiClient.ts|apiClient]]
- Usa [[frontend/src/shared/errorMessage.ts|errorMessage]]
- Usa [[frontend/src/features/auth/useAuthStore.ts|useAuthStore]]
