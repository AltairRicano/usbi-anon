import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { TriviaEngine } from './TriviaEngine';

describe('TriviaEngine', () => {
  const mockQuestions = [
    { question: 'Q1', options: ['A', 'B', 'C', 'D'], correct_index: 0 },
    { question: 'Q2', options: ['A', 'B', 'C', 'D'], correct_index: 1 },
  ];

  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize with default state', () => {
    const engine = new TriviaEngine(mockQuestions);
    const state = engine.getState();
    expect(state.currentQuestionIndex).toBe(0);
    expect(state.score).toBe(0);
    expect(state.selectedAnswer).toBeNull();
    expect(state.isFinished).toBe(false);
  });

  it('should handle correct answer', () => {
    const engine = new TriviaEngine(mockQuestions);
    engine.submitAnswer(0);
    const state = engine.getState();
    expect(state.selectedAnswer).toBe(0);
    expect(state.score).toBeGreaterThan(0);
  });

  it('should handle incorrect answer', () => {
    const engine = new TriviaEngine(mockQuestions);
    engine.submitAnswer(1);
    const state = engine.getState();
    expect(state.selectedAnswer).toBe(1);
    expect(state.score).toBe(0);
  });

  it('should prevent duplicate answers', () => {
    const engine = new TriviaEngine(mockQuestions);
    engine.submitAnswer(0);
    const firstScore = engine.getState().score;
    engine.submitAnswer(1);
    expect(engine.getState().selectedAnswer).toBe(0);
    expect(engine.getState().score).toBe(firstScore);
  });

  it('should move to next question after timeout', () => {
    const engine = new TriviaEngine(mockQuestions);
    engine.startTimer();
    vi.advanceTimersByTime(31000); // 31s
    expect(engine.getState().selectedAnswer).toBe(-1); // Timeout
    vi.advanceTimersByTime(1600); // Wait for transition
    expect(engine.getState().currentQuestionIndex).toBe(1);
  });

  it('should finish game when questions exhausted', () => {
    const engine = new TriviaEngine(mockQuestions);
    engine.submitAnswer(0); // Q1
    vi.advanceTimersByTime(1600);
    engine.submitAnswer(1); // Q2
    vi.advanceTimersByTime(1600);
    expect(engine.getState().isFinished).toBe(true);
  });

  it('should reset properly', () => {
    const engine = new TriviaEngine(mockQuestions);
    engine.submitAnswer(0);
    vi.advanceTimersByTime(1600);
    engine.reset();
    expect(engine.getState().currentQuestionIndex).toBe(0);
    expect(engine.getState().score).toBe(0);
    expect(engine.getState().isFinished).toBe(false);
  });
});
