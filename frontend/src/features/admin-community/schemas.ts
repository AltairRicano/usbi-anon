import { z } from 'zod';

// Espeja interestlinks.CategoryResponse/LinkResponse/CategoriesResponse/
// LinksResponse en Go (backend/internal/interestlinks/dto.go). C4
// (estado_proyecto.md 2026-09-10): ambos listados responden {"items": […]}.
export const InterestLinkCategorySchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  display_order: z.number().int(),
  created_at: z.string(),
  updated_at: z.string(),
});
export type InterestLinkCategory = z.infer<typeof InterestLinkCategorySchema>;

export const CategoriesResponseSchema = z.object({
  items: z.array(InterestLinkCategorySchema),
});

export const InterestLinkSchema = z.object({
  id: z.string().uuid(),
  category_id: z.string().uuid(),
  title: z.string(),
  description: z.string(),
  color: z.string(),
  url: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});
export type InterestLink = z.infer<typeof InterestLinkSchema>;

export const LinksResponseSchema = z.object({
  items: z.array(InterestLinkSchema),
});

// Espeja suggestions.SuggestionResponse/Page en Go
// (backend/internal/suggestions/dto.go).
export const SuggestionSchema = z.object({
  id: z.string().uuid(),
  description: z.string(),
  levels_completed_snapshot: z.number().int(),
  xp_snapshot: z.number().int(),
  submitted_at: z.string(),
});
export type Suggestion = z.infer<typeof SuggestionSchema>;

export const SuggestionsPageSchema = z.object({
  items: z.array(SuggestionSchema),
  next_cursor: z.string().optional(),
});
