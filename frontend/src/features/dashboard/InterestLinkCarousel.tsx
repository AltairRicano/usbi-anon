import { useState } from 'react';
import type { CategoryWithLinks } from './interestLinksSchemas';

// Elige texto negro o blanco según la luminancia relativa del color de fondo
// (WCAG 2.2), porque el color de cada tarjeta lo elige libremente la persona
// administradora (AdminCommunityPage) y no se puede asumir que sea oscuro.
function readableTextColor(hex: string): string {
  const m = /^#?([0-9a-f]{6})$/i.exec(hex.trim());
  if (!m) return '#000000';
  const n = parseInt(m[1], 16);
  const r = (n >> 16) & 255;
  const g = (n >> 8) & 255;
  const b = n & 255;
  const toLinear = (c: number) => {
    const s = c / 255;
    return s <= 0.03928 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
  };
  const luminance = 0.2126 * toLinear(r) + 0.7152 * toLinear(g) + 0.0722 * toLinear(b);
  return luminance > 0.5 ? '#000000' : '#FFFFFF';
}

// Carrusel "infinito" de tarjetas por categoría: una tarjeta central grande
// (fondo = color elegido por el admin, título+descripción centrados, tarjeta
// completa como hipervínculo sin subrayar) con flechas a los lados y un
// borde asomando de las tarjetas vecinas, como fichas sostenidas en abanico.
// El índice avanza/retrocede con módulo, así que derecha→izquierda siempre
// regresa a la tarjeta anterior sin necesidad de guardar historial aparte.
export function InterestLinkCarousel({ group }: { group: CategoryWithLinks }) {
  const { links } = group;
  const [index, setIndex] = useState(0);
  const count = links.length;

  const prevLink = links[(index - 1 + count) % count];
  const current = links[index];
  const nextLink = links[(index + 1) % count];

  const goPrev = () => setIndex((i) => (i - 1 + count) % count);
  const goNext = () => setIndex((i) => (i + 1) % count);

  return (
    <section className="rounded-lg bg-[--color-card] text-[--color-text-card] p-6 shadow-sm">
      <h2 className="mb-6 text-center text-2xl font-semibold">{group.category.name}</h2>

      <div className="flex items-center justify-center gap-1 sm:gap-3">
        <button
          type="button"
          onClick={goPrev}
          disabled={count < 2}
          aria-label="Tarjeta anterior"
          className="flex shrink-0 items-center justify-center rounded-full text-2xl font-bold text-[--color-primary] hover:text-[--color-primary-hover] disabled:opacity-30"
          style={{ minWidth: 'var(--spacing-touch)', minHeight: 'var(--spacing-touch)' }}
        >
          ‹
        </button>

        <div className="flex items-center">
          {count > 1 && (
            <div
              aria-hidden="true"
              className="hidden h-32 w-4 shrink-0 rounded-l-xl sm:block"
              style={{ backgroundColor: prevLink.color }}
            />
          )}

          <a
            key={current.id}
            href={current.url}
            target="_blank"
            rel="noopener noreferrer"
            className="relative z-10 flex min-h-[10rem] w-64 sm:w-72 flex-col items-center justify-center gap-2 rounded-2xl p-6 text-center no-underline shadow-md transition-transform hover:scale-[1.03]"
            style={{ backgroundColor: current.color }}
          >
            {/* Los <h1>-<h6> tienen color: var(--color-primary) en el
                layer base global, que gana sobre el color heredado del
                <a> padre — hay que fijar el color aquí explícitamente. */}
            <h3 className="text-lg font-bold" style={{ color: readableTextColor(current.color) }}>
              {current.title}
            </h3>
            <p className="text-sm" style={{ color: readableTextColor(current.color) }}>
              {current.description}
            </p>
          </a>

          {count > 1 && (
            <div
              aria-hidden="true"
              className="hidden h-32 w-4 shrink-0 rounded-r-xl sm:block"
              style={{ backgroundColor: nextLink.color }}
            />
          )}
        </div>

        <button
          type="button"
          onClick={goNext}
          disabled={count < 2}
          aria-label="Tarjeta siguiente"
          className="flex shrink-0 items-center justify-center rounded-full text-2xl font-bold text-[--color-primary] hover:text-[--color-primary-hover] disabled:opacity-30"
          style={{ minWidth: 'var(--spacing-touch)', minHeight: 'var(--spacing-touch)' }}
        >
          ›
        </button>
      </div>

      {count > 1 && (
        <p className="mt-3 text-center text-xs text-[--color-muted]">
          {index + 1} / {count}
        </p>
      )}
    </section>
  );
}
