import { z } from "zod";
import { SnakesSchema } from "./snakes";

// F10.7 (plan/05_Contenido_maker_y_juego.md §4): este archivo es una versión
// CURADA de ../usbi/frontend/packages/schema/index.ts, no una copia
// verbatim. Se retiraron a propósito RegisterSchema, LoginSchema,
// TutorConsentSchema, ArcoSchema, LevelAttemptItemSchema, SyncPayloadSchema
// y SyncEventSchema: son esquemas de identidad/PII (correo, nombre completo,
// correo de tutor) que pertenecen al modelo de ../usbi que este proyecto
// reemplazó por el cuestionario de gustos — traerlos de vuelta reintroduciría
// exactamente lo que el rediseño de identidad quitó. Ninguno se usa desde
// ningún componente de contenido/juego (verificado por grep en ../usbi antes
// de portar). También se retira DragAndDropSchema: no es una de las siete
// plantillas de `levels.template_type` en el esquema SQL y no tiene
// formulario ni motor asociado en ningún lado — nunca se llegó a usar.
//
// Lo que queda es exactamente lo que consumen frontend/src/features/content/
// y frontend/src/features/games/: el contrato de contenido de las siete
// plantillas, más los tipos que usa el maker local para su propio formato de
// exportación (LevelMetadataSchema, LevelExportSchema) — ver plan/05 §7
// sobre por qué ese formato NO es el mismo que acepta POST /levels.

export const MultipleChoiceSchema = z.object({
  question: z.string().min(1),
  options: z.array(z.string()).min(2).max(4),
  correct_index: z.number().int().min(0).max(3),
  media_url: z.string().optional(),
});

export const MemoryPairSchema = z.object({
  id: z.string(),
  content1: z.string(),
  content2: z.string(),
  color: z.string().optional(),
});

export const MemorySchema = z.object({
  back_color: z.string().optional(),
  pairs: z.array(MemoryPairSchema).min(4), // Specification says min 4 pairs
});

export const FakeNewsItemSchema = z.object({
  title: z.string().min(1),
  content: z.string().min(1),
  isFake: z.boolean(),
  explanation: z.string().optional(),
  imageUrl: z.string().optional(),
  reference: z.string(),
});

export const FakeNewsSchema = z.object({
  news: z.array(FakeNewsItemSchema).min(1),
});

export const CrosswordWordSchema = z.object({
  word: z.string().min(2),
  clue: z.string().min(1),
});

export const CrosswordSchema = z.object({
  words: z.array(CrosswordWordSchema).min(2),
});

export const WordSearchSchema = z.object({
  words: z.array(z.string().min(2)).min(2),
  width: z.number().int().min(5).max(24).optional(),
  height: z.number().int().min(5).max(24).optional(),
  seed: z.number().int().optional(),
});

export const PuzzleSchema = z.object({
  phrase: z.string().min(1),
  pieces: z.number().int().min(3).max(20).default(3),
  seed: z.number().int().optional(),
});

export const TemplateTypeSchema = z.enum([
  "trivia",
  "puzzle",
  "word_search",
  "fake_news",
  "crossword",
  "memory",
  "snakes_ladders",
]);

// Forma del JSON que exporta/acepta el maker LOCAL (frontend/src/features/
// maker/MakerPage.tsx), no la que acepta la API. author/creation_date/id no
// existen como columnas en `levels` — plan/05 §7 documenta la traducción.
export const LevelMetadataSchema = z.object({
  id: z.string().uuid(),
  title: z.string().min(1),
  author: z.string().min(1),
  creation_date: z.string(),
  color: z.string().min(1),
  difficulty: z.number().int().min(1).max(10),
  template_type: TemplateTypeSchema,
});

export const LevelExportSchema = z.object({
  metadata: LevelMetadataSchema,
  content: z.union([
    z.array(MultipleChoiceSchema),
    MemorySchema,
    FakeNewsSchema,
    CrosswordSchema,
    WordSearchSchema,
    PuzzleSchema,
    SnakesSchema,
  ]),
});

export type MultipleChoice = z.infer<typeof MultipleChoiceSchema>;
export type MemoryPair = z.infer<typeof MemoryPairSchema>;
export type Memory = z.infer<typeof MemorySchema>;
export type FakeNewsItem = z.infer<typeof FakeNewsItemSchema>;
export type FakeNews = z.infer<typeof FakeNewsSchema>;
export type CrosswordWord = z.infer<typeof CrosswordWordSchema>;
export type Crossword = z.infer<typeof CrosswordSchema>;
export type WordSearch = z.infer<typeof WordSearchSchema>;
export type Puzzle = z.infer<typeof PuzzleSchema>;
export { SnakesSchema, type Snakes } from "./snakes";
export type TemplateType = z.infer<typeof TemplateTypeSchema>;
export type LevelMetadata = z.infer<typeof LevelMetadataSchema>;
export type LevelExport = z.infer<typeof LevelExportSchema>;
