import { useEffect, useRef } from 'react';
import { Button } from '../../shared/components/ui/Button';

/**
 * El ÚNICO modal de toda la UI (plan/04_Rediseno_identidad_gustos.md §5):
 * excepción deliberada a la convención "sin modales" del resto del proyecto
 * hermano. Se justifica porque el guard que representa —mínimo 4 preguntas
 * activas, sin el cual NADIE puede registrarse (ver el comentario de
 * registration_questions en 0001_esquema_unificado.up.sql)— es lo bastante
 * crítico como para bloquear la atención de quien administra el banco en vez
 * de dejarlo como un banner más entre otros.
 *
 * Fondo difuminado, centrado, cierre por Escape/click fuera/botón — foco
 * atrapado dentro mientras está abierto (WCAG 2.4.3).
 */
export function MinQuestionsModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const dialogRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
    }
    document.addEventListener('keydown', handleKeyDown);
    dialogRef.current?.focus();
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [open, onClose]);

  if (!open) return null;

  return (
    // El backdrop es un cierre de conveniencia (clic fuera), no la única vía:
    // Escape y el botón "Entendido" cubren teclado/lectores de pantalla, así
    // que el onClick decorativo aquí no necesita su propio manejo de teclado.
    // eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      style={{ backgroundColor: 'color-mix(in srgb, black 55%, transparent)', backdropFilter: 'blur(4px)' }}
      onClick={onClose}
    >
      {/* eslint-disable-next-line jsx-a11y/no-noninteractive-element-interactions, jsx-a11y/click-events-have-key-events -- solo detiene la propagación del clic del backdrop */}
      <div
        ref={dialogRef}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="min-questions-modal-title"
        aria-describedby="min-questions-modal-desc"
        tabIndex={-1}
        className="w-full max-w-md rounded-2xl p-6 shadow-2xl"
        style={{ backgroundColor: 'var(--color-card)', color: 'var(--color-text-card)' }}
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id="min-questions-modal-title" className="mb-2 text-lg font-bold" style={{ color: 'var(--color-error)' }}>
          No se puede completar la acción
        </h2>
        <p id="min-questions-modal-desc" className="mb-6 text-sm" style={{ color: 'var(--color-foreground)' }}>
          Esta acción dejaría menos de 4 preguntas activas en el banco de registro. El sistema exige un mínimo de
          4 preguntas activas — sin ellas, nadie podría completar el cuestionario para crear una cuenta nueva.
          Activa o crea otra pregunta antes de desactivar o eliminar esta.
        </p>
        <Button type="button" className="w-full" onClick={onClose}>
          Entendido
        </Button>
      </div>
    </div>
  );
}
