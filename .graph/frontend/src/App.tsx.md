---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Componente principal de la aplicación que organiza el enrutamiento mediante React Router y la carga diferida de páginas. Aplica configuraciones globales de accesibilidad en el documento y gestiona la redirección por desautorización mediante un puente de eventos.

## Funciones

### App
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Renderiza el árbol de rutas de la aplicación, aplica clases de preferencia visual en el documento HTML y configura el estado de autenticación inicial.

### AuthEventBridge
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Escucha el evento personalizado `auth:unauthorized` emitido por el cliente HTTP para redirigir automáticamente al usuario hacia la vista `/login`.

## Relaciones

- [[frontend/src/features/auth/ProtectedRoute.tsx.md|ProtectedRoute]] — componente para proteger rutas privadas
- [[frontend/src/features/auth/LoginPage.tsx.md|LoginPage]] — página de login cargada perezosamente
- [[frontend/src/features/auth/RegisterPage.tsx.md|RegisterPage]] — página de registro cargada perezosamente
- [[frontend/src/features/settings/useSettingsStore.ts.md|useSettingsStore]] — store global para preferencias de accesibilidad y temas
- [[frontend/src/features/legal/PrivacyVersionBanner.tsx.md|PrivacyVersionBanner]] — componente de aviso de privacidad
- [[frontend/src/features/admin-quiz-bank/AdminQuizBankPage.tsx.md|AdminQuizBankPage]] — página de administración del banco de preguntas
- [[frontend/src/features/admin-accounts/AdminAccountsPage.tsx.md|AdminAccountsPage]] — página de administración de cuentas
- [[frontend/src/features/admin-badges/AdminBadgesPage.tsx.md|AdminBadgesPage]] — página de administración de insignias
- [[frontend/src/features/admin-security/AdminSecurityPage.tsx.md|AdminSecurityPage]] — página de auditoría y seguridad
- [[frontend/src/features/admin-community/AdminCommunityPage.tsx.md|AdminCommunityPage]] — página de administración de comunidad
- [[frontend/src/features/content/AdminContentPage.tsx.md|AdminContentPage]] — página de administración de contenido
- [[frontend/src/features/dashboard/DashboardPage.tsx.md|DashboardPage]] — página principal del jugador
- [[frontend/src/features/content/SectionLevelsPage.tsx.md|SectionLevelsPage]] — página de niveles de una sección
- [[frontend/src/features/content/OfficialLevelPage.tsx.md|OfficialLevelPage]] — página de juego oficial
- [[frontend/src/features/content/LocalLevelPage.tsx.md|LocalLevelPage]] — página de juego local
