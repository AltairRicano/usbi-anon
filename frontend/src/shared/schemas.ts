import { z } from 'zod';

// Espejan domain.User / domain.UserRole / domain.UserStatus en Go
// (backend/internal/domain). Sin full_name/email: el DTO nunca incluye dato
// identificable directo, solo UUID + nickname generado (plan/04 §1).
export const UserRoleSchema = z.enum(['player', 'admin']);
export const UserStatusSchema = z.enum(['active', 'suspended', 'deleted']);

export const UserSchema = z.object({
  id: z.string().uuid(),
  nickname: z.string(),
  display_alias: z.string(),
  is_adult: z.boolean(),
  role: UserRoleSchema,
  status: UserStatusSchema,
  created_at: z.string(),
});
export type User = z.infer<typeof UserSchema>;

/** POST /auth/login y POST /auth/refresh devuelven esta misma forma
 * (auth.LoginResponse en Go). */
export const AuthResponseSchema = z.object({
  access_token: z.string().min(1),
  refresh_token: z.string(),
  token_type: z.string(),
  access_token_expires_in: z.number().int(),
  refresh_token_expires_at: z.string(),
  user: UserSchema,
});
export type AuthResponse = z.infer<typeof AuthResponseSchema>;
