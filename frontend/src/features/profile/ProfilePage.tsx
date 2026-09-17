import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '../../shared/components/ui/Button';
import { HomeButton } from '../../shared/components/ui/HomeButton';
import { LinkButton } from '../../shared/components/ui/LinkButton';
import { apiClient } from '../../shared/apiClient';
import { errorMessage } from '../../shared/errorMessage';
import { useAuthStore } from '../auth/useAuthStore';
import type { ProfileProgressResponse } from '../content/types';
import { templateTypeLabel } from '../content/types';
import { ProfileProgressResponseSchema } from '../content/schemas';

export default function ProfilePage() {
  const user = useAuthStore((s) => s.user);
  const updateUser = useAuthStore((s) => s.updateUser);
  const logout = useAuthStore((s) => s.logout);
  const navigate = useNavigate();

  const [progress, setProgress] = useState<ProfileProgressResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  const [ageUpMessage, setAgeUpMessage] = useState<string | null>(null);
  const [ageUpError, setAgeUpError] = useState<string | null>(null);
  const [ageUpLoading, setAgeUpLoading] = useState(false);

  const [confirmingCancel, setConfirmingCancel] = useState(false);
  const [cancelError, setCancelError] = useState<string | null>(null);
  const [cancelLoading, setCancelLoading] = useState(false);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const { data } = await apiClient.get('/profile/progress');
        if (!cancelled) setProgress(ProfileProgressResponseSchema.parse(data));
      } catch {
        if (!cancelled) setError('No se pudo cargar el progreso.');
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  async function handleAgeUp() {
    if (!user) return;
    setAgeUpLoading(true);
    setAgeUpError(null);
    setAgeUpMessage(null);
    try {
      await apiClient.post('/auth/age-up');
      updateUser({ ...user, is_adult: true });
      setAgeUpMessage('Estatus actualizado: tu cuenta ya está marcada como mayor de edad.');
    } catch (err) {
      setAgeUpError(errorMessage(err, 'No se pudo actualizar el estatus de mayoría de edad.'));
    } finally {
      setAgeUpLoading(false);
    }
  }

  async function handleCancelAccount() {
    setCancelLoading(true);
    setCancelError(null);
    try {
      await apiClient.delete('/auth/me');
      logout();
      navigate('/login', { replace: true });
    } catch (err) {
      setCancelError(errorMessage(err, 'No se pudo eliminar la cuenta.'));
      setCancelLoading(false);
    }
  }

  return (
    <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
      {confirmingCancel && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
          <div className="max-w-md rounded-2xl bg-[--color-card] p-6 shadow-2xl border border-[--color-border] space-y-4">
            <h2 className="text-xl font-bold" style={{ color: 'var(--color-error)' }}>¿Eliminar tu cuenta?</h2>
            <p className="text-sm">
              Esta acción es inmediata e irreversible: se cancela tu cuenta y se cierra tu sesión en todos
              los dispositivos. No hace falta aprobación de nadie más.
            </p>
            {cancelError && <p className="text-sm" style={{ color: 'var(--color-error)' }}>{cancelError}</p>}
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setConfirmingCancel(false)} disabled={cancelLoading}>Cancelar</Button>
              <Button variant="danger" onClick={() => void handleCancelAccount()} disabled={cancelLoading}>
                {cancelLoading ? 'Eliminando…' : 'Eliminar mi cuenta'}
              </Button>
            </div>
          </div>
        </div>
      )}

      <div className="mx-auto max-w-5xl space-y-6">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">Perfil y Progreso</h1>
            <p className="text-sm text-[--color-muted]">Avance oficial calculado por el backend.</p>
          </div>
          <HomeButton />
        </header>

        {error && <p className="rounded border border-[--color-error] bg-[--color-card] p-3 text-[--color-error]">{error}</p>}
        {ageUpError && <p className="rounded border border-[--color-error] bg-[--color-card] p-3 text-[--color-error]">{ageUpError}</p>}
        {ageUpMessage && <p className="rounded border border-[--color-border] bg-[--color-card] p-3 text-[--color-primary]">{ageUpMessage}</p>}

        {user && !user.is_adult && (
          <section className="rounded-lg bg-[--color-card] p-5 shadow-sm">
            <h2 className="text-xl font-semibold">Confirmación de mayoría de edad</h2>
            <p className="mt-2 text-sm text-[--color-muted]">
              Al confirmar, el backend actualiza tu estatus a mayor de edad (Ley 251). El límite es de 3 confirmaciones por cuenta.
            </p>
            <Button className="mt-4" onClick={() => void handleAgeUp()} disabled={ageUpLoading}>
              {ageUpLoading ? 'Actualizando...' : 'Confirmar mayoría de edad'}
            </Button>
          </section>
        )}

        {progress && (
          <>
            <section className="grid gap-4 md:grid-cols-4">
              <Metric label="XP total" value={progress.total_xp} />
              <Metric label="Niveles completados" value={progress.completed_levels} />
              <Metric label="Intentos" value={progress.total_attempts} />
              <Metric label="Racha" value={progress.current_streak} />
            </section>

            <section className="rounded-lg bg-[--color-card] p-5 shadow-sm">
              <h2 className="mb-4 text-xl font-semibold">Insignias</h2>
              <div className="mb-6 flex flex-wrap gap-2">
                {(progress.badges ?? []).map((badge) => (
                  <span
                    key={badge.id}
                    className="flex items-center gap-2 rounded-full border border-[--color-border] px-3 py-1 text-sm"
                    title={`Ganada el ${new Date(badge.earned_at).toLocaleDateString()}`}
                  >
                    <span aria-hidden="true">{badge.icon_key}</span>
                    {badge.name}
                    <span className="text-xs text-[--color-muted]">{badge.xp_threshold} XP</span>
                  </span>
                ))}
                {(progress.badges ?? []).length === 0 && <p className="text-sm text-[--color-muted]">Aún no hay insignias.</p>}
              </div>

              <div className="mb-6 flex flex-wrap gap-4">
                <div className="rounded-lg border border-[--color-border] p-5 shadow-sm bg-[--color-card] flex flex-col justify-between max-w-xs flex-1">
                  <div>
                    <h3 className="text-lg font-bold mb-2">Configuración</h3>
                    <p className="text-sm text-[--color-muted] mb-4">Filtro de daltonización, tema, tamaño de texto, movimiento y sonido.</p>
                  </div>
                  <Button variant="outline" onClick={() => navigate('/settings')} className="w-full mt-4">
                    Ir a Configuración
                  </Button>
                </div>

                <div className="rounded-lg border border-[--color-border] p-5 shadow-sm bg-[--color-card] flex flex-col justify-between max-w-xs flex-1">
                  <div>
                    <h3 className="text-lg font-bold mb-2">Privacidad y datos</h3>
                    <p className="text-sm text-[--color-muted] mb-4">Qué datos usa el sistema y qué versión del aviso aceptó tu cuenta.</p>
                  </div>
                  <LinkButton to="/privacidad" variant="outline" size="md" className="w-full mt-4">
                    Ver aviso de privacidad
                  </LinkButton>
                </div>
              </div>

              <h2 className="mb-4 text-xl font-semibold">Niveles jugados</h2>
              <div className="divide-y">
                {progress.levels.map((level) => (
                  <div key={level.level_id} className="flex flex-wrap items-center justify-between gap-3 py-3">
                    <div>
                      <p className="font-semibold">{level.title}</p>
                      <p className="text-sm text-[--color-muted]">
                        {templateTypeLabel(level.template_type)} · XP {level.xp_total_for_level} · intentos {level.attempts_count}
                      </p>
                    </div>
                    <LinkButton to={`/levels/${level.level_id}/play`}>Jugar de nuevo</LinkButton>
                  </div>
                ))}
                {progress.levels.length === 0 && <p className="py-4 text-sm text-[--color-muted]">Aún no hay progreso oficial.</p>}
              </div>
            </section>
          </>
        )}

        <section className="rounded-lg border p-5 shadow-sm" style={{ borderColor: 'var(--color-error)' }}>
          <h2 className="text-xl font-semibold" style={{ color: 'var(--color-error)' }}>Zona de peligro</h2>
          <p className="mt-2 text-sm text-[--color-muted]">
            Eliminar tu cuenta es inmediato e irreversible: no requiere aprobación de nadie más.
          </p>
          <Button variant="danger" className="mt-4" onClick={() => setConfirmingCancel(true)}>
            Eliminar mi cuenta
          </Button>
        </section>
      </div>
    </main>
  );
}

function Metric({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-lg bg-[--color-card] p-5 shadow-sm">
      <p className="text-sm text-[--color-muted]">{label}</p>
      <p className="text-3xl font-bold text-[--color-primary]">{value}</p>
    </div>
  );
}
