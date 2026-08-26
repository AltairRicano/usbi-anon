import { weightedRandom } from './WeightedRandom';

export type GameState = 'idle' | 'starting' | 'player_turn' | 'player_rolling' | 'player_moving' | 'resolving_player_tile' | 'ai_turn' | 'ai_thinking' | 'ai_rolling' | 'ai_moving' | 'resolving_ai_tile' | 'game_over';

export interface SnakeLadderState {
  state: GameState;
  playerPosition: number;
  aiPosition: number;
  winner: 'player' | 'ai' | null;
  lastRoll: number;
  message: string;
}

export interface SnakeLadderConfig {
  boardSize: number;
  boardWidth?: number;
  boardHeight?: number;
  startPosition?: number;
  endPosition?: number;
  snakes: {start: number, end: number}[];
  ladders: {start: number, end: number}[];
  aiDifficulty: 'EASY' | 'MEDIUM' | 'HARD';
  randomFn?: () => number;
}

export class SnakeLadderEngine {
  private config: SnakeLadderConfig;
  public state: SnakeLadderState;
  private random: () => number;

  constructor(config: SnakeLadderConfig) {
    this.config = config;
    this.random = config.randomFn || Math.random;
    const startPosition = config.startPosition ?? 1;
    this.state = {
      state: 'idle',
      playerPosition: startPosition,
      aiPosition: startPosition,
      winner: null,
      lastRoll: 0,
      message: ''
    };
  }

  public start() {
    if (this.state.state !== 'idle') throw new Error("Can only start from idle");
    this.state.state = 'player_turn';
    this.state.message = "¡Juego iniciado. Tu turno!";
  }

  public rollPlayer() {
    if (this.state.state !== 'player_turn') throw new Error("Not player's turn to roll");
    this.state.state = 'player_rolling';
    const roll = Math.floor(this.random() * 6) + 1;
    this.state.lastRoll = roll;
    
    // Engine advances to player_moving, then UI will animate.
    this.state.state = 'player_moving';
    let target = this.state.playerPosition + roll;
    
    // Exceeding last cell: bounce back logic
    const endPosition = this.winPosition();
    if (target > endPosition) {
      const excess = target - endPosition;
      target = endPosition - excess;
    }
    
    this.state.playerPosition = target;
    this.state.state = 'resolving_player_tile';
    this.resolveTile('player', target);
  }

  private resolveTile(who: 'player' | 'ai', position: number) {
    let finalPos = position;
    
    // Check snakes
    const snake = this.config.snakes.find(s => s.start === position);
    if (snake) {
      finalPos = snake.end;
      this.state.message = `${who === 'player' ? '¡El jugador' : '¡La IA'} cayó en una serpiente!`;
    }
    
    // Check ladders
    const ladder = this.config.ladders.find(l => l.start === position);
    if (ladder) {
      finalPos = ladder.end;
      this.state.message = `${who === 'player' ? '¡El jugador' : '¡La IA'} encontró una escalera!`;
    }
    
    if (who === 'player') {
      this.state.playerPosition = finalPos;
      if (finalPos === this.winPosition()) {
        this.state.state = 'game_over';
        this.state.winner = 'player';
        this.state.message = "¡El jugador gana!";
      } else {
        this.state.state = 'ai_turn';
        this.state.message = "AI's turn";
      }
    } else {
      this.state.aiPosition = finalPos;
      if (finalPos === this.winPosition()) {
        this.state.state = 'game_over';
        this.state.winner = 'ai';
        this.state.message = "¡La IA gana!";
      } else {
        this.state.state = 'player_turn';
        this.state.message = "Player's turn";
      }
    }
  }

  public playAITurn() {
    if (this.state.state !== 'ai_turn') throw new Error("Not AI's turn");
    this.state.state = 'ai_thinking';
    
    // Probabilistic logic for AI using weightedRandom
    const endPosition = this.winPosition();
    const remaining = endPosition - this.state.aiPosition;
    
    // By default, fair roll
    let failProbability = 0;
    if (this.config.aiDifficulty === 'HARD') failProbability = 0.1;
    if (this.config.aiDifficulty === 'EASY') failProbability = 0.8;
    if (remaining <= 10) failProbability = 0.8; // Specific rule
    
    // Choose between "Fail" (random 1-6) and "Optimal" (roll exact if possible)
    const options = ["FAIL", "OPTIMAL"];
    const weights = [failProbability, 1 - failProbability];
    const decision = weightedRandom(options, weights, this.random);
    
    let roll = Math.floor(this.random() * 6) + 1;
    if (decision === "OPTIMAL" && remaining <= 6) {
      roll = remaining; // optimal move to win
    }

    this.state.lastRoll = roll;
    this.state.state = 'ai_rolling';
    
    // Engine advances to ai_moving
    this.state.state = 'ai_moving';
    let target = this.state.aiPosition + roll;
    
    if (target > endPosition) {
      const excess = target - endPosition;
      target = endPosition - excess;
    }
    
    this.state.aiPosition = target;
    this.state.state = 'resolving_ai_tile';
    this.resolveTile('ai', target);
  }

  private winPosition(): number {
    return this.config.endPosition ?? this.config.boardSize;
  }
}
