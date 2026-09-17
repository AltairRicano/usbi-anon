import { useEffect, useState, useMemo } from 'react';
import { PuzzleEngine, PuzzleState, type PuzzlePiece, type GameResult } from '@usbi/engine';
import { Card, CardTitle } from '../../shared/components/ui/Card';
import { Reorder } from 'framer-motion';

// -700 shades (not the original -500s): verified white-text contrast >= 4.5:1
// for every color in the palette (audit finding C8 — bg-yellow-500 was 1.92:1).
const COLORS = [
  'bg-red-700', 'bg-blue-700', 'bg-green-700', 'bg-yellow-700',
  'bg-purple-700', 'bg-pink-700', 'bg-indigo-700', 'bg-teal-700'
];

interface PuzzleGameProps {
  phrase: string;
  pieces?: number;
  seed?: number;
  onFinish?: (result: GameResult) => void;
}

export function PuzzleGame({ phrase, pieces = 3, seed = 1234, onFinish }: PuzzleGameProps) {
  const engine = useMemo(() => new PuzzleEngine(phrase, pieces, seed), [phrase, pieces, seed]);
  const [state, setState] = useState<PuzzleState>(() => engine.getState());
  const [items, setItems] = useState(state.pieces);

  useEffect(() => {
    setItems(state.pieces);
  }, [state.pieces]);

  useEffect(() => {
    setState({ ...engine.getState() });

    const unsubscribe = engine.subscribe((newState) => {
      setState({ ...newState });
      if (newState.isFinished && onFinish) {
        onFinish(engine.getResult());
      }
    });

    return () => {
      unsubscribe();
      engine.destroy();
    };
  }, [engine, onFinish]);

  const handleReorder = (newItems: PuzzlePiece[]) => {
    // Only dispatch reorder to engine when drag ends or logic triggers update
    // For simplicity, we just diff the first difference
    const diffIndex = items.findIndex((p, i) => p.id !== newItems[i]?.id);
    if (diffIndex !== -1) {
      const movedPiece = items[diffIndex];
      const newIndex = newItems.findIndex(p => p.id === movedPiece.id);
      if (newIndex !== -1 && !state.isFinished) {
         engine.reorderPieces(diffIndex, newIndex);
      }
    }
  };

  return (
    <Card className="w-full max-w-4xl mx-auto mt-8 flex flex-col items-center p-6 gap-6">
      <CardTitle className="text-2xl">Mensaje Secreto</CardTitle>

      <div className="flex gap-4 items-center">
        <p className="font-bold text-lg text-gray-700 dark:text-gray-300">Movimientos: {state.moves}</p>
        <p className="font-bold text-lg text-[--color-primary]">Puntuación: {state.score}</p>
      </div>

      <div className="w-full min-h-[200px] border-4 border-[--color-border] rounded-xl p-4 bg-gray-50 shadow-inner overflow-x-auto">
        <Reorder.Group 
          axis="x" 
          values={items} 
          onReorder={handleReorder} 
          className="flex flex-wrap gap-2 justify-center"
        >
          {items.map((piece) => {
            const colorClass = COLORS[piece.originalIndex % COLORS.length];
            return (
              <Reorder.Item 
                key={piece.id} 
                value={piece}
                className={`px-4 py-3 rounded-lg shadow-md cursor-grab active:cursor-grabbing text-white font-bold text-xl select-none flex-shrink-0 ${colorClass}`}
              >
                {piece.text}
              </Reorder.Item>
            );
          })}
        </Reorder.Group>
      </div>

      {state.isFinished && (
        <div className="p-4 bg-green-100 text-green-800 rounded-xl font-bold text-center animate-bounce text-xl mt-4">
          ¡Felicidades, descubriste el mensaje!
        </div>
      )}
    </Card>
  );
}
