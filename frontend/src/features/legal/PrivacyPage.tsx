import { Link } from 'react-router-dom';
import { useAuthStore } from '../auth/useAuthStore';
import { usePrivacyNotice } from './usePrivacyNotice';
import { useMyPrivacyStatus } from './useMyPrivacyStatus';
import { NoticeSections } from './NoticeSections';

// Ruta pública /privacidad (M2.2): simplificado arriba, integral debajo, y —
// si hay sesión— la versión que esa cuenta aceptó y cuándo. Debe poder
// leerse sin cuenta, igual que /settings.
export default function PrivacyPage() {
  const { notice, loading, error } = usePrivacyNotice();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const { status } = useMyPrivacyStatus(isAuthenticated);

  return (
    <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
      <div className="mx-auto max-w-3xl space-y-6">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <h1 className="text-3xl font-bold">Privacidad y datos</h1>
          <Link
            to={isAuthenticated ? '/perfil' : '/login'}
            style={{ color: 'var(--color-primary)' }}
            className="hover:underline"
          >
            {isAuthenticated ? '← Volver a mi perfil' : '← Ir a iniciar sesión'}
          </Link>
        </header>

        {loading && <p className="text-sm text-[--color-muted]" aria-live="polite">Cargando aviso de privacidad…</p>}
        {error && <p className="text-sm text-[--color-error]" role="alert">{error}</p>}

        {notice && (
          <>
            <section className="space-y-3">
              <h2 className="text-xl font-semibold">Resumen</h2>
              <NoticeSections sections={notice.simplified} defaultOpenFirst />
            </section>

            <section className="space-y-3">
              <h2 className="text-xl font-semibold">Aviso completo</h2>
              <NoticeSections sections={notice.full} />
            </section>

            <p className="text-xs text-[--color-muted]">
              Versión vigente: {notice.version} · {notice.effective_date}
            </p>

            {isAuthenticated && status && (
              <p className="rounded-lg border border-[--color-border] bg-[--color-card] p-3 text-sm text-[--color-muted]">
                Tu cuenta aceptó la versión <strong>{status.privacy_notice_version}</strong>
                {status.privacy_notice_version !== status.current_privacy_notice_version && (
                  <> — hay una versión más reciente disponible arriba.</>
                )}
              </p>
            )}
          </>
        )}
      </div>
    </main>
  );
}
