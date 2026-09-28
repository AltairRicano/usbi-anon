import { z } from 'zod';
import { UserRoleSchema } from '../../shared/schemas';

// Espejan auth.AdminAccountResponse / auth.AdminQuizAnswersResponse /
// auth.AdminResetPasswordResponse en Go (backend/internal/auth/dto.go).
export const AdminAccountResponseSchema = z.object({
  id: z.string().uuid(),
  nickname: z.string(),
  role: UserRoleSchema,
  display_alias: z.string(),
  created_at: z.string(),
});
export type AdminAccountResponse = z.infer<typeof AdminAccountResponseSchema>;

export const AdminQuizAnswerSchema = z.object({
  question_text_snapshot: z.string(),
  answer_text: z.string(),
  created_at: z.string(),
});

export const AdminQuizAnswersResponseSchema = z.object({
  items: z.array(AdminQuizAnswerSchema),
});

export const AdminResetPasswordResponseSchema = z.object({
  new_password: z.string(),
});
