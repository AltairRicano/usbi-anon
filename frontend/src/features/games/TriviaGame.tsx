import { useEffect, useState, useRef } from 'react';
import { TriviaEngine, TriviaState, GameResult } from '@usbi/engine';
import { MultipleChoice } from '@usbi/schema';
import { Card, CardHeader, CardTitle, CardContent } from '../../shared/components/ui/Card';
import { Button } from '../../shared/components/ui/Button';

interface TriviaGameProps {
  questions: MultipleChoice[];
  onFinish?: (result: GameResult) => void;
}

export function TriviaGame({ questions, onFinish }: TriviaGameProps) {
  const engineRef = useRef<TriviaEngine | null>(null);
  const [state, setState] = useState<TriviaState | null>(null);

  useEffect(() => {
    const engine = new TriviaEngine(questions);
    engineRef.current = engine;

    const unsubscribe = engine.subscribe((newState) => {
      setState({ ...newState });
      if (newState.isFinished && onFinish) {
        onFinish(engine.getResult());
      }
    });

    engine.startTimer();

    return () => {
      unsubscribe();
      engine.destroy();
    };
  }, [questions, onFinish]);

  if (!state) return <div>Loading...</div>;

  // La pantalla de cierre (stats verificados por servidor, XP, reintentar)
  // la muestra la página contenedora (OfficialLevelPage/LocalLevelPage) una
  // vez que onFinish resuelve; renderizar aquí una tarjeta propia duplicaba
  // esa pantalla y aparecía un instante de por medio mientras se esperaba
  // la respuesta del servidor.
  if (state.isFinished) {
    return null;
  }

  const currentQ = state.questions[state.currentQuestionIndex];

  return (
    <Card className="w-full max-w-2xl mx-auto mt-8 relative overflow-hidden">
      <div 
        className="absolute top-0 left-0 h-2 bg-[var(--color-primary)] transition-all duration-1000"
        style={{ width: `${(state.timeLeft / 30) * 100}%` }}
      />
      <CardHeader>
        <div className="flex justify-between items-center w-full">
          <CardTitle>Pregunta {state.currentQuestionIndex + 1} de {state.questions.length}</CardTitle>
          <span className="font-bold text-lg">{state.score} pts</span>
        </div>
        <p className="text-[var(--color-muted)] text-sm">Tiempo restante: {state.timeLeft}s</p>
      </CardHeader>
      <CardContent className="flex flex-col gap-6">
        <div className="text-xl font-medium text-center py-4">
          {currentQ.question}
        </div>
        {currentQ.media_url && (
          <img src={currentQ.media_url} alt="Pregunta" className="w-full max-h-64 object-contain rounded-xl" />
        )}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {currentQ.options.map((option, idx) => {
            let variant: 'primary' | 'outline' | 'danger' | 'secondary' = 'outline';
            if (state.selectedAnswer !== null) {
              if (idx === currentQ.correct_index) {
                variant = 'primary'; // Correct answer highlights in primary
              } else if (idx === state.selectedAnswer) {
                variant = 'danger'; // Wrong answer chosen
              }
            }
            return (
              <Button
                key={idx}
                variant={variant}
                size="lg"
                disabled={state.selectedAnswer !== null}
                onClick={() => engineRef.current?.submitAnswer(idx)}
                className="w-full justify-start text-left h-auto py-4 whitespace-normal"
              >
                {option}
              </Button>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
