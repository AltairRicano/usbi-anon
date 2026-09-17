import { useEffect, useRef, useState, useMemo } from 'react';
import Phaser from 'phaser';
import { CrosswordEngine, type CrosswordState, type PlacedWord, type GameResult } from '@usbi/engine';
import type { CrosswordWord } from '@usbi/schema';
import { CrosswordScene } from './CrosswordScene';
import { Card, CardTitle, CardContent } from '../../shared/components/ui/Card';
import { PhaserGame, type IRefPhaserGame } from '../../shared/PhaserGame';

interface CrosswordGameProps {
  words: CrosswordWord[];
  onFinish?: (result: GameResult) => void;
}

export function CrosswordGame({ words, onFinish }: CrosswordGameProps) {
  const phaserRef = useRef<IRefPhaserGame | null>(null);

  const engine = useMemo(() => new CrosswordEngine(words), [words]);
  const placedWords = useMemo(() => engine.getPlacedWords(), [engine]);
  const unplacedWords = useMemo(() => engine.getUnplacedWords(), [engine]);
  const [state, setState] = useState<CrosswordState>(() => engine.getState());


  useEffect(() => {
    const unsubscribe = engine.subscribe((newState) => {
      setState({ ...newState });
    });

    return () => {
      unsubscribe();
      engine.destroy();
    };
  }, [engine]);

  // We want to assign numbers based on grid position (y, then x)
  const numberedWords = useMemo(() => {
    const sorted = [...placedWords].sort((a, b) => {
      if (a.y === b.y) return a.x - b.x;
      return a.y - b.y;
    });
    
    // We only increment the number if it's a NEW cell. Two words starting at the same cell share the same number.
    const startMap = new Map<string, number>();
    let currentNumber = 1;
    
    return sorted.map(pw => {
      const key = `${pw.x},${pw.y}`;
      let num = startMap.get(key);
      if (num === undefined) {
        num = currentNumber++;
        startMap.set(key, num);
      }
      return { ...pw, num };
    });
  }, [placedWords]);

  const horizontals = useMemo(() => numberedWords.filter(pw => !pw.isVertical), [numberedWords]);
  const verticals = useMemo(() => numberedWords.filter(pw => pw.isVertical), [numberedWords]);

  const isWordSolved = (pw: PlacedWord) => {
    for (let i = 0; i < pw.word.length; i++) {
      const cx = pw.isVertical ? pw.x : pw.x + i;
      const cy = pw.isVertical ? pw.y + i : pw.y;
      if (!state.lockedCells.has(`${cx},${cy}`)) return false;
    }
    return true;
  };

  const gameConfig: Phaser.Types.Core.GameConfig = useMemo(() => ({
    type: Phaser.AUTO,
    width: 720,
    height: 720,
    backgroundColor: '#f8f9fa',
    scene: [CrosswordScene],
    scale: {
      mode: Phaser.Scale.FIT,
      autoCenter: Phaser.Scale.CENTER_BOTH,
    },
    callbacks: {
      preBoot: (game) => {
        game.registry.set('crosswordEngine', engine);
        game.registry.set('crosswordOnFinish', onFinish);
      },
    },
  }), [engine, onFinish]);


  return (
    <Card className="w-full max-w-5xl mx-auto mt-8 flex flex-col md:flex-row gap-4">
      <CardContent className="flex-1 p-0 relative">
        <div
          className="relative mx-auto aspect-square w-full max-w-[720px] overflow-hidden rounded-xl shadow-inner"
          style={{ background: '#f8f9fa' }}
        >
           <PhaserGame
             key={words.map((word) => `${word.word}:${word.clue}`).join('|')}
             ref={phaserRef}
             config={gameConfig}
             ariaLabel="Crucigrama interactivo. Use las flechas para moverse y escriba para completar."
           />
        </div>
      </CardContent>
      <div className="w-full md:w-96 p-6 border-t md:border-t-0 md:border-l border-[--color-border] flex flex-col gap-4 overflow-y-auto md:max-h-[720px]">
        <CardTitle className="text-xl">Crucigrama</CardTitle>
        {/* aria-live mirrors the same score/completion text rendered below (C5):
            a screen reader user gets it announced without navigating to it. */}
        <div aria-live="polite" className="flex flex-col gap-2 bg-white dark:bg-[--color-card] p-4 rounded-xl border border-[--color-border]">
          <p className="text-sm text-[--color-text-muted]">Puntuación</p>
          <p className="font-bold text-2xl text-[--color-primary]">{state.score}</p>
          {state.isFinished && <p className="sr-only">Crucigrama completado.</p>}
        </div>
        {state.isFinished && (
          <div className="mt-2 p-4 bg-green-100 text-green-800 rounded-xl font-bold text-center animate-pulse">
            ¡Completado!
          </div>
        )}
        {unplacedWords.length > 0 && (
          <div className="mt-2 rounded-xl border border-[--color-error] bg-red-50 p-3 text-sm text-[--color-error]">
            Hay palabras que no se pudieron cruzar: {unplacedWords.map((word) => word.word).join(', ')}.
          </div>
        )}
        <div className="mt-4 flex flex-col gap-4">
          <div>
            <h4 className="font-bold text-lg border-b border-[--color-border] pb-2 mb-3">Horizontales</h4>
            <ul className="flex flex-col gap-2">
              {horizontals.map((pw, i) => (
                <li
                  key={`h-${i}`}
                  className={`text-sm flex items-start gap-2 ${isWordSolved(pw) ? 'text-green-600 line-through' : 'text-[--color-text]'}`}
                >
                  <span className="font-bold shrink-0">{pw.num}.</span>
                  <span>{pw.clue}</span>
                  {isWordSolved(pw) && <span aria-hidden="true">✓</span>}
                </li>
              ))}
            </ul>
          </div>
          <div>
            <h4 className="font-bold text-lg border-b border-[--color-border] pb-2 mb-3">Verticales</h4>
            <ul className="flex flex-col gap-2">
              {verticals.map((pw, i) => (
                <li
                  key={`v-${i}`}
                  className={`text-sm flex items-start gap-2 ${isWordSolved(pw) ? 'text-green-600 line-through' : 'text-[--color-text]'}`}
                >
                  <span className="font-bold shrink-0">{pw.num}.</span>
                  <span>{pw.clue}</span>
                  {isWordSolved(pw) && <span aria-hidden="true">✓</span>}
                </li>
              ))}
            </ul>
          </div>
        </div>
      </div>
    </Card>
  );
}
