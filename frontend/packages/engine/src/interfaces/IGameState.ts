export interface IGameState {
  status: 'IDLE' | 'PLAYING' | 'GAME_OVER';
  score: number;
  maxScore: number;
  timeLeft?: number;
}
