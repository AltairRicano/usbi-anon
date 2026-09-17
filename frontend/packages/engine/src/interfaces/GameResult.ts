export interface GameResult {
  completed: boolean;
  score: number;
  maxScore: number;
  /** Payload específico de plantilla para que el servidor verifique el resultado (Fase B de M1). Ausente si la plantilla no es verificable. */
  answers?: unknown;
}
