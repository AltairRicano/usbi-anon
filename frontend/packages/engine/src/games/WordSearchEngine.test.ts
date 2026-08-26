import { describe, it, expect } from 'vitest';
import { WordSearchEngine } from './WordSearchEngine';

describe('WordSearchEngine', () => {
  it('should throw if no words provided', () => {
    expect(() => new WordSearchEngine([])).toThrow();
  });

  it('should normalize words and remove duplicates', () => {
    const engine = new WordSearchEngine(['Café', 'café', 'Niño', 'niño', 'hello-world']);
    const state = engine.getState();
    expect(state.words).toContain('CAFE');
    expect(state.words).toContain('NINO');
    expect(state.words).toContain('HELLOWORLD');
    expect(state.words.length).toBe(3); // CAFE, NINO, HELLOWORLD
  });

  it('should place words and generate grid deterministically', () => {
    const engine1 = new WordSearchEngine(['TEST'], 10, 10, 123);
    const engine2 = new WordSearchEngine(['TEST'], 10, 10, 123);
    expect(engine1.getState().grid).toEqual(engine2.getState().grid);
  });

  it('should check selected words correctly (forward and backward)', () => {
    const engine = new WordSearchEngine(['CAT'], 5, 5, 123);
    // Find where CAT is
    const state = engine.getState();
    let catCoords: {x: number, y: number}[] = [];
    outer: for(let y = 0; y < state.height; y++) {
      for(let x = 0; x < state.width; x++) {
        if(state.grid[y][x] === 'C') {
          // just try right
          if(x+2 < state.width && state.grid[y][x+1] === 'A' && state.grid[y][x+2] === 'T') {
             catCoords = [{x, y}, {x:x+1, y}, {x:x+2, y}];
             break outer;
          }
          // try down
          if(y+2 < state.height && state.grid[y+1][x] === 'A' && state.grid[y+2][x] === 'T') {
             catCoords = [{x, y}, {x, y:y+1}, {x, y:y+2}];
             break outer;
          }
          // try diag
          if(y+2 < state.height && x+2 < state.width && state.grid[y+1][x+1] === 'A' && state.grid[y+2][x+2] === 'T') {
             catCoords = [{x, y}, {x:x+1, y:y+1}, {x:x+2, y:y+2}];
             break outer;
          }
        }
      }
    }
    
    expect(catCoords.length).toBe(3);

    // Test selection backward
    const reverseCoords = [...catCoords].reverse();
    engine.checkWord(reverseCoords);
    expect(engine.getState().foundWords).toContain('CAT');
    expect(engine.getState().score).toBe(100);
    expect(engine.getState().isFinished).toBe(true);
  });

  it('should ignore invalid selection', () => {
    const engine = new WordSearchEngine(['CAT'], 5, 5, 123);
    engine.checkWord([{x:0,y:0}, {x:1,y:1}]); // Invalid
    expect(engine.getState().foundWords).toHaveLength(0);
  });

  it('should ignore duplicate correct selection', () => {
    const engine = new WordSearchEngine(['CAT'], 5, 5, 123);
    const state = engine.getState();
    let catCoords: {x: number, y: number}[] = [];
    outer: for(let y = 0; y < state.height; y++) {
      for(let x = 0; x < state.width; x++) {
        if(state.grid[y][x] === 'C' && x+2 < state.width && state.grid[y][x+1] === 'A' && state.grid[y][x+2] === 'T') {
            catCoords = [{x, y}, {x:x+1, y}, {x:x+2, y}];
            break outer;
        } else if (state.grid[y][x] === 'C' && y+2 < state.height && state.grid[y+1][x] === 'A' && state.grid[y+2][x] === 'T') {
            catCoords = [{x, y}, {x, y:y+1}, {x, y:y+2}];
            break outer;
        } else if (state.grid[y][x] === 'C' && y+2 < state.height && x+2 < state.width && state.grid[y+1][x+1] === 'A' && state.grid[y+2][x+2] === 'T') {
            catCoords = [{x, y}, {x:x+1, y:y+1}, {x:x+2, y:y+2}];
            break outer;
        }
      }
    }

    engine.checkWord(catCoords);
    engine.checkWord(catCoords); // duplicate
    expect(engine.getState().foundWords).toHaveLength(1);
    expect(engine.getState().score).toBe(100); // Only scored once
  });

  it('should reset properly', () => {
    const engine = new WordSearchEngine(['DOG'], 5, 5, 123);
    engine.getState().foundWords.push('DOG');
    engine.getState().score = 100;
    engine.getState().isFinished = true;
    
    engine.reset();
    expect(engine.getState().foundWords).toHaveLength(0);
    expect(engine.getState().score).toBe(0);
    expect(engine.getState().isFinished).toBe(false);
  });
});
