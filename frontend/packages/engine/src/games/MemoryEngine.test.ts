import { describe, it, expect } from 'vitest';
import { MemoryEngine } from './MemoryEngine';
import { MemoryPair } from '@usbi/schema';

describe('MemoryEngine', () => {
  const pairs: MemoryPair[] = [
    { id: 'p1', content1: 'A', content2: 'A' },
    { id: 'p2', content1: 'B', content2: 'B' },
  ];

  function matchPair(engine: MemoryEngine, pairId: string) {
    const cards = engine.getCards();
    const indices = cards
      .map((c, i) => ({ c, i }))
      .filter(({ c }) => c.pairId === pairId)
      .map(({ i }) => i);
    engine.flipCard(indices[0]);
    engine.flipCard(indices[1]);
    return engine.checkMatch();
  }

  it('getResult: not completed until every pair is matched', () => {
    const engine = new MemoryEngine(pairs);
    matchPair(engine, 'p1');
    const result = engine.getResult();
    expect(result).toEqual({ completed: false, score: 1, maxScore: 2 });
  });

  it('getResult: completed once every pair is matched (no losing state exists)', () => {
    const engine = new MemoryEngine(pairs);
    matchPair(engine, 'p1');
    matchPair(engine, 'p2');
    const result = engine.getResult();
    expect(result).toEqual({ completed: true, score: 2, maxScore: 2 });
    expect(engine.isGameOver()).toBe(true);
  });

  it('does not match cards from different pairs and flips them back', () => {
    const engine = new MemoryEngine(pairs);
    const cards = engine.getCards();
    const p1Index = cards.findIndex((c) => c.pairId === 'p1');
    const p2Index = cards.findIndex((c) => c.pairId === 'p2');
    engine.flipCard(p1Index);
    engine.flipCard(p2Index);
    const outcome = engine.checkMatch();
    expect(outcome).toEqual({ match: false, gameOver: false });
    expect(engine.getResult().completed).toBe(false);
  });
});
