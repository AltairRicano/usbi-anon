import { CrosswordWord } from "@usbi/schema";

export interface CrosswordCell {
  x: number;
  y: number;
  char: string;
}

export interface PlacedWord {
  word: string;
  clue: string;
  x: number;
  y: number;
  isVertical: boolean;
}

export interface CrosswordState {
  score: number;
  isFinished: boolean;
  userGrid: Map<string, string>;
  selectedCell: { x: number; y: number } | null;
  orientation: 'horizontal' | 'vertical';
  /** Cell keys ("x,y") whose letter is correct — locked against further edits. */
  lockedCells: Set<string>;
}

export class CrosswordEngine {
  private grid: Map<string, CrosswordCell> = new Map();
  private placedWords: PlacedWord[] = [];
  private unplacedWords: CrosswordWord[] = [];
  private maxX = 0;
  private maxY = 0;
  
  private state: CrosswordState = {
    score: 0,
    isFinished: false,
    userGrid: new Map(),
    selectedCell: null,
    orientation: 'horizontal',
    lockedCells: new Set()
  };
  private listeners: Set<(state: CrosswordState) => void> = new Set();
  private initialWords: CrosswordWord[];

  constructor(words: CrosswordWord[]) {
    this.initialWords = words;
    this.generateGrid(words);
  }

  private generateGrid(words: CrosswordWord[]) {
    const layout = buildCrosswordLayout(words);
    this.grid = layout.grid;
    this.placedWords = layout.placedWords;
    this.unplacedWords = layout.unplacedWords;
    this.maxX = layout.maxX;
    this.maxY = layout.maxY;
  }

  getPlacedWords(): PlacedWord[] {
    return this.placedWords;
  }

  getUnplacedWords(): CrosswordWord[] {
    return this.unplacedWords;
  }

  isBuildComplete(): boolean {
    return this.unplacedWords.length === 0 && this.placedWords.length === this.initialWords.length;
  }

  getGridCells(): CrosswordCell[] {
    return Array.from(this.grid.values());
  }

  // --- IGameState Extensions ---
  public getState(): CrosswordState {
    return this.state;
  }

  public subscribe(listener: (state: CrosswordState) => void): () => void {
    this.listeners.add(listener);
    listener(this.state);
    return () => this.listeners.delete(listener);
  }

  private notify() {
    this.listeners.forEach((l) => l(this.state));
  }

  public inputChar(char: string) {
    if (this.state.isFinished || !this.state.selectedCell) return;
    const key = `${this.state.selectedCell.x},${this.state.selectedCell.y}`;
    if (!this.grid.has(key)) return;
    if (this.state.lockedCells.has(key)) return;

    const cleanChar = normalizeCrosswordAnswer(char).slice(0, 1);
    if (!cleanChar) {
      this.state.userGrid.delete(key);
      this.validate();
      this.notify();
      return;
    }

    this.state.userGrid.set(key, cleanChar);
    
    // Move to next cell automatically
    const nextX = this.state.selectedCell.x + (this.state.orientation === 'horizontal' ? 1 : 0);
    const nextY = this.state.selectedCell.y + (this.state.orientation === 'vertical' ? 1 : 0);
    const nextKey = `${nextX},${nextY}`;
    
    if (this.grid.has(nextKey)) {
      this.state.selectedCell = { x: nextX, y: nextY };
    }
    
    this.validate();
    this.notify();
  }

  public selectCell(x: number, y: number) {
    const key = `${x},${y}`;
    if (!this.grid.has(key)) return;

    if (this.state.selectedCell && this.state.selectedCell.x === x && this.state.selectedCell.y === y) {
      this.state.orientation = this.state.orientation === 'horizontal' ? 'vertical' : 'horizontal';
    } else {
      this.state.selectedCell = { x, y };
      // Try to guess orientation based on what's available
      const horizHas = this.grid.has(`${x+1},${y}`) || this.grid.has(`${x-1},${y}`);
      const vertHas = this.grid.has(`${x},${y+1}`) || this.grid.has(`${x},${y-1}`);
      if (horizHas && !vertHas) this.state.orientation = 'horizontal';
      if (vertHas && !horizHas) this.state.orientation = 'vertical';
    }
    this.notify();
  }
  
