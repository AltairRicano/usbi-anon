import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { HomeButton } from '../../shared/components/ui/HomeButton';
import { LinkButton } from '../../shared/components/ui/LinkButton';
import { apiClient } from '../../shared/apiClient';
import { errorMessage } from '../../shared/errorMessage';
import type { LevelSummaryDTO, SectionDTO } from './types';
import { templateTypeLabel } from './types';
import { LevelsPageDTOSchema, SectionsResponseSchema } from './schemas';

export function SectionLevelsPage() {
  const { sectionId } = useParams();
  const [section, setSection] = useState<SectionDTO | null>(null);
  const [levels, setLevels] = useState<LevelSummaryDTO[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!sectionId) return;
    let cancelled = false;
    (async () => {
      try {
        const [sectionResp, levelResp] = await Promise.all([
          apiClient.get('/sections'),
          apiClient.get(`/levels?section_id=${sectionId}&page_size=50`),
        ]);
        if (cancelled) return;
        const sections = SectionsResponseSchema.parse(sectionResp.data).items;
        setSection(sections.find((item) => item.id === sectionId) ?? null);
        setLevels(LevelsPageDTOSchema.parse(levelResp.data).items);
      } catch (err) {
        if (!cancelled) setError(errorMessage(err, 'No se pudieron cargar los niveles.'));
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [sectionId]);

  return (
    <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
      <div className="mx-auto max-w-5xl space-y-6">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">{section?.title ?? 'Sección'}</h1>
            <p className="text-sm text-[var(--color-muted)]">Niveles oficiales publicados.</p>
          </div>
          <HomeButton />
        </header>

        {error && <p className="rounded border border-[var(--color-error)] bg-[var(--color-card)] p-3 text-[var(--color-error)]">{error}</p>}

        <section className="grid gap-4 md:grid-cols-2">
          {levels.map((level) => (
            <article key={level.id} className="rounded-lg bg-[var(--color-card)] p-5 shadow-sm">
              <div className="mb-4 h-2 rounded" style={{ backgroundColor: level.color }} />
              <h2 className="text-xl font-semibold">{level.title}</h2>
              <p className="mb-4 text-sm text-[var(--color-muted)]">Dificultad {level.difficulty} · {templateTypeLabel(level.template_type)}</p>
              <LinkButton to={`/levels/${level.id}/play`} variant="primary" size="sm">Jugar</LinkButton>
            </article>
          ))}
          {levels.length === 0 && <p className="rounded-lg bg-[var(--color-card)] p-5 text-[var(--color-muted)]">No hay niveles publicados en esta sección.</p>}
        </section>
      </div>
    </main>
  );
}
