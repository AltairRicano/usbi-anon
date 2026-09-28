import type { ChangeEvent } from 'react';
import { Button } from '../../../../shared/components/ui/Button';
import { Input } from '../../../../shared/components/ui/Input';
import { SnakeLadderBoardPreview } from '../previews/SnakeLadderPreview';
import {
  clampBoardSize,
  generateSnakeLadderLinks,
  maxFeatureCount,
  type SnakeLadderContent,
} from '../snakesLayout';

type SnakeQuestion = NonNullable<SnakeLadderContent['questions']>[number];

// Cada pregunta gatea un turno de dado con exactamente dos opciones, y el
// backend exige un mínimo de 8 para que la cola no se repita muy rápido.
const MIN_QUESTIONS = 8;

export function SnakeLadderForm({
  value,
  onChange,
}: {
  value: SnakeLadderContent;
  onChange: (val: SnakeLadderContent) => void;
}) {
  const boardWidth = clampBoardSize(value.board_width ?? 6);
  const boardHeight = clampBoardSize(value.board_height ?? 6);
  const seed = Number.isFinite(value.seed) ? Number(value.seed) : 2026;
  const maxCount = maxFeatureCount(boardWidth, boardHeight);
  const snakeCount = Math.min(value.snakes?.length ?? 3, maxCount);
  const ladderCount = Math.min(value.ladders?.length ?? 3, maxCount);
  const questions: SnakeQuestion[] = value.questions || [];
  const previewValue = normalizeContent(value, boardWidth, boardHeight, snakeCount, ladderCount, seed);

  const addQuestion = () => onChange({ ...value, questions: [...questions, createEmptyQuestion()] });
  const removeQuestion = (idx: number) => {
    if (questions.length <= MIN_QUESTIONS) return;
    onChange({ ...value, questions: questions.filter((_, i) => i !== idx) });
  };
  const updateQuestion = (idx: number, updated: SnakeQuestion) => {
    const next = [...questions];
    next[idx] = updated;
    onChange({ ...value, questions: next });
  };

  function regenerate(updates: Partial<{ boardWidth: number; boardHeight: number; seed: number; snakeCount: number; ladderCount: number }>) {
    const nextWidth = clampBoardSize(updates.boardWidth ?? boardWidth);
    const nextHeight = clampBoardSize(updates.boardHeight ?? boardHeight);
    const nextSeed = Number.isFinite(updates.seed) ? Number(updates.seed) : seed;
    const nextMax = maxFeatureCount(nextWidth, nextHeight);
    const nextSnakeCount = clampCount(updates.snakeCount ?? snakeCount, nextMax);
    const nextLadderCount = clampCount(updates.ladderCount ?? ladderCount, nextMax);
    onChange(normalizeContent(value, nextWidth, nextHeight, nextSnakeCount, nextLadderCount, nextSeed));
  }

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Input
          label="Columnas"
          type="number"
          min={2}
          max={10}
          value={boardWidth}
          onChange={(event: ChangeEvent<HTMLInputElement>) => regenerate({ boardWidth: Number(event.target.value) })}
          required
        />
        <Input
          label="Filas"
          type="number"
          min={2}
          max={10}
          value={boardHeight}
          onChange={(event: ChangeEvent<HTMLInputElement>) => regenerate({ boardHeight: Number(event.target.value) })}
          required
        />
        <Input
          label="Serpientes"
          type="number"
          min={0}
          max={maxCount}
          value={snakeCount}
          onChange={(event: ChangeEvent<HTMLInputElement>) => regenerate({ snakeCount: Number(event.target.value) })}
          required
        />
        <Input
          label="Escaleras"
          type="number"
          min={0}
          max={maxCount}
          value={ladderCount}
          onChange={(event: ChangeEvent<HTMLInputElement>) => regenerate({ ladderCount: Number(event.target.value) })}
          required
        />
      </div>

      <div className="grid gap-4 md:grid-cols-[1fr_220px]">
        <Input
          label="Seed"
          type="number"
          value={seed}
          onChange={(event: ChangeEvent<HTMLInputElement>) => regenerate({ seed: Number(event.target.value) })}
          required
        />
        <Button
          type="button"
          variant="outline"
          className="self-end"
          onClick={() => regenerate({ seed: Math.floor(Math.random() * 100_000) })}
        >
          Seed aleatoria
        </Button>
      </div>

      <label className="flex flex-col gap-1 text-sm font-medium" htmlFor="snakes-ai-difficulty">
        Dificultad de IA
        <select
          id="snakes-ai-difficulty"
          className="min-h-[44px] rounded-lg border border-[var(--color-border)] bg-[var(--color-background)] px-3"
          value={value.ai_config?.difficulty ?? 'MEDIUM'}
          onChange={(event) => onChange({ ...previewValue, ai_config: { difficulty: event.currentTarget.value as 'EASY' | 'MEDIUM' | 'HARD' } })}
        >
          <option value="EASY">Fácil</option>
          <option value="MEDIUM">Media</option>
          <option value="HARD">Difícil</option>
        </select>
      </label>

      <div className="mt-8 border-t pt-6">
        <div className="flex justify-between items-center mb-4">
          <div>
            <h4 className="text-xl font-bold">Preguntas de Cultura General</h4>
            <p className="text-sm text-[var(--color-muted)]">
              Cada pregunta debe tener exactamente dos opciones. Mínimo {MIN_QUESTIONS} preguntas ({questions.length}/{MIN_QUESTIONS}).
            </p>
          </div>
          <Button onClick={addQuestion}>+ Agregar Pregunta</Button>
        </div>
        <div className="space-y-6">
          {questions.map((q, idx) => (
            <div key={idx} className="p-4 border rounded-lg bg-gray-50 relative">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => removeQuestion(idx)}
                disabled={questions.length <= MIN_QUESTIONS}
                className="absolute top-2 right-2 text-red-500 hover:bg-red-50"
                aria-label="Eliminar pregunta"
                title="Eliminar pregunta"
              >X</Button>
              <div className="mb-4 pr-8">
                <Input
                  label={`Pregunta ${idx + 1}`}
                  value={q.question}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => updateQuestion(idx, { ...q, question: e.target.value })}
                  className="w-full"
                />
              </div>
              <div className="space-y-2">
                <p className="font-semibold text-sm">Opciones (marca la correcta):</p>
                {q.options.map((opt: string, optIdx: number) => (
                  <div key={optIdx} className="flex gap-2 items-center">
                    <label className="flex items-center gap-1 p-2 cursor-pointer">
                      <input
                        type="radio"
                        name={`correct-${idx}`}
                        checked={q.correct_index === optIdx}
                        onChange={() => updateQuestion(idx, { ...q, correct_index: optIdx })}
                      />
                      <span className="sr-only">Marcar opción {optIdx + 1} como correcta</span>
                    </label>
                    <Input
                      label=""
                      value={opt}
                      onChange={(e: ChangeEvent<HTMLInputElement>) => {
                        const newOpts = [...q.options];
                        newOpts[optIdx] = e.target.value;
                        updateQuestion(idx, { ...q, options: newOpts });
                      }}
                    />
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
      <SnakeLadderBoardPreview value={previewValue} />
    </div>
  );
}

function normalizeContent(
  current: SnakeLadderContent,
  boardWidth: number,
  boardHeight: number,
  snakeCount: number,
  ladderCount: number,
  seed: number,
): SnakeLadderContent {
  const totalCells = boardWidth * boardHeight;
  const links = generateSnakeLadderLinks({ boardWidth, boardHeight, snakeCount, ladderCount, seed });
  return {
    ...current,
    board_width: boardWidth,
    board_height: boardHeight,
    start_position: 1,
    end_position: totalCells,
    seed,
    snakes: links.snakes,
    ladders: links.ladders,
    ai_config: current.ai_config ?? { difficulty: 'MEDIUM' },
    questions: current.questions ?? [],
  };
}

function createEmptyQuestion(): SnakeQuestion {
  return { question: '', options: ['', ''], correct_index: 0 };
}

function clampCount(value: number, max: number): number {
  const normalized = Number.isFinite(value) ? Math.trunc(value) : 0;
  return Math.max(0, Math.min(max, normalized));
}
