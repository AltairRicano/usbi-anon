import React, { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { Snakes } from '@usbi/schema';
import { SnakeLadderEngine, SnakeLadderConfig, GameResult } from '@usbi/engine';
import { PhaserGame, IRefPhaserGame } from '../../../shared/PhaserGame';
import { Button } from '../../../shared/components/ui/Button';
import { SnakeLadderScene } from '../phaser/SnakeLadderScene';
import Phaser from 'phaser';

interface SnakeLadderGameProps {
  level: Snakes;
  onFinish: (result: GameResult) => void;
}

export const SnakeLadderGame: React.FC<SnakeLadderGameProps> = ({ level, onFinish }) => {
  const phaserRef = useRef<IRefPhaserGame | null>(null);
  
  const engineConfig = useMemo<SnakeLadderConfig>(() => ({
      boardSize: level.board_width * level.board_height,
      boardWidth: level.board_width,
      boardHeight: level.board_height,
      startPosition: level.start_position,
      endPosition: level.end_position,
      snakes: level.snakes || [],
      ladders: level.ladders || [],
      aiDifficulty: level.ai_config?.difficulty || 'MEDIUM'
  }), [level]);

  const engine = useMemo(() => new SnakeLadderEngine(engineConfig), [engineConfig]);
  const [isGameOver, setIsGameOver] = useState(false);
  const [canRoll, setCanRoll] = useState(false);
  const [winner, setWinner] = useState<string | null>(null);
  // Mirrors engine.state.message/lastRoll — today drawn ONLY on the Phaser
  // canvas (messageText/diceText) — into the DOM so a screen reader can
  // announce turn/dice-roll/correct-incorrect changes (audit finding C5).
  // engine.state is a plain mutable object (no subscribe API), so it's
  // sampled on the same interval already used below for canRoll.
  const [liveMessage, setLiveMessage] = useState(engine.state.message);

  useEffect(() => {
    engine.start();
  }, [engine]);

  const gameConfig: Phaser.Types.Core.GameConfig = useMemo(() => ({
      type: Phaser.AUTO,
      width: 720,
      height: 780,
      backgroundColor: '#f8fafc',
      scale: {
          mode: Phaser.Scale.FIT,
          autoCenter: Phaser.Scale.CENTER_BOTH,
      },
      scene: [SnakeLadderScene],
      callbacks: {
          preBoot: (game) => {
              game.registry.set('config', engineConfig);
              game.registry.set('engine', engine);
          }
      }
  }), [engineConfig, engine]);

  // phaserRef never changes identity (it's a ref), and Phaser assigns
  // phaserRef.current.scene asynchronously after boot — an effect keyed on
  // phaserRef alone runs once, immediately, before that assignment happens,
  // and never re-runs to notice it later. onGameReady instead fires exactly
  // when the Phaser game (and its registered scene) actually exist, so we
  // stash that scene in state and let a normal effect react to it.
  const [readyScene, setReadyScene] = useState<SnakeLadderScene | null>(null);
  const handleGameReady = useCallback((game: Phaser.Game) => {
      setReadyScene((game.scene.getScene('SnakeLadderScene') as SnakeLadderScene) ?? null);
  }, []);

  useEffect(() => {
     if (!readyScene) return;

     const handleGameOver = (data: { winner: string }) => {
         setIsGameOver(true);
         setWinner(data.winner);
         onFinish(engine.getResult());
     };

     readyScene.events.on('GAME_OVER', handleGameOver);

     // Re-check roll button state periodically since scene drives it
     const interval = setInterval(() => {
         setCanRoll(engine.state.state === 'player_turn' && !readyScene.isAnimating);
         setLiveMessage(engine.state.message);
     }, 100);

     return () => {
         readyScene.events.off('GAME_OVER', handleGameOver);
         clearInterval(interval);
     };
  }, [readyScene, engine, onFinish]);

  // Cola de preguntas pendientes por responder: cada tiro de dado consume la
  // de al frente. Una respuesta correcta la retira definitivamente de esta
  // vuelta; una incorrecta la manda al final de la cola (se repite más
  // adelante) y cede el turno a la IA. Cuando se vacía (todas respondidas
  // bien al menos una vez) se rebaraja para que siempre haya pregunta.
  const questions = useMemo(() => level.questions ?? [], [level]);
  const [queue, setQueue] = useState<number[]>(() => shuffledIndices(questions.length));
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState<number | null>(null);
  const showQuestionModal = currentQuestionIndex !== null;
  const currentQuestion = currentQuestionIndex !== null ? questions[currentQuestionIndex] : null;

  const handleRollClick = () => {
      if (questions.length === 0) {
          executeRoll();
          return;
      }
      const pending = queue.length > 0 ? queue : shuffledIndices(questions.length);
      const [nextIndex, ...rest] = pending;
      setQueue(rest);
      setCurrentQuestionIndex(nextIndex);
  };

  const executeRoll = () => {
      if (phaserRef.current?.scene && canRoll) {
          setCanRoll(false);
          phaserRef.current.scene.events.emit('ROLL_DICE');
      }
  };

  const handleAnswer = (index: number) => {
      if (currentQuestionIndex === null || !currentQuestion) return;
      const answeredIndex = currentQuestionIndex;
      setCurrentQuestionIndex(null);
      if (index === currentQuestion.correct_index) {
          executeRoll();
      } else {
          setQueue((prev) => [...prev, answeredIndex]);
          setCanRoll(false);
          engine.state.message = "¡Respuesta incorrecta! La pregunta vuelve a la cola. Pierdes el turno.";
          engine.state.state = 'ai_turn';
          if (phaserRef.current?.scene) {
              const scene = phaserRef.current.scene;
              const snakeScene = scene as unknown as { messageText?: Phaser.GameObjects.Text };
              if (snakeScene.messageText) {
                  snakeScene.messageText.setText(engine.state.message);
              }
              scene.events.emit('AI_PLAY');
          }
      }
  };

  return (
      <div className="flex w-full flex-col items-center justify-center gap-4 px-2 py-4 sm:px-4">
         <p aria-live="polite" className="sr-only">{liveMessage}</p>
         {!isGameOver ? (
             <>
                 <div
                     className="relative w-full overflow-hidden rounded-lg border border-[--color-border] bg-[--color-card] shadow-sm"
                     style={{ maxWidth: 'min(92vw, 720px)', aspectRatio: '720 / 780' }}
                 >
                     <PhaserGame
                         key={`${level.board_width}x${level.board_height}-${level.seed ?? 'random'}-${level.snakes?.length ?? 0}-${level.ladders?.length ?? 0}`}
                         ref={phaserRef}
                         config={gameConfig}
                         ariaLabel="Tablero de serpientes y escaleras. Use el botón Tirar Dado."
                         onGameReady={handleGameReady}
                     />
                 </div>
                 <Button
                     onClick={handleRollClick}
                     disabled={!canRoll}
                     size="lg"
                     className="rounded-full shadow-md"
                 >
                     Tirar Dado
                 </Button>
             </>
         ) : (
             <div aria-live="polite" className="flex flex-col items-center justify-center p-8 text-center text-white bg-slate-800 rounded-xl">
                 <h2 className="text-3xl font-bold mb-4">Juego Terminado</h2>
                 <p className="text-xl">Ganador: {winner === 'player' ? '¡Tú!' : 'La Inteligencia Artificial'}</p>
             </div>
         )}
         
         {showQuestionModal && currentQuestion && (
             // Tarjeta de pregunta con colores fijos (blanco/negro), no con las
             // variables --color-* institucionales: los minijuegos usan
             // deliberadamente una paleta propia separada de esa (ver nota C9
             // en index.css). Ademas, --color-card/--color-primary aqui se
             // usaban sin var(...), CSS invalido que dejaba la tarjeta
             // transparente y el texto en el negro por defecto del navegador
             // sobre el overlay oscuro -- ilegible en movil.
             <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
                 <div className="bg-white text-black p-6 rounded-xl shadow-xl max-w-md w-full border border-slate-200">
                     <h3 className="text-xl font-bold mb-4 text-[#18529D]">Para tirar el dado, responde:</h3>
                     <p className="text-lg mb-6">{currentQuestion.question}</p>
                     <div className="flex flex-col gap-3">
                         {currentQuestion.options.map((opt: string, idx: number) => (
                             <button
                                 key={idx}
                                 onClick={() => handleAnswer(idx)}
                                 className="px-4 py-3 text-left border border-slate-300 rounded-lg hover:bg-[#18529D]/10 hover:border-[#18529D] transition-colors"
                             >
                                 {opt}
                             </button>
                         ))}
                     </div>
                 </div>
             </div>
         )}
      </div>
  );
};

function shuffledIndices(length: number): number[] {
  const indices = Array.from({ length }, (_, i) => i);
  for (let i = indices.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [indices[i], indices[j]] = [indices[j], indices[i]];
  }
  return indices;
}
