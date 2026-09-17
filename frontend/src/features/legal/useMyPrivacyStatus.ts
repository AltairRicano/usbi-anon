import { useCallback, useEffect, useState } from 'react';
import { apiClient } from '../../shared/apiClient';
import { MyPrivacyStatusSchema, type MyPrivacyStatus } from './schemas';

// Consulta GET /auth/me solo cuando `enabled` (sesión activa) — esta pantalla
// es pública y también la visita gente sin cuenta.
export function useMyPrivacyStatus(enabled: boolean) {
  const [status, setStatus] = useState<MyPrivacyStatus | null>(null);

  const refresh = useCallback(async () => {
    if (!enabled) {
      setStatus(null);
      return;
    }
    try {
      const { data } = await apiClient.get('/auth/me');
      setStatus(MyPrivacyStatusSchema.parse(data));
    } catch {
      // Silencioso a propósito: ni PrivacyPage ni el banner de versión
      // dependen de esto para ser útiles (D-06, informativo).
    }
  }, [enabled]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return { status, refresh };
}
