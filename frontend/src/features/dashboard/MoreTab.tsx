import { useEffect, useState, type FormEvent } from 'react';
import { Button } from '../../shared/components/ui/Button';
import { apiClient } from '../../shared/apiClient';
import { errorMessage } from '../../shared/errorMessage';
import { InterestLinksResponseSchema, type CategoryWithLinks } from './interestLinksSchemas';
import { InterestLinkCarousel } from './InterestLinkCarousel';

const SUGGESTION_MAX_LEN = 1000;

export function MoreTab() {
  const [groups, setGroups] = useState<CategoryWithLinks[]>([]);
  const [loadError, setLoadError] = useState<string | null>(null);

  const [description, setDescription] = useState('');
  const [sending, setSending] = useState(false);
  const [sendError, setSendError] = useState<string | null>(null);
  const [sent, setSent] = useState(false);

  useEffect(() => {
    apiClient
      .get('/interest-links')
      .then((resp) => setGroups(InterestLinksResponseSchema.parse(resp.data).items))
      .catch(() => setLoadError('No se pudieron cargar los enlaces de interés.'));
  }, []);

  async function submitSuggestion(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setSendError(null);
    setSent(false);
    if (!description.trim()) return;
    setSending(true);
    try {
      // Anónima por diseño (backend/internal/suggestions): no se manda ni se
      // recibe de vuelta ningún dato de identidad, solo la confirmación.
      await apiClient.post('/suggestions', { description: description.trim() });
      setDescription('');
      setSent(true);
    } catch (err) {
      setSendError(errorMessage(err, 'No se pudo enviar la sugerencia.'));
    } finally {
      setSending(false);
    }
  }

  return (
    <div className="space-y-6">
      {loadError && <p className="rounded border border-[--color-error] bg-[--color-card] p-3 text-[--color-error]">{loadError}</p>}

      {groups.map((group) =>
        group.links.length === 0 ? (
          <section key={group.category.id} className="rounded-lg bg-[--color-card] text-[--color-text-card] p-6 shadow-sm">
            <h2 className="mb-4 text-center text-2xl font-semibold">{group.category.name}</h2>
            <p className="text-center text-sm text-[--color-muted]">Sin enlaces por ahora.</p>
          </section>
        ) : (
          <InterestLinkCarousel key={group.category.id} group={group} />
        )
      )}
      {groups.length === 0 && !loadError && (
        <p className="rounded-lg border border-[--color-border] bg-[--color-card] p-5 text-[--color-muted]">
          No hay enlaces de interés publicados.
        </p>
      )}

      <section className="rounded-lg bg-[--color-card] text-[--color-text-card] p-6 shadow-sm">
        <h2 className="mb-2 text-xl font-semibold">Buzón de sugerencias</h2>
        <p className="mb-4 text-sm text-[--color-muted]">
          Tu sugerencia se envía de forma anónima: no queda ligada a tu cuenta.
        </p>

        {sendError && <p className="mb-3 rounded border border-[--color-error] bg-[--color-card] p-3 text-[--color-error]">{sendError}</p>}
        {sent && <p className="mb-3 rounded border border-[--color-border] bg-[--color-card] p-3 text-[--color-primary]">¡Gracias! Tu sugerencia se envió.</p>}

        <form onSubmit={submitSuggestion} className="space-y-3">
          <textarea
            value={description}
            onChange={(e) => setDescription(e.currentTarget.value)}
            maxLength={SUGGESTION_MAX_LEN}
            rows={4}
            required
            placeholder="¿Qué te gustaría que agregáramos o cambiáramos?"
            className="w-full rounded-lg border px-4 py-2 text-base border-[--color-border] bg-[--color-background] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[--color-primary]"
          />
          <Button type="submit" disabled={sending}>
            {sending ? 'Enviando…' : 'Enviar sugerencia'}
          </Button>
        </form>
      </section>
    </div>
  );
}
