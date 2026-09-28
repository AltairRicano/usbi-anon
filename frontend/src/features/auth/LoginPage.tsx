import { useState, type FormEvent } from 'react';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import { Button } from '../../shared/components/ui/Button';
import { Input } from '../../shared/components/ui/Input';
import { UsbiEmblem, Spinner } from '../../shared/components/ui/Brand';
import { SettingsEntry } from '../../shared/components/SettingsEntry';
import { useAuthStore } from './useAuthStore';
import { apiClient } from '../../shared/apiClient';
import { errorMessage } from '../../shared/errorMessage';
import { AuthResponseSchema } from '../../shared/schemas';
import { registerCurrentDevice } from '../offline-processes/deviceIdentity';
import { ZodError } from 'zod';

const EyeIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
    <circle cx="12" cy="12" r="3" />
  </svg>
);
const EyeOffIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24" />
    <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68" />
    <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61" />
    <line x1="2" x2="22" y1="2" y2="22" />
  </svg>
);

export default function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const login = useAuthStore((s) => s.login);

  const [nickname, setNickname] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice] = useState<string | null>(() => (location.state as { message?: string } | null)?.message ?? null);
  const [loading, setLoading] = useState(false);

  const from = (location.state as { from?: { pathname: string } })?.from?.pathname ?? '/';

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      const response = await apiClient.post('/auth/login', { nickname, password });
      const data = AuthResponseSchema.parse(response.data);
      login(data.user, data.access_token, data.refresh_token);
      void registerCurrentDevice();
      navigate(from, { replace: true });
    } catch (err: unknown) {
      if (err instanceof ZodError) {
        console.error('[LoginPage] Respuesta de /auth/login con forma inesperada:', err.issues);
        setError('El servidor respondió de forma inesperada. Intenta de nuevo más tarde.');
      } else {
        setError(errorMessage(err, 'No se pudo conectar con el servidor.'));
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen flex items-center justify-center p-4" style={{ backgroundColor: 'var(--color-surface)' }}>
      <div className="w-full max-w-md overflow-hidden rounded-2xl shadow-xl" style={{ backgroundColor: 'var(--color-card)', color: 'var(--color-text-card)' }}>
        <div aria-hidden="true" className="h-1.5 w-full" style={{ background: 'linear-gradient(90deg, var(--color-primary), var(--color-secondary))' }} />

        <div className="space-y-6 p-8">
          <div className="flex justify-end">
            <SettingsEntry />
          </div>

          <header className="flex flex-col items-center gap-3 text-center">
            <UsbiEmblem />
            <div className="space-y-0.5">
              <h1 className="text-2xl font-bold" style={{ color: 'var(--color-primary)' }}>USBI</h1>
              <p className="text-sm" style={{ color: 'var(--color-muted)' }}>Universidad Veracruzana</p>
            </div>
          </header>

          {notice && (
            <p
              className="rounded-lg border p-3 text-sm"
              style={{ borderColor: 'var(--color-border)', color: 'var(--color-primary)', backgroundColor: 'color-mix(in srgb, var(--color-primary) 8%, transparent)' }}
            >
              {notice}
            </p>
          )}

          <form onSubmit={handleSubmit} className="space-y-4" aria-label="Formulario de inicio de sesión" noValidate>
            <Input
              id="login-nickname"
              label="Nickname"
              type="text"
              autoComplete="username"
              placeholder="jaguar42x7"
              value={nickname}
              onChange={(e) => setNickname(e.currentTarget.value)}
              required
              aria-required="true"
            />

            <Input
              id="login-password"
              label="Contraseña"
              type={showPassword ? 'text' : 'password'}
              autoComplete="current-password"
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.currentTarget.value)}
              required
              aria-required="true"
              rightIcon={
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  aria-label={showPassword ? 'Ocultar contraseña' : 'Mostrar contraseña'}
                  className="p-1 hover:text-[var(--color-foreground)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--color-primary)] rounded"
                >
                  {showPassword ? <EyeOffIcon /> : <EyeIcon />}
                </button>
              }
            />

            {error && (
              <p
                role="alert"
                aria-live="assertive"
                className="flex items-start gap-2 rounded-lg border p-3 text-sm"
                style={{ borderColor: 'var(--color-error)', color: 'var(--color-error)', backgroundColor: 'color-mix(in srgb, var(--color-error) 8%, transparent)' }}
              >
                <span aria-hidden="true">⚠</span>
                <span>{error}</span>
              </p>
            )}

            <Button type="submit" size="lg" className="w-full font-bold" disabled={loading} aria-busy={loading}>
              {loading && <Spinner />}
              {loading ? 'Iniciando sesión…' : 'Iniciar sesión'}
            </Button>
          </form>

          <div className="text-center text-sm">
            <Link to="/register" style={{ color: 'var(--color-primary)' }} className="inline-flex items-center justify-center py-2 hover:underline">
              ¿No tienes cuenta? Regístrate
            </Link>
          </div>

          <div className="text-center text-xs">
            <Link to="/privacidad" style={{ color: 'var(--color-muted)' }} className="inline-flex items-center justify-center py-2 hover:underline">
              Aviso de privacidad
            </Link>
          </div>
        </div>
      </div>
    </main>
  );
}
