import { useState } from 'react';
import { Button } from '../../shared/components/ui/Button';
import { apiClient } from '../../shared/apiClient';
import { useAuthStore } from '../auth/useAuthStore';
import { useMyPrivacyStatus } from './useMyPrivacyStatus';

// Banner informativo, NO bloqueante (D-06): se monta una vez en App.tsx, por
// encima de las rutas, y solo se muestra con sesión activa cuando la versión
// aceptada por la cuenta quedó atrás de la vigente.
export function PrivacyVersionBanner() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const { status, refresh } = useMyPrivacyStatus(isAuthenticated);
  const [dismissed, setDismissed] = useState(false);
  const [accepting, setAccepting] = useState(false);

  if (!status || dismissed) return null;
  if (status.privacy_notice_version === status.current_privacy_notice_version) return null;

  async function handleAccept() {
    setAccepting(true);
    try {
      await apiClient.post('/legal/accept');
      await refresh();
    } catch {
      // D-06: informativo, no bloqueante. Si falla, se descarta localmente
      // igual — la persona puede seguir usando el sistema sin interrupción.
    } finally {
      setAccepting(false);
      setDismissed(true);
    }
  }

  return (
    <div
      role="status"
      className="flex w-full flex-wrap items-center justify-between gap-3 border-b px-4 py-3 text-sm"
      style={{ borderColor: 'var(--color-border)', backgroundColor: 'var(--color-card)' }}
    >
      <span>
        El aviso de privacidad cambió de versión.{' '}
        <a href="/privacidad" target="_blank" rel="noopener noreferrer" className="underline" style={{ color: 'var(--color-primary)' }}>
          Léelo aquí
        </a>
        .
      </span>
      <div className="flex gap-2">
        <Button variant="ghost" size="sm" onClick={() => setDismissed(true)} disabled={accepting}>
          Después
        </Button>
        <Button size="sm" onClick={() => void handleAccept()} disabled={accepting}>
          {accepting ? 'Guardando…' : 'Entendido'}
        </Button>
      </div>
    </div>
  );
}
