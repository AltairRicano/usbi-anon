import { lazy, Suspense, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import { ProtectedRoute } from './features/auth/ProtectedRoute';
import { applyDocumentClasses, useSettingsStore } from './features/settings/useSettingsStore';

const LoginPage = lazy(() => import('./features/auth/LoginPage'));
const RegisterPage = lazy(() => import('./features/auth/RegisterPage'));
const HomePage = lazy(() => import('./features/home/HomePage'));
const AdminQuizBankPage = lazy(() => import('./features/admin-quiz-bank/AdminQuizBankPage'));
const AdminAccountsPage = lazy(() => import('./features/admin-accounts/AdminAccountsPage'));
const SettingsPage = lazy(() => import('./features/settings/SettingsPage'));
const MakerPage = lazy(() => import('./features/maker').then((mod) => ({ default: mod.MakerPage })));
const AdminContentPage = lazy(() => import('./features/content/AdminContentPage').then((mod) => ({ default: mod.AdminContentPage })));

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

          <Route
            path="/"
            element={
              <ProtectedRoute>
                <HomePage />
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