  public navigate(dx: number, dy: number) {
    if (!this.state.selectedCell) return;
    const nx = this.state.selectedCell.x + dx;
    const ny = this.state.selectedCell.y + dy;
    if (this.grid.has(`${nx},${ny}`)) {
      this.state.selectedCell = { x: nx, y: ny };
      this.notify();
    }
  }

  public validate() {
    let allCorrect = true;
    let correctCount = 0;
    const lockedCells = new Set<string>();

    for (const cell of this.grid.values()) {
      const key = `${cell.x},${cell.y}`;
      const userChar = this.state.userGrid.get(key);
      const expectedChar = cell.char.toUpperCase().normalize("NFD").replace(/[\u0300-\u036f]/g, "");

      if (userChar === expectedChar) {
        correctCount++;
        lockedCells.add(key);
      } else {
        allCorrect = false;
      }
    }

    this.state.score = correctCount * 10;
    this.state.lockedCells = lockedCells;

    if (allCorrect && this.grid.size > 0) {
      this.state.isFinished = true;
    }
  }

  public reset() {
    this.state.userGrid.clear();
    this.state.score = 0;
    this.state.isFinished = false;
    this.state.selectedCell = null;
    this.state.orientation = 'horizontal';
    this.state.lockedCells = new Set();
    this.notify();
  }

  public destroy() {
    this.listeners.clear();
  }
}

interface NormalizedCrosswordWord extends CrosswordWord {
  normalizedWord: string;
}

interface CrosswordLayout {
  grid: Map<string, CrosswordCell>;
  placedWords: PlacedWord[];
  unplacedWords: CrosswordWord[];
  maxX: number;
  maxY: number;
}

interface Placement {
  x: number;
  y: number;
  isVertical: boolean;
  score: number;
}

export function normalizeCrosswordAnswer(value: string): string {
  return value
    .trim()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toUpperCase()
    .replace(/[^A-ZÑ]/g, "");
}

export function canBuildConnectedCrossword(words: CrosswordWord[]): boolean {
  const layout = buildCrosswordLayout(words);
  return layout.unplacedWords.length === 0 && layout.placedWords.length === words.length;
}

export function buildCrosswordLayout(words: CrosswordWord[]): CrosswordLayout {
  const normalizedWords = words
    .map((word) => ({ ...word, normalizedWord: normalizeCrosswordAnswer(word.word) }))
    .filter((word) => word.normalizedWord.length >= 2 && word.clue.trim().length > 0);

  if (normalizedWords.length === 0) {
    return emptyLayout(words);
  }

  const sortedWords = [...normalizedWords].sort((a, b) => b.normalizedWord.length - a.normalizedWord.length);
  let bestLayout: CrosswordLayout = emptyLayout(words);

  for (let firstIndex = 0; firstIndex < sortedWords.length; firstIndex++) {
    const orderedWords = [
      sortedWords[firstIndex],
      ...sortedWords.slice(0, firstIndex),
      ...sortedWords.slice(firstIndex + 1),
    ];
    const layout = buildGreedyLayout(orderedWords, words);
    if (layout.placedWords.length > bestLayout.placedWords.length) {
      bestLayout = layout;
    }
    if (layout.unplacedWords.length === 0) {
      return layout;
    }
  }

  return bestLayout;
}

function buildGreedyLayout(words: NormalizedCrosswordWord[], originalWords: CrosswordWord[]): CrosswordLayout {
  const grid: Map<string, CrosswordCell> = new Map();
  const placedWords: PlacedWord[] = [];
  const unplacedWords: CrosswordWord[] = [];

  placeWord(words[0], 0, 0, false, grid, placedWords);

  for (const word of words.slice(1)) {
    const placement = findBestPlacement(word, grid, placedWords);
    if (placement) {
      placeWord(word, placement.x, placement.y, placement.isVertical, grid, placedWords);
    } else {
      unplacedWords.push({ word: word.word, clue: word.clue });
    }
  }

  const normalized = normalizeLayout(grid, placedWords);
  return {
    ...normalized,
    unplacedWords: originalWords.filter((word) => {
      const normalizedWord = normalizeCrosswordAnswer(word.word);
      return !normalized.placedWords.some((placed) => placed.word === normalizedWord);
    }),
  };
}

