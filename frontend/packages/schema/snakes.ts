import { z } from "zod";

export const SnakeLadderItemSchema = z.object({
  start: z.number().int().min(1),
  end: z.number().int().min(1),
});

export const SnakesSchema = z.object({
  board_width: z.number().int().min(1),
  board_height: z.number().int().min(1),
  start_position: z.number().int().min(1),
  end_position: z.number().int().min(1),
  seed: z.number().int().optional(),
  snakes: z.array(SnakeLadderItemSchema).optional(),
  ladders: z.array(SnakeLadderItemSchema).optional(),
  ai_config: z.object({
    difficulty: z.enum(["EASY", "MEDIUM", "HARD"]),
    fail_probability: z.number().min(0).max(1).optional(),
    weights: z.array(z.number().min(0)).optional()
  }).optional(),
  // Gates every dice roll behind a two-option question drawn from a queue
  // (see SnakeLadderGame); a minimum bank of 8 keeps the queue from cycling
  // through the same handful of questions too fast during a single match.
  questions: z.array(z.object({
    question: z.string().min(1),
    options: z.array(z.string()).length(2),
    correct_index: z.number().int().min(0)
  })).min(8),
}).superRefine((data, ctx) => {
  const total_cells = data.board_width * data.board_height;
  if (data.end_position > total_cells) {
    ctx.addIssue({ code: z.ZodIssueCode.custom, message: "End position out of range" });
  }
  
  const origins = new Set<number>();

  if (data.snakes) {
    for (const s of data.snakes) {
      if (s.start <= s.end) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: "Snake must go down" });
      }
      if (origins.has(s.start)) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: "Duplicate origin" });
      }
      origins.add(s.start);
    }
  }

  if (data.ladders) {
    for (const l of data.ladders) {
      if (l.start >= l.end) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: "Ladder must go up" });
      }
      if (origins.has(l.start)) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: "Duplicate origin" });
      }
      origins.add(l.start);
    }
  }

  if (data.questions) {
    data.questions.forEach((question, questionIndex) => {
      if (question.correct_index >= question.options.length) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ["questions", questionIndex, "correct_index"],
          message: "Correct option index out of range",
        });
      }
      question.options.forEach((option, optionIndex) => {
        if (option.trim().length === 0) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            path: ["questions", questionIndex, "options", optionIndex],
            message: "Option cannot be empty",
          });
        }
      });
    });
  }
});
export type Snakes = z.infer<typeof SnakesSchema>;
