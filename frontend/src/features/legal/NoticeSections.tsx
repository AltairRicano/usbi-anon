import type { PrivacyNoticeSection } from './schemas';

// Acordeón accesible con <details>/<summary> nativos — no modal (M2.3): este
// proyecto ya arrastró un fallo real donde el filtro de daltonismo rompía
// todos los overlays position:fixed, los modales se comportan mal con zoom
// al 200%, y <details> es navegable con teclado sin trampa de foco.
export function NoticeSections({ sections, defaultOpenFirst = false }: { sections: PrivacyNoticeSection[]; defaultOpenFirst?: boolean }) {
  return (
    <div className="space-y-2">
      {sections.map((section, i) => (
        <details
          key={i}
          open={defaultOpenFirst && i === 0}
          className="rounded-lg border border-[--color-border] bg-[--color-card]"
        >
          <summary
            className="flex min-h-[44px] cursor-pointer select-none items-center rounded-lg px-4 py-3 font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[--color-primary]"
            style={{ color: 'var(--color-primary)' }}
          >
            {section.heading}
          </summary>
          <div className="space-y-2 px-4 pb-4 text-sm text-[--color-text]">
            {section.paragraphs.map((paragraph, j) => (
              <p key={j}>{paragraph}</p>
            ))}
          </div>
        </details>
      ))}
    </div>
  );
}
