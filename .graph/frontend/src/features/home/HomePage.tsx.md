---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-28
---

Página de bienvenida post-autenticación. Muestra el alias o nickname y rol del usuario, permite cerrar sesión o acceder a configuración, ofrece acceso al maker local de niveles y despliega enlaces administrativos únicamente para usuarios con rol admin.

## Funciones

### HomePage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente principal del cuadro de mando inicial post-login.

## Relaciones

- Usa [[frontend/src/shared/components/ui/Button.tsx|Button]]
- Usa [[frontend/src/shared/components/ui/Card.tsx|Card]]
- Usa [[frontend/src/shared/components/SettingsEntry.tsx|SettingsEntry]]
- Usa [[frontend/src/features/auth/useAuthStore.ts|useAuthStore]]
