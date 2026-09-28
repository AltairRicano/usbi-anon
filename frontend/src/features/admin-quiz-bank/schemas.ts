import { z } from 'zod';

// Espejan quiz.QuestionResponse / quiz.QuestionsResponse en Go
// (backend/internal/quiz/bank.go).
export const RegistrationQuestionSchema = z.object({
  id: z.string().uuid(),
  question_text: z.string(),
  is_active: z.boolean(),
  display_order: z.number().int(),
  created_at: z.string(),
  updated_at: z.string(),
});
export type RegistrationQuestion = z.infer<typeof RegistrationQuestionSchema>;

export const RegistrationQuestionsResponseSchema = z.object({
  items: z.array(RegistrationQuestionSchema),
});

// El backend no expone GET /admin/registration-settings (solo PUT, ver
// plan/04 §3) — el valor actual de max_questions_shown se lee de la
// respuesta de POST /auth/register/questions, el único endpoint que lo
// devuelve.
export const CurrentSettingsSchema = z.object({
  max_questions_shown: z.number().int(),
});
