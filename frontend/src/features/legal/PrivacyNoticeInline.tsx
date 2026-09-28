import type { PrivacyNoticeResponse } from './schemas';

// Bloque reutilizable para el registro (M2.2): el aviso simplificado se
// muestra COMPLETO en la propia página, no tras un enlace — esconder tras un
// enlace el texto que se está aceptando convertiría la casilla en un trámite
// vacío. El enlace lleva al aviso INTEGRAL, que sí es información adicional.
export function PrivacyNoticeInline({ notice }: { notice: PrivacyNoticeResponse }) {
  return (
    <div className="space-y-2">
      <div className="max-h-48 space-y-3 overflow-y-auto rounded-lg border border-[var(--color-border)] p-3 text-sm">
        {notice.simplified.map((section, i) => (
          <div key={i}>
            <h3 className="font-semibold">{section.heading}</h3>
            {section.paragraphs.map((paragraph, j) => (
              <p key={j} className="mt-1 text-[var(--color-muted)]">{paragraph}</p>
            ))}
          </div>
        ))}
      </div>
      {/* Pestaña nueva a propósito (M2.2): el registro es un flujo de 3 pasos
          con estado en memoria — navegar fuera y volver perdería las
          respuestas del cuestionario ya capturadas. */}
      <a
        href="/privacidad"
        target="_blank"
        rel="noopener noreferrer"
        className="inline-block text-sm underline"
        style={{ color: 'var(--color-primary)' }}
      >
        Ver aviso completo
      </a>
    </div>
  );
}
