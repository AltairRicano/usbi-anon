import { describe, it, expect } from 'vitest';
import { SnakeLadderEngine } from './SnakeLadderEngine';

describe('SnakeLadderEngine', () => {
  it('getResult: not started yet is not completed', () => {
    const engine = new SnakeLadderEngine({
      boardSize: 6,
      snakes: [],
      ladders: [],
      aiDifficulty: 'MEDIUM',
    });
    expect(engine.getResult()).toEqual({ completed: false, score: 0, maxScore: 1 });
  });

  it('getResult: completed with full score when the player wins', () => {
    // random() = 0.7 -> floor(0.7*6)+1 = 5 -> from position 1 lands exactly on 6 (win).
    const engine = new SnakeLadderEngine({
      boardSize: 6,
      snakes: [],
      ladders: [],
      aiDifficulty: 'MEDIUM',
      randomFn: () => 0.7,
    });
    engine.start();
    engine.rollPlayer();
    expect(engine.state.winner).toBe('player');
    expect(engine.getResult()).toEqual({ completed: true, score: 1, maxScore: 1 });
  });

  it('getResult: not completed (score 0) when the AI wins', () => {
    // random() = 0.99 makes the player's first roll bounce back from the goal,
    // then forces the AI's weighted decision to OPTIMAL with an exact winning roll.
    const engine = new SnakeLadderEngine({
      boardSize: 6,
      snakes: [],
      ladders: [],
      aiDifficulty: 'MEDIUM',
      randomFn: () => 0.99,
    });
    engine.start();
    engine.rollPlayer();
    expect(engine.state.state).toBe('ai_turn');
    engine.playAITurn();
    expect(engine.state.winner).toBe('ai');
    expect(engine.getResult()).toEqual({ completed: false, score: 0, maxScore: 1 });
  });
});