function findBestPlacement(
  word: NormalizedCrosswordWord,
  grid: Map<string, CrosswordCell>,
  placedWords: PlacedWord[],
): Placement | null {
  let bestPlacement: Placement | null = null;
  const letters = Array.from(word.normalizedWord);

  for (let wordIndex = 0; wordIndex < letters.length; wordIndex++) {
    const letter = letters[wordIndex];
    for (const placed of placedWords) {
      const placedLetters = Array.from(placed.word);
      for (let placedIndex = 0; placedIndex < placedLetters.length; placedIndex++) {
        if (placedLetters[placedIndex] !== letter) continue;

        const intersectionX = placed.isVertical ? placed.x : placed.x + placedIndex;
        const intersectionY = placed.isVertical ? placed.y + placedIndex : placed.y;
        const isVertical = !placed.isVertical;
        const x = isVertical ? intersectionX : intersectionX - wordIndex;
        const y = isVertical ? intersectionY - wordIndex : intersectionY;

        if (canPlaceWord(letters, x, y, isVertical, grid)) {
          const score = Math.abs(x) + Math.abs(y);
          if (!bestPlacement || score < bestPlacement.score) {
            bestPlacement = { x, y, isVertical, score };
          }
        }
      }
    }
  }

  return bestPlacement;
}

function placeWord(
  word: NormalizedCrosswordWord,
  x: number,
  y: number,
  isVertical: boolean,
  grid: Map<string, CrosswordCell>,
  placedWords: PlacedWord[],
) {
  placedWords.push({ word: word.normalizedWord, clue: word.clue, x, y, isVertical });
  Array.from(word.normalizedWord).forEach((char, index) => {
    const cx = isVertical ? x : x + index;
    const cy = isVertical ? y + index : y;
    grid.set(`${cx},${cy}`, { x: cx, y: cy, char });
  });
}

function canPlaceWord(
  letters: string[],
  startX: number,
  startY: number,
  isVertical: boolean,
  grid: Map<string, CrosswordCell>,
): boolean {
  let intersections = 0;

  for (let i = 0; i < letters.length; i++) {
    const x = isVertical ? startX : startX + i;
    const y = isVertical ? startY + i : startY;
    const cell = grid.get(`${x},${y}`);

    if (cell) {
      if (cell.char !== letters[i]) return false;
      intersections++;
      continue;
    }

    const adjacentA = isVertical ? `${x - 1},${y}` : `${x},${y - 1}`;
    const adjacentB = isVertical ? `${x + 1},${y}` : `${x},${y + 1}`;
    if (grid.has(adjacentA) || grid.has(adjacentB)) return false;
  }

  const beforeX = isVertical ? startX : startX - 1;
  const beforeY = isVertical ? startY - 1 : startY;
  const afterX = isVertical ? startX : startX + letters.length;
  const afterY = isVertical ? startY + letters.length : startY;
  if (grid.has(`${beforeX},${beforeY}`) || grid.has(`${afterX},${afterY}`)) return false;

  return intersections > 0;
}

function normalizeLayout(grid: Map<string, CrosswordCell>, placedWords: PlacedWord[]): Omit<CrosswordLayout, 'unplacedWords'> {
  const cells = Array.from(grid.values());
  if (cells.length === 0) {
    return { grid, placedWords, maxX: 0, maxY: 0 };
  }

  const minX = Math.min(...cells.map((cell) => cell.x));
  const minY = Math.min(...cells.map((cell) => cell.y));
  const normalizedGrid: Map<string, CrosswordCell> = new Map();

  cells.forEach((cell) => {
    const x = cell.x - minX;
    const y = cell.y - minY;
    normalizedGrid.set(`${x},${y}`, { ...cell, x, y });
  });

  const normalizedPlacedWords = placedWords.map((word) => ({
    ...word,
    x: word.x - minX,
    y: word.y - minY,
  }));

  return {
    grid: normalizedGrid,
    placedWords: normalizedPlacedWords,
    maxX: Math.max(...Array.from(normalizedGrid.values()).map((cell) => cell.x)),
    maxY: Math.max(...Array.from(normalizedGrid.values()).map((cell) => cell.y)),
  };
}

function emptyLayout(unplacedWords: CrosswordWord[]): CrosswordLayout {
  return {
    grid: new Map(),
    placedWords: [],
    unplacedWords,
    maxX: 0,
    maxY: 0,
  };
}
