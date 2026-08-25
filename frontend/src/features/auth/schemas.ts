import { z } from 'zod';

// ── Validación anti-inyección de answer_text ─────────────────────────────────
// Duplica en el frontend las tres reglas de backend/internal/auth/validation.go
// (validateAnswerText) — el frontend nunca es la única barrera
// (plan/04_Rediseno_identidad_gustos.md §5): un cliente que se salte esta
// validación igual choca con la misma regla en Go.
// Intencional: espeja backend/internal/auth/validation.go (validateAnswerText, r < 0x20).
// eslint-disable-next-line no-control-regex
const CONTROL_CHAR_RE = /[\x00-\x1f]/;
const JSON_LOOKALIKE_RE = /^[[{]/;
const HTML_TAG_RE = /<[a-z][\s\S]*>/i;

export const AnswerTextSchema = z
  .string()
  .trim()
  .min(1, 'Escribe una respuesta.')
  .max(200, 'Máximo 200 caracteres.')
  .refine((v) => !CONTROL_CHAR_RE.test(v), 'La respuesta no puede contener caracteres de control.')
  .refine((v) => !JSON_LOOKALIKE_RE.test(v), 'La respuesta no puede empezar con [ o {.')
  .refine((v) => !HTML_TAG_RE.test(v), 'La respuesta no puede contener etiquetas HTML.');

export const NicknameSchema = z
  .string()
  .regex(/^[a-z0-9]{6,20}$/, 'Debe tener entre 6 y 20 letras minúsculas o dígitos.');

// ── Registro en 3 pasos (contrato de POST /auth/register/*, plan/04 §3) ─────

export const QuestionOptionSchema = z.object({
  id: z.string().uuid(),
  text: z.string(),
});

export const RegisterQuestionsResponseSchema = z.object({
  questions: z.array(QuestionOptionSchema),
  max_questions_shown: z.number().int(),
});

export const AnswerInputSchema = z.object({
  question_id: z.string().uuid(),
  answer_text: AnswerTextSchema,
});

export const RegisterAnswersResponseSchema = z.object({
  registration_token: z.string().min(1),
  nickname_candidates: z.array(z.string()),
});

export const RegisterConfirmResponseSchema = z.object({
  account_id: z.string().uuid(),
  nickname: z.string(),
  password: z.string(),
  display_alias: z.string(),
});

// ── Login ─────────────────────────────────────────────────────────────────

export const LoginSchema = z.object({
  nickname: z.string().trim().min(1, 'Ingresa tu nickname.'),
  password: z.string().min(1, 'Ingresa tu contraseña.'),
});
