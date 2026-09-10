import { lazy, Suspense, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import { ProtectedRoute } from './features/auth/ProtectedRoute';
import { applyDocumentClasses, useSettingsStore } from './features/settings/useSettingsStore';

const LoginPage = lazy(() => import('./features/auth/LoginPage'));
const RegisterPage = lazy(() => import('./features/auth/RegisterPage'));
const AdminQuizBankPage = lazy(() => import('./features/admin-quiz-bank/AdminQuizBankPage'));
const AdminAccountsPage = lazy(() => import('./features/admin-accounts/AdminAccountsPage'));
const SettingsPage = lazy(() => import('./features/settings/SettingsPage'));
const MakerPage = lazy(() => import('./features/maker').then((mod) => ({ default: mod.MakerPage })));
const AdminContentPage = lazy(() => import('./features/content/AdminContentPage').then((mod) => ({ default: mod.AdminContentPage })));

// F10.10 — vista de jugador
const DashboardPage = lazy(() => import('./features/dashboard/DashboardPage'));
const SectionLevelsPage = lazy(() => import('./features/content/SectionLevelsPage').then((mod) => ({ default: mod.SectionLevelsPage })));
const OfficialLevelPage = lazy(() => import('./features/content/OfficialLevelPage').then((mod) => ({ default: mod.OfficialLevelPage })));
const LocalLevelPage = lazy(() => import('./features/content/LocalLevelPage').then((mod) => ({ default: mod.LocalLevelPage })));
const ProfilePage = lazy(() => import('./features/profile/ProfilePage'));

// Bloque C / auditoría frontend↔backend (estado_proyecto.md 2026-09-10)
const AdminBadgesPage = lazy(() => import('./features/admin-badges/AdminBadgesPage'));
const AdminSecurityPage = lazy(() => import('./features/admin-security/AdminSecurityPage'));
const AdminCommunityPage = lazy(() => import('./features/admin-community/AdminCommunityPage'));
const OfflineProcessesPage = lazy(() => import('./features/offline-processes/OfflineProcessesPage'));

export default function App() {
  const { theme, colorBlindFilter, reduceMotion, textScale } = useSettingsStore();

  useEffect(() => {
    applyDocumentClasses({ theme, colorBlindFilter, reduceMotion, textScale });
  }, [theme, colorBlindFilter, reduceMotion, textScale]);

  return (
    <BrowserRouter>
      <AuthEventBridge />
      <Suspense fallback={<RouteFallback />}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          {/* Pública a propósito: alguien con baja visión o daltonismo debe
              poder ajustar apariencia/accesibilidad ANTES de tener cuenta. */}
          <Route path="/settings" element={<SettingsPage />} />

          {/* / → dashboard; HomePage fue una landing temporal de F10 que ya no hace falta. */}
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <DashboardPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/dashboard"
            element={<Navigate to="/" replace />}
          />

          {/* F10.10 — secciones, niveles oficial/local, perfil */}
          <Route
            path="/sections/:sectionId"
            element={
              <ProtectedRoute>
                <SectionLevelsPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/levels/:levelId/play"
            element={
              <ProtectedRoute>
                <OfficialLevelPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/local-levels/:levelId/play"
            element={
              <ProtectedRoute>
                <LocalLevelPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/perfil"
            element={
              <ProtectedRoute>
                <ProfilePage />
              </ProtectedRoute>
            }
          />

          {/* Sin restricción de rol a propósito: el maker local guarda en
              localStorage, no toca la base de datos ni necesita sección —
              cualquier persona con sesión puede probar una idea de nivel
              (plan/05_Contenido_maker_y_juego.md §5). */}
          <Route
            path="/maker"
            element={
              <ProtectedRoute>
                <MakerPage />
              </ProtectedRoute>
            }
          />

          <Route
            path="/admin/registration-questions"
            element={
              <ProtectedRoute allowedRoles={['admin']}>
                <AdminQuizBankPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/admin/accounts"
            element={
              <ProtectedRoute allowedRoles={['admin']}>
                <AdminAccountsPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/admin/content"
            element={
              <ProtectedRoute allowedRoles={['admin']}>
                <AdminContentPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/admin/insignias"
            element={
              <ProtectedRoute allowedRoles={['admin']}>
                <AdminBadgesPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/admin/seguridad"
            element={
              <ProtectedRoute allowedRoles={['admin']}>
                <AdminSecurityPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/admin/comunidad"
            element={
              <ProtectedRoute allowedRoles={['admin']}>
                <AdminCommunityPage />
              </ProtectedRoute>
            }
          />

          {/* Sin restricción de rol a propósito (D5, estado_proyecto.md
              2026-09-10): un admin que jugó sin conexión ve sus propios
              dispositivos igual que un jugador. */}
          <Route
            path="/procesos-offline"
            element={
              <ProtectedRoute>
                <OfflineProcessesPage />
              </ProtectedRoute>
            }
          />

          <Route
            path="/unauthorized"
            element={
              <main className="min-h-screen flex items-center justify-center">
                <div className="text-center space-y-4">
                  <h1 style={{ color: 'var(--color-error)' }}>Acceso no autorizado</h1>
                  <p style={{ color: 'var(--color-muted)' }}>No tienes permiso para ver esta página.</p>
                </div>
              </main>
            }
          />

          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  );
}

function RouteFallback() {
  return (
    <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
      <p className="text-[--color-muted]">Cargando...</p>
    </main>
  );
}

function AuthEventBridge() {
  const navigate = useNavigate();

  useEffect(() => {
    const handleUnauthorized = () => navigate('/login', { replace: true });
    window.addEventListener('auth:unauthorized', handleUnauthorized);
    return () => window.removeEventListener('auth:unauthorized', handleUnauthorized);
  }, [navigate]);

  return null;
}
