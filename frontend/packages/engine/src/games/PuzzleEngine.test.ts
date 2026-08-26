import { describe, it, expect } from 'vitest';
import { PuzzleEngine } from './PuzzleEngine';

describe('PuzzleEngine', () => {
  it('should throw on invalid config', () => {
    expect(() => new PuzzleEngine("phrase", 2)).toThrow();
    expect(() => new PuzzleEngine("phrase", 21)).toThrow();
    expect(() => new PuzzleEngine("", 5)).toThrow();
  });

  it('should generate pieces and shuffle them deterministically and distinctly from solution', () => {
    const engine1 = new PuzzleEngine("secret message", 4, 123);
    const engine2 = new PuzzleEngine("secret message", 4, 123);
    
    expect(engine1.getState().pieces).toEqual(engine2.getState().pieces);
    
    // Check distinct from solution
    const state = engine1.getState();
    const isSolved = state.pieces.every((p, i) => p.originalIndex === i);
    expect(isSolved).toBe(false);
  });

  it('should reorder pieces', () => {
    const engine = new PuzzleEngine("secret message", 4, 123);
    const state = engine.getState();
    const firstPiece = state.pieces[0];
    const secondPiece = state.pieces[1];
    
    engine.reorderPieces(0, 1);
    
    const newState = engine.getState();
    expect(newState.pieces[0]).toEqual(secondPiece);
    expect(newState.pieces[1]).toEqual(firstPiece);
    expect(newState.moves).toBe(1);
  });

  it('should finish game when correctly ordered', () => {
    const engine = new PuzzleEngine("abc", 3, 123);
    
    let state = engine.getState();
    while(!state.isFinished) {
      let moved = false;
      for(let i=0; i<state.pieces.length; i++) {
        if(state.pieces[i].originalIndex !== i) {
          const targetIndex = state.pieces.findIndex(p => p.originalIndex === i);
          engine.reorderPieces(targetIndex, i);
          moved = true;
          break;
        }
      }
      if(!moved) break;
      state = engine.getState();
    }
    
    expect(engine.getState().isFinished).toBe(true);
    expect(engine.getState().score).toBeGreaterThan(0);
  });

  it('should reset properly', () => {
    const engine = new PuzzleEngine("abc", 3, 123);
    engine.reorderPieces(0, 1);
    
    engine.reset();
    
    const state = engine.getState();
    expect(state.moves).toBe(0);
    expect(state.score).toBe(0);
    expect(state.isFinished).toBe(false);
  });
});
