import { z } from 'zod';

// Espejan levels.SectionResponse / levels.LevelResponse / levels.LevelSummary
// / levels.LevelsPage / levels.ArchivedSectionsResponse /
// levels.ArchivedLevelsResponse en Go (backend/internal/levels/dto.go).

export const TemplateTypeSchema = z.enum([
  'trivia',
  'puzzle',
  'word_search',
  'fake_news',
  'crossword',
  'memory',
  'snakes_ladders',
]);

export const SectionDTOSchema = z.object({
  id: z.string().uuid(),
  title: z.string(),
  description: z.string(),
  color: z.string(),
  is_published: z.boolean(),
  created_by_admin_id: z.string().uuid().optional(),
  created_at: z.string().optional(),
});
export type SectionDTO = z.infer<typeof SectionDTOSchema>;

export const SectionsResponseSchema = z.object({
  items: z.array(SectionDTOSchema),
});

export const ArchivedSectionsResponseSchema = z.object({
  items: z.array(SectionDTOSchema),
});

export const LevelSummaryDTOSchema = z.object({
  id: z.string().uuid(),
  section_id: z.string().uuid(),
  title: z.string(),
  color: z.string(),
  template_type: TemplateTypeSchema,
  difficulty: z.number(),
  is_published: z.boolean(),
  created_at: z.string(),
});
export type LevelSummaryDTO = z.infer<typeof LevelSummaryDTOSchema>;

export const LevelDTOSchema = LevelSummaryDTOSchema.extend({
  content: z.unknown(),
  created_by_admin_id: z.string().uuid().optional(),
  updated_at: z.string().optional(),
});
export type LevelDTO = z.infer<typeof LevelDTOSchema>;

export const LevelsPageDTOSchema = z.object({
  items: z.array(LevelSummaryDTOSchema),
  next_cursor: z.string().optional(),
});

export const ArchivedLevelsResponseSchema = z.object({
  items: z.array(LevelSummaryDTOSchema),
});
