import { z } from 'zod';

// Espeja badges.BadgeResponse en Go (backend/internal/badges/dto.go).
export const BadgeSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  xp_threshold: z.number().int(),
  icon_key: z.string(),
});
export type Badge = z.infer<typeof BadgeSchema>;

// C4 (estado_proyecto.md 2026-09-10): GET /admin/badges responde
// {"items": […]} desde ese cambio — antes era un arreglo JSON crudo.
export const BadgesResponseSchema = z.object({
  items: z.array(BadgeSchema),
});
