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

// Espejan levels.BadgeResponse / levels.CompleteLevelResponse /
// levels.ProfileProgressResponse (F10.10, plan/05 §8).
export const BadgeSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  xp_threshold: z.number(),
  icon_key: z.string(),
  earned_at: z.string(),
});

export const CompleteLevelResponseSchema = z.object({
  level_id: z.string().uuid(),
  completed: z.boolean(),
  attempt_number: z.number(),
  xp_awarded: z.number(),
  total_xp: z.number(),
  current_streak: z.number(),
  badges_awarded: z.array(BadgeSchema),
});

export const ProfileProgressResponseSchema = z.object({
  total_xp: z.number(),
  completed_levels: z.number(),
  total_attempts: z.number(),
  current_streak: z.number(),
  badges: z.array(BadgeSchema),
  levels: z.array(z.object({
    level_id: z.string().uuid(),
    title: z.string(),
    template_type: TemplateTypeSchema,
    difficulty: z.number(),
    best_score: z.number(),
    xp_total_for_level: z.number(),
    attempts_count: z.number(),
    first_completed_at: z.string().optional(),
    last_completed_at: z.string().optional(),
  })),
});
