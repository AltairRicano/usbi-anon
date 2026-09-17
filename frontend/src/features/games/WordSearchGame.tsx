import { useEffect, useRef, useState, useMemo, useCallback } from 'react';
import Phaser from 'phaser';
import { WordSearchEngine, WordSearchState, GameResult } from '@usbi/engine';
import { WordSearchScene } from './WordSearchScene';
import { Card, CardTitle, CardContent } from '../../shared/components/ui/Card';
import { PhaserGame, IRefPhaserGame } from '../../shared/PhaserGame';

interface WordSearchGameProps {
  words: string[];
  width?: number;
  height?: number;
  seed?: number;
  onFinish?: (result: GameResult) => void;
}

export function WordSearchGame({ words, width = 10, height = 10, seed = 1234, onFinish }: WordSearchGameProps) {
  const phaserRef = useRef<IRefPhaserGame | null>(null);

  const engine = useMemo(() => new WordSearchEngine(words, width, height, seed), [words, width, height, seed]);
  const [state, setState] = useState<WordSearchState>(() => engine.getState());

  useEffect(() => {
    // Set initial state from engine
    setState({ ...engine.getState() });

    const unsubscribe = engine.subscribe((newState) => {
      setState({ ...newState });
    });

    return () => {
      unsubscribe();
      engine.destroy();
    };
  }, [engine]);

  const gameConfig: Phaser.Types.Core.GameConfig = useMemo(() => ({
    type: Phaser.AUTO,
    width: 500,
    height: 500,
    backgroundColor: '#ffffff',
    scene: [WordSearchScene],
    scale: {
      mode: Phaser.Scale.FIT,
      autoCenter: Phaser.Scale.CENTER_BOTH,
    },
  }), []);

  // This callback is guaranteed to fire after Phaser is ready
  const handleGameReady = useCallback((game: Phaser.Game) => {
    game.scene.start('WordSearchScene', { engine, onFinish });
  }, [engine, onFinish]);


  return (
    <Card className="w-full max-w-4xl mx-auto mt-8 flex flex-col md:flex-row gap-4">
      <CardContent className="flex-1 p-0">
        <div className="w-full rounded-xl overflow-hidden relative" style={{ minHeight: '500px' }}>
           <PhaserGame
             ref={phaserRef}
             config={gameConfig}
             onGameReady={handleGameReady}
             ariaLabel="Sopa de letras interactiva. Use las flechas para moverse y Enter para marcar el inicio y fin de una palabra."
           />
        </div>
      </CardContent>
      <div className="w-full md:w-64 p-6 border-l border-[--color-border] flex flex-col gap-4">
        <CardTitle className="text-xl">Palabras a buscar</CardTitle>
        {/* aria-live mirrors the score already shown here (C5), so a screen
            reader announces it changing without the user navigating to it. */}
        <p aria-live="polite" className="font-bold text-[--color-primary]">Score: {state.score}</p>
        <ul className="flex flex-col gap-2">
          {state.words.map((word, i) => (
            <li 
              key={i}
              className={`text-lg font-semibold transition-colors ${state.foundWords.includes(word) ? 'text-[--color-muted] line-through' : 'text-[--color-text]'}`}
            >
              {word}
            </li>
          ))}
        </ul>
      </div>
    </Card>
  );
}
