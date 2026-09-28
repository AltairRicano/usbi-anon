// F10.10 cerrada: features/games/PuzzleGame ya existe — se conecta aquí el
// componente real en lugar del placeholder de texto que F10.8 dejó pendiente.
import { lazy, Suspense } from 'react';
import type { Puzzle } from '@usbi/schema';

const PuzzleGame = lazy(() => import('../../../games/PuzzleGame').then((mod) => ({ default: mod.PuzzleGame })));

export function PuzzlePreview({ value }: { value: Partial<Puzzle> }) {
  if (!value.phrase) {
    return (
      <div className="p-4 bg-gray-50 border rounded-md text-[var(--color-muted)] text-center">
        Ingresa una frase para ver la vista previa.
      </div>
    );
  }

  return (
    <Suspense
      fallback={
        <div className="p-4 bg-gray-50 border rounded-md text-center text-[var(--color-muted)]">
          Cargando previsualización...
        </div>
      }
    >
      <PuzzleGame
        phrase={value.phrase}
        pieces={value.pieces ?? 3}
        seed={value.seed ?? 1}
        onFinish={() => undefined}
      />
    </Suspense>
  );
}
