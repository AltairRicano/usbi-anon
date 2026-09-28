import { lazy, Suspense, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';
import type { GameResult } from '@usbi/engine';
import { HomeButton } from '../../shared/components/ui/HomeButton';
import { LinkButton } from '../../shared/components/ui/LinkButton';
import { Button } from '../../shared/components/ui/Button';
import { apiClient } from '../../shared/apiClient';
import { errorMessage } from '../../shared/errorMessage';
import { CompleteLevelResponseSchema, LevelDTOSchema } from './schemas';
import type { CompleteLevelResponse, LevelDTO } from './types';
import {
  normalizeCrosswordContent,
  normalizeFakeNewsContent,
  normalizeMemoryBackColorContent,
  normalizeMemoryContent,
  normalizePuzzleContent,
  normalizeSnakesContent,
  normalizeTriviaContent,
  normalizeWordSearchContent,
  templateTypeLabel,
} from './types';

const TriviaGame = lazy(() => import('../games/TriviaGame').then((mod) => ({ default: mod.TriviaGame })));
const MemoryGame = lazy(() => import('../games/components/MemoryGame').then((mod) => ({ default: mod.MemoryGame })));
const FakeNewsGame = lazy(() => import('../games/components/FakeNewsGame').then((mod) => ({ default: mod.FakeNewsGame })));
const WordSearchGame = lazy(() => import('../games/WordSearchGame').then((mod) => ({ default: mod.WordSearchGame })));
const PuzzleGame = lazy(() => import('../games/PuzzleGame').then((mod) => ({ default: mod.PuzzleGame })));
const CrosswordGame = lazy(() => import('../games/CrosswordGame').then((mod) => ({ default: mod.CrosswordGame })));
const SnakeLadderGame = lazy(() => import('../games/components/SnakeLadderGame').then((mod) => ({ default: mod.SnakeLadderGame })));

export function OfficialLevelPage() {
  const { levelId } = useParams();
  const [level, setLevel] = useState<LevelDTO | null>(null);
  const [levelError, setLevelError] = useState(false);
  const [result, setResult] = useState<CompleteLevelResponse | null>(null);
  const [gameResult, setGameResult] = useState<GameResult | null>(null);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [attemptKey, setAttemptKey] = useState(0);
  const submittedRef = useRef(false);

  useEffect(() => {
    if (!levelId) return;
    let cancelled = false;
    (async () => {
      try {
        const { data } = await apiClient.get(`/levels/${levelId}`);
        if (!cancelled) setLevel(LevelDTOSchema.parse(data));
      } catch {
        if (!cancelled) setLevelError(true);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [levelId]);

  // Memoized on level.content: OfficialLevelPage re-renders whenever finishLevel
  // resolves (setResult/setSaveError), and a fresh array/object reference here
  // would make the game components below think their content prop changed,
  // tearing down and losing their in-progress/finished engine state.
  // These must run before any early return below — Hooks can't be called
  // conditionally, and a component that calls fewer hooks on one render than
  // another (e.g. skipping them while `level` is still loading) crashes React.
  const triviaQuestions = useMemo(
    () => (level && level.template_type === 'trivia' ? normalizeTriviaContent(level.content) : []),
    [level],
  );
  const memoryPairs = useMemo(
    () => (level && level.template_type === 'memory' ? normalizeMemoryContent(level.content) : []),
    [level],
  );
  const memoryBackColor = useMemo(
    () => (level && level.template_type === 'memory' ? normalizeMemoryBackColorContent(level.content) : undefined),
    [level],
  );
  const fakeNews = useMemo(
    () => (level && level.template_type === 'fake_news' ? normalizeFakeNewsContent(level.content) : []),
    [level],
  );
  const wordSearch = useMemo(
    () => (level && level.template_type === 'word_search' ? normalizeWordSearchContent(level.content) : null),
    [level],
  );
  const puzzle = useMemo(
    () => (level && level.template_type === 'puzzle' ? normalizePuzzleContent(level.content) : null),
    [level],
  );
  const crosswordWords = useMemo(
    () => (level && level.template_type === 'crossword' ? normalizeCrosswordContent(level.content) : []),
    [level],
  );
  const snakes = useMemo(
    () => (level && level.template_type === 'snakes_ladders' ? normalizeSnakesContent(level.content) : null),
    [level],
  );

  const finishLevel = useCallback(async (gameResult: GameResult) => {
    if (!levelId || submittedRef.current || !level?.is_published) return;
    submittedRef.current = true;
    setGameResult(gameResult);
    setSaveError(null);
    setIsSubmitting(true);
    try {
      const { data } = await apiClient.post(`/levels/${levelId}/complete`, {
        score: gameResult.score,
        completed: gameResult.completed,
        answers: gameResult.answers,
        client_finished_at: new Date().toISOString(),
      });
      setResult(CompleteLevelResponseSchema.parse(data));
    } catch (err) {
      submittedRef.current = false;
      setSaveError(errorMessage(err, 'No se pudo guardar el resultado en línea.'));
    } finally {
      setIsSubmitting(false);
    }
  }, [level?.is_published, levelId]);

  const retryLevel = useCallback(() => {
    submittedRef.current = false;
    setResult(null);
    setGameResult(null);
    setSaveError(null);
    setAttemptKey((key) => key + 1);
  }, []);

  if (levelError) {
    return (
      <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
        <div className="mx-auto max-w-3xl rounded-lg bg-[--color-card] p-5">
          <p className="text-[--color-error]">No se pudo cargar el nivel.</p>
          <HomeButton className="mt-4" />
        </div>
      </main>
    );
  }

  if (!level) {
    return <main className="min-h-screen p-6">Cargando nivel...</main>;
  }

  return (
    <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
      <div className="mx-auto max-w-4xl space-y-5">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">{level.title}</h1>
            <p className="text-sm text-[--color-muted]">Dificultad {level.difficulty} · {templateTypeLabel(level.template_type)}</p>
          </div>
          <HomeButton />
        </header>

        {isSubmitting && (
          <section className="rounded-lg bg-[--color-card] p-5 shadow-sm border border-[--color-border] flex justify-center items-center h-32">
            <p className="text-[--color-muted] animate-pulse">Guardando resultados...</p>
          </section>
        )}

        {result && (
          <section className="rounded-lg bg-[--color-card] p-5 shadow-sm border border-[--color-border]">
            <h2 className="text-xl font-semibold">
              {result.completed ? '¡Nivel superado!' : 'Nivel no superado'}
            </h2>
            {gameResult && (
              <p className="text-sm text-[--color-muted]">
                {result.completed
                  ? `Puntuación: ${gameResult.score} de ${gameResult.maxScore}.`
                  : `Puntuación: ${gameResult.score} de ${gameResult.maxScore}, necesitas más para superarlo.`}
              </p>
            )}
            {result.completed && (
              <p className="text-sm text-[--color-muted]">
                XP otorgada: {result.xp_awarded} · intento {result.attempt_number} · XP total: {result.total_xp} · racha: {result.current_streak}
              </p>
            )}
            {(result.badges_awarded ?? []).length > 0 && (
              <div className="mt-3 flex flex-wrap gap-2">
                {(result.badges_awarded ?? []).map((badge) => (
                  <span key={badge.id} className="rounded-full border border-[--color-border] px-3 py-1 text-sm">
                    {badge.name}
                  </span>
                ))}
              </div>
            )}
            <div className="mt-4 flex flex-wrap gap-3">
              <Button variant="primary" onClick={retryLevel}>Volver a jugar</Button>
              <LinkButton to="/perfil">Ver progreso</LinkButton>
            </div>
          </section>
        )}

        {!level.is_published && (
          <section className="rounded-lg bg-[--color-card] p-4 text-sm text-[--color-muted] shadow-sm">
            Vista previa de borrador. Este intento no modifica XP ni progreso oficial.
          </section>
        )}
        {saveError && <p className="rounded border border-[--color-error] bg-[--color-card] p-3 text-[--color-error]">{saveError}</p>}

        {!result && !isSubmitting && (
        <Suspense fallback={<GameFallback />}>
          {level.template_type === 'trivia' && triviaQuestions.length > 0 && (
            <TriviaGame key={attemptKey} questions={triviaQuestions} onFinish={finishLevel} />
          )}
          {level.template_type === 'memory' && memoryPairs.length >= 2 && (
            <MemoryGame key={attemptKey} pairs={memoryPairs} backColor={memoryBackColor} onFinish={finishLevel} />
          )}
          {level.template_type === 'fake_news' && fakeNews.length > 0 && (
            <FakeNewsGame key={attemptKey} news={fakeNews} onFinish={finishLevel} />
          )}
          {level.template_type === 'word_search' && wordSearch && wordSearch.words.length > 0 && (
            <WordSearchGame
              key={attemptKey}
              words={wordSearch.words}
              width={wordSearch.width}
              height={wordSearch.height}
              seed={wordSearch.seed}
              onFinish={finishLevel}
            />
          )}
          {level.template_type === 'puzzle' && puzzle && (
            <PuzzleGame key={attemptKey} phrase={puzzle.phrase} pieces={puzzle.pieces} seed={puzzle.seed} onFinish={finishLevel} />
          )}
          {level.template_type === 'crossword' && crosswordWords.length >= 2 && (
            <CrosswordGame key={attemptKey} words={crosswordWords} onFinish={finishLevel} />
          )}
          {level.template_type === 'snakes_ladders' && snakes && (
            <SnakeLadderGame key={attemptKey} level={snakes} onFinish={finishLevel} />
          )}
        </Suspense>
        )}
        {!hasPlayableContent(level) && (
          <section className="rounded-lg bg-[--color-card] p-5 shadow-sm">
            <p className="text-[--color-muted]">El contenido de este nivel no cumple el contrato mínimo de su plantilla.</p>
          </section>
        )}
      </div>
    </main>
  );
}

function GameFallback() {
  return (
    <section className="rounded-lg bg-[--color-card] p-5 shadow-sm">
      <p className="text-[--color-muted]">Cargando juego...</p>
    </section>
  );
}

function hasPlayableContent(level: LevelDTO): boolean {
  switch (level.template_type) {
    case 'trivia':
      return normalizeTriviaContent(level.content).length > 0;
    case 'memory':
      return normalizeMemoryContent(level.content).length >= 2;
    case 'fake_news':
      return normalizeFakeNewsContent(level.content).length > 0;
    case 'word_search':
      return normalizeWordSearchContent(level.content).words.length > 0;
    case 'puzzle':
      return normalizePuzzleContent(level.content) !== null;
    case 'crossword':
      return normalizeCrosswordContent(level.content).length >= 2;
    case 'snakes_ladders':
      return normalizeSnakesContent(level.content) !== null;
  }
}
