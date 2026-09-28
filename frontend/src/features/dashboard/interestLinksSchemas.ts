import { z } from 'zod';

// Espeja interestlinks.CategoryWithLinks/InterestLinksResponse en Go
// (backend/internal/interestlinks/dto.go) — la forma agrupada que devuelve
// GET /interest-links para cualquier cuenta autenticada (jugador o admin).
// C4 (estado_proyecto.md 2026-09-10): responde {"items": […]} desde ese
// cambio, antes era un arreglo JSON crudo.
const InterestLinkCardSchema = z.object({
  id: z.string().uuid(),
  category_id: z.string().uuid(),
  title: z.string(),
  description: z.string(),
  color: z.string(),
  url: z.string(),
});

const CategoryWithLinksSchema = z.object({
  category: z.object({
    id: z.string().uuid(),
    name: z.string(),
    display_order: z.number().int(),
  }),
  links: z.array(InterestLinkCardSchema),
});

export const InterestLinksResponseSchema = z.object({
  items: z.array(CategoryWithLinksSchema),
});
export type CategoryWithLinks = z.infer<typeof CategoryWithLinksSchema>;

// Espeja suggestions.SubmitResponse en Go.
export const SubmitSuggestionResponseSchema = z.object({
  id: z.string().uuid(),
  submitted_at: z.string(),
});
