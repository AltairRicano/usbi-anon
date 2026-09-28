import { useEffect, useState } from 'react';
import { apiClient } from '../../shared/apiClient';
import { errorMessage } from '../../shared/errorMessage';
import { PrivacyNoticeResponseSchema, type PrivacyNoticeResponse } from './schemas';

// Caché a nivel de módulo: el aviso vigente es contenido público e idéntico
// para cualquier visitante (RegisterPage y PrivacyPage lo consumen por
// separado); no hace sentido volver a pedirlo en cada montaje.
let cached: PrivacyNoticeResponse | null = null;

export interface UsePrivacyNoticeResult {
  notice: PrivacyNoticeResponse | null;
  loading: boolean;
  error: string | null;
}

export function usePrivacyNotice(): UsePrivacyNoticeResult {
  const [notice, setNotice] = useState<PrivacyNoticeResponse | null>(cached);
  const [loading, setLoading] = useState(!cached);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (cached) return;
    let cancelled = false;
    (async () => {
      try {
        const { data } = await apiClient.get('/legal/privacy-notice');
        const parsed = PrivacyNoticeResponseSchema.parse(data);
        cached = parsed;
        if (!cancelled) setNotice(parsed);
      } catch (err) {
        if (!cancelled) setError(errorMessage(err, 'No se pudo cargar el aviso de privacidad.'));
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  return { notice, loading, error };
}
