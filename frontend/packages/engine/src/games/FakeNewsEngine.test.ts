import { describe, it, expect } from 'vitest';
import { FakeNewsEngine } from './FakeNewsEngine';
import { FakeNewsItem } from '@usbi/schema';

describe('FakeNewsEngine', () => {
  const news: FakeNewsItem[] = [
    { title: 'N1', content: 'C1', isFake: true, reference: 'R1' },
    { title: 'N2', content: 'C2', isFake: false, reference: 'R2' },
    { title: 'N3', content: 'C3', isFake: false, reference: 'R3' },
  ];

  it('should serve items in order and detect end of game', () => {
    const engine = new FakeNewsEngine(news);
    expect(engine.getCurrentItem()).toEqual(news[0]);
    expect(engine.isGameOver()).toBe(false);
    engine.answer(true);
    engine.answer(false);
    engine.answer(false);
    expect(engine.getCurrentItem()).toBeNull();
    expect(engine.isGameOver()).toBe(true);
  });

  it('should score only correct guesses', () => {
    const engine = new FakeNewsEngine(news);
    expect(engine.answer(true)).toBe(true); // N1 is fake
    expect(engine.answer(true)).toBe(false); // N2 is not fake
    expect(engine.answer(false)).toBe(true); // N3 is not fake
    expect(engine.getScore()).toBe(2);
    expect(engine.getMaxScore()).toBe(3);
  });

  it('getResult: not completed when below pass ratio', () => {
    const engine = new FakeNewsEngine(news);
    engine.answer(true); // correct
    engine.answer(true); // incorrect
    engine.answer(true); // incorrect
    expect(engine.getResult()).toEqual({ completed: false, score: 1, maxScore: 3 });
  });

  it('getResult: completed when hits pass ratio (>=60% correct)', () => {
    const engine = new FakeNewsEngine(news);
    engine.answer(true); // correct
    engine.answer(false); // correct
    engine.answer(false); // correct
    expect(engine.getResult()).toEqual({ completed: true, score: 3, maxScore: 3 });
  });
});
