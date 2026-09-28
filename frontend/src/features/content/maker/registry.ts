import { z } from "zod";
import { 
  CrosswordSchema, 
  FakeNewsSchema, 
  MemorySchema, 
  PuzzleSchema, 
  SnakesSchema, 
  WordSearchSchema, 
} from '@usbi/schema';

// Locally define the MultipleChoice schema to avoid bundler/circular initialization issues
const LocalMultipleChoiceSchema = z.object({
  question: z.string().min(1),
  options: z.array(z.string()).min(2).max(4),
  correct_index: z.number().int().min(0).max(3),
  media_url: z.string().optional(),
});

import { TriviaForm } from './forms/TriviaForm';
import { CrosswordForm } from './forms/CrosswordForm';
import { WordSearchForm } from './forms/WordSearchForm';
import { PuzzleForm } from './forms/PuzzleForm';
import { FakeNewsForm } from './forms/FakeNewsForm';
import { MemoryForm } from './forms/MemoryForm';
import { SnakeLadderForm } from './forms/SnakeLadderForm';

import { TriviaPreview } from './previews/TriviaPreview';
import { CrosswordPreview } from './previews/CrosswordPreview';
import { WordSearchPreview } from './previews/WordSearchPreview';
import { PuzzlePreview } from './previews/PuzzlePreview';
import { FakeNewsPreview } from './previews/FakeNewsPreview';
import { MemoryPreview } from './previews/MemoryPreview';
import { SnakeLadderPreview } from './previews/SnakeLadderPreview';
import { generateSnakeLadderLinks } from './snakesLayout';
import { canBuildConnectedCrossword, createMemoryPairs, DEFAULT_MEMORY_BACK_COLOR, normalizeCrosswordAnswer } from '@usbi/engine';

const triviaDefaults = () => [
  { question: '', options: ['', ''], correct_index: 0 },
  { question: '', options: ['', ''], correct_index: 0 },
  { question: '', options: ['', ''], correct_index: 0 },
];

const snakeLadderDefaults = () => {
  const boardWidth = 6;
  const boardHeight = 6;
  const seed = 2026;
  const links = generateSnakeLadderLinks({ boardWidth, boardHeight, snakeCount: 3, ladderCount: 3, seed });
  return {
    board_width: boardWidth,
    board_height: boardHeight,
    start_position: 1,
    end_position: boardWidth * boardHeight,
    seed,
    snakes: links.snakes,
    ladders: links.ladders,
    ai_config: { difficulty: 'MEDIUM' },
    questions: Array.from({ length: 8 }, () => ({ question: '', options: ['', ''], correct_index: 0 })),
  };
};

const MakerCrosswordSchema = CrosswordSchema.superRefine((content, ctx) => {
  const answers = content.words.map((word) => normalizeCrosswordAnswer(word.word));
  const hasDuplicates = new Set(answers).size !== answers.length;

  if (hasDuplicates) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ['words'],
      message: 'Las respuestas del crucigrama no pueden repetirse.',
    });
  }

  if (!canBuildConnectedCrossword(content.words)) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ['words'],
      message: 'Las palabras deben compartir letras suficientes para generar un crucigrama conectado.',
    });
  }
});

export const levelTemplateRegistry = {
  trivia: {
    schema: z.array(LocalMultipleChoiceSchema).min(3),
    FormComponent: TriviaForm,
    PreviewComponent: TriviaPreview,
    getDefaults: triviaDefaults,
  },
  crossword: {
    schema: MakerCrosswordSchema,
    FormComponent: CrosswordForm,
    PreviewComponent: CrosswordPreview,
    getDefaults: () => ({ words: [{ word: '', clue: '' }, { word: '', clue: '' }] }),
  },
  word_search: {
    schema: WordSearchSchema,
    FormComponent: WordSearchForm,
    PreviewComponent: WordSearchPreview,
    getDefaults: () => ({ words: ['', ''], width: 12, height: 12, seed: 1234 }),
  },
  puzzle: {
    schema: PuzzleSchema,
    FormComponent: PuzzleForm,
    PreviewComponent: PuzzlePreview,
    getDefaults: () => ({ phrase: '', pieces: 3, seed: 1234 }),
  },
  fake_news: {
    schema: FakeNewsSchema,
    FormComponent: FakeNewsForm,
    PreviewComponent: FakeNewsPreview,
    getDefaults: () => ({ news: [{ title: '', content: '', isFake: false, reference: '' }] }),
  },
  memory: {
    schema: MemorySchema,
    FormComponent: MemoryForm,
    PreviewComponent: MemoryPreview,
    getDefaults: () => ({ back_color: DEFAULT_MEMORY_BACK_COLOR, pairs: createMemoryPairs(4) }),
  },
  snakes_ladders: {
    schema: SnakesSchema,
    FormComponent: SnakeLadderForm,
    PreviewComponent: SnakeLadderPreview,
    getDefaults: snakeLadderDefaults,
  }
} as const;
