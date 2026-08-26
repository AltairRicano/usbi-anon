import type { MemoryPairWithColor } from '@usbi/engine';
import { getMemoryBackCardStyle, getMemoryFrontCardStyle, getMemoryReadableTextColor, normalizeMemoryBackColor, normalizeMemoryPairs } from '@usbi/engine';

export function MemoryPreview({ value }: { value: { back_color?: string; pairs?: MemoryPairWithColor[] } }) {
  const pairs = normalizeMemoryPairs(value.pairs ?? []);
  const backColor = normalizeMemoryBackColor(value.back_color);

  return (
    <div className="space-y-4">
      <section className="rounded-lg border border-[--color-border] bg-[--color-card] p-4 shadow-sm">
        <div className="grid gap-4 sm:grid-cols-[160px_1fr]">
          <MemoryBackSample color={backColor} />
          <div className="flex flex-col justify-center">
            <p className="text-sm font-semibold">Dorso común</p>
            <p className="text-xs text-[--color-muted]">{backColor.toUpperCase()}</p>
          </div>
        </div>
      </section>
      <div className="grid gap-4 md:grid-cols-2">
        {pairs.map((pair, index) => (
          <article key={pair.id} className="rounded-lg border border-[--color-border] bg-[--color-card] p-4 shadow-sm">
            <div className="mb-3 flex items-center justify-between gap-3">
              <div>
                <p className="text-sm font-semibold">Par {index + 1}</p>
                <p className="text-xs text-[--color-muted]">{pair.color.toUpperCase()}</p>
              </div>
              <span className="h-4 w-4 rounded-full border border-[--color-border]" style={{ backgroundColor: pair.color }} aria-hidden="true" />
            </div>
            <div className="grid gap-3 sm:grid-cols-2">
              <MemoryFrontSample label="Tarjeta A" content={pair.content1 || 'Concepto'} color={pair.color} />
              <MemoryFrontSample label="Tarjeta B" content={pair.content2 || 'Descripción'} color={pair.color} />
            </div>
          </article>
        ))}
      </div>
      {pairs.length === 0 && <p className="rounded-lg border border-[--color-border] bg-[--color-card] p-4 text-sm text-[--color-muted]">Aún no hay pares configurados.</p>}
    </div>
  );
}

function MemoryBackSample({ color }: { color: string }) {
  return (
    <div
      className="flex h-28 items-center justify-center rounded-xl border border-[--color-border] p-3 text-center shadow-inner"
      style={getMemoryBackCardStyle(color)}
    >
      <div className="space-y-1">
        <p className="text-xs font-semibold uppercase tracking-[0.24em]">Dorso</p>
      </div>
    </div>
  );
}

function MemoryFrontSample({ label, content, color }: { label: string; content: string; color: string }) {
  const textColor = getMemoryReadableTextColor(color);
  return (
    <div
      className="flex min-h-28 flex-col justify-between rounded-xl border border-[--color-border] p-3 shadow-sm"
      style={getMemoryFrontCardStyle(color)}
    >
      <p className="text-[11px] font-semibold uppercase tracking-[0.18em]" style={{ color: textColor }}>{label}</p>
      <p className="text-sm font-semibold leading-tight" style={{ color: textColor }}>
        {content}
      </p>
    </div>
  );
}
