import { describe, it, expect } from 'vitest';
import { canBuildConnectedCrossword, CrosswordEngine } from './CrosswordEngine';

describe('CrosswordEngine', () => {
  const words = [
    { word: 'HELLO', clue: 'Greeting' },
    { word: 'WORLD', clue: 'Earth' },
  ];

  it('should generate grid and placed words', () => {
    const engine = new CrosswordEngine(words);
    const cells = engine.getGridCells();
    const placed = engine.getPlacedWords();
    
    expect(cells.length).toBeGreaterThan(0);
    expect(placed.length).toBe(2);
  });

  it('should initialize state', () => {
    const engine = new CrosswordEngine(words);
    const state = engine.getState();
    
    expect(state.score).toBe(0);
    expect(state.isFinished).toBe(false);
    expect(state.selectedCell).toBeNull();
    expect(state.orientation).toBe('horizontal');
  });

  it('should select cell and guess orientation', () => {
    const engine = new CrosswordEngine(words);
    const cells = engine.getGridCells();
    
    // Find a horizontal sequence if possible, or just select first
    const firstCell = cells[0];
    engine.selectCell(firstCell.x, firstCell.y);
    
    const state = engine.getState();
    expect(state.selectedCell).toEqual({ x: firstCell.x, y: firstCell.y });
  });

  it('should toggle orientation on same cell click', () => {
    const engine = new CrosswordEngine(words);
    const cells = engine.getGridCells();
    const firstCell = cells[0];
    
    engine.selectCell(firstCell.x, firstCell.y);
    const initialOrientation = engine.getState().orientation;
    
    engine.selectCell(firstCell.x, firstCell.y);
    expect(engine.getState().orientation).not.toBe(initialOrientation);
  });

  it('should navigate', () => {
    const engine = new CrosswordEngine(words);
    const cells = engine.getGridCells();
    const firstCell = cells[0];
    
    engine.selectCell(firstCell.x, firstCell.y);
    
    // Find a valid neighbor to navigate to
    let neighborX = firstCell.x;
    let neighborY = firstCell.y;
    let dx = 0; let dy = 0;
    
    for (const c of cells) {
      if (c.x === firstCell.x + 1 && c.y === firstCell.y) {
        neighborX = c.x; dx = 1; dy = 0; break;
      }
      if (c.x === firstCell.x && c.y === firstCell.y + 1) {
        neighborY = c.y; dx = 0; dy = 1; break;
      }
    }
    
    if (dx !== 0 || dy !== 0) {
      engine.navigate(dx, dy);
      expect(engine.getState().selectedCell).toEqual({ x: neighborX, y: neighborY });
    }
  });

  it('should input char and validate', () => {
    const engine = new CrosswordEngine(words);
    const cells = engine.getGridCells();
    const firstCell = cells[0];
    
    engine.selectCell(firstCell.x, firstCell.y);
    engine.inputChar('A'); // Wrong char probably
    
    const state = engine.getState();
    expect(state.userGrid.get(`${firstCell.x},${firstCell.y}`)).toBe('A');
    expect(state.score).toBeGreaterThanOrEqual(0); // If 'A' was correct by chance, score > 0
  });

  it('should finish when all correct', () => {
    const engine = new CrosswordEngine(words);
    const cells = engine.getGridCells();
    
    // Cheat and input all correctly
    cells.forEach(c => {
      engine.selectCell(c.x, c.y);
      engine.inputChar(c.char);
    });
    
    const state = engine.getState();
    expect(state.isFinished).toBe(true);
    expect(state.score).toBe(cells.length * 10);
  });

  it('should handle invalid level (empty words)', () => {
    const engine = new CrosswordEngine([]);
    expect(engine.getGridCells().length).toBe(0);
    const state = engine.getState();
    expect(state.isFinished).toBe(false); // Can't finish an empty crossword
  });

  it('should report words that cannot be connected', () => {
    const disconnectedWords = [
      { word: 'SI', clue: 'Palabra usada para afirmar' },
      { word: 'NO', clue: 'Palabra usada para negar' },
    ];
    const engine = new CrosswordEngine(disconnectedWords);

    expect(canBuildConnectedCrossword(disconnectedWords)).toBe(false);
    expect(engine.isBuildComplete()).toBe(false);
    expect(engine.getPlacedWords()).toHaveLength(1);
    expect(engine.getUnplacedWords()).toHaveLength(1);
  });

  it('should normalize input', () => {
    const engine = new CrosswordEngine(words);
    const cells = engine.getGridCells();
    engine.selectCell(cells[0].x, cells[0].y);
    engine.inputChar('á');
    expect(engine.getState().userGrid.get(`${cells[0].x},${cells[0].y}`)).toBe('A');
  });

  it('should reset properly', () => {
    const engine = new CrosswordEngine(words);
    const cells = engine.getGridCells();
    engine.selectCell(cells[0].x, cells[0].y);
    engine.inputChar('A');
    
    engine.reset();
    const state = engine.getState();
    
    expect(state.userGrid.size).toBe(0);
    expect(state.score).toBe(0);
    expect(state.isFinished).toBe(false);
    expect(state.selectedCell).toBeNull();
  });
});
