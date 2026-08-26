import type { Snakes } from '@usbi/schema';

export interface SnakeLadderLink {
  start: number;
  end: number;
}

export interface SnakeLadderContent extends Snakes {
  seed?: number;
}

interface GenerateLayoutParams {
  boardWidth: number;
  boardHeight: number;
  snakeCount: number;
  ladderCount: number;
  seed: number;
}

export function clampBoardSize(value: number): number {
  return clampInt(value, 2, 10);
}

export function maxFeatureCount(boardWidth: number, boardHeight: number): number {
  return Math.max(0, Math.floor(((boardWidth * boardHeight) - 2) / 2));
}

export function generateSnakeLadderLinks({
  boardWidth,
  boardHeight,
  snakeCount,
  ladderCount,
  seed,
}: GenerateLayoutParams): { snakes: SnakeLadderLink[]; ladders: SnakeLadderLink[] } {
  const totalCells = boardWidth * boardHeight;
  const maxCount = maxFeatureCount(boardWidth, boardHeight);
  const random = seededRandom(seed);
  const origins = new Set<number>();

  const snakes = Array.from({ length: clampInt(snakeCount, 0, maxCount) })
    .map(() => makeLink(totalCells, origins, random, 'snake'))
    .filter((link): link is SnakeLadderLink => link !== null);

  const ladders = Array.from({ length: clampInt(ladderCount, 0, maxCount) })
    .map(() => makeLink(totalCells, origins, random, 'ladder'))
    .filter((link): link is SnakeLadderLink => link !== null);

  return { snakes, ladders };
}

export function cellAt(row: number, col: number, boardWidth: number, boardHeight: number): number {
  const rowFromBottom = boardHeight - 1 - row;
  const base = rowFromBottom * boardWidth;
  const cellOffset = rowFromBottom % 2 === 0 ? col : boardWidth - 1 - col;
  return base + cellOffset + 1;
}

export function cellCenterPercent(position: number, boardWidth: number, boardHeight: number): { x: number; y: number } {
  const index = Math.max(1, Math.min(position, boardWidth * boardHeight)) - 1;
  const rowFromBottom = Math.floor(index / boardWidth);
  const y = boardHeight - 1 - rowFromBottom;
  const colInRow = index % boardWidth;
  const x = rowFromBottom % 2 === 0 ? colInRow : boardWidth - 1 - colInRow;

  return {
    x: ((x + 0.5) / boardWidth) * 100,
    y: ((y + 0.5) / boardHeight) * 100,
  };
}

function makeLink(
  totalCells: number,
  origins: Set<number>,
  random: () => number,
  type: 'snake' | 'ladder',
): SnakeLadderLink | null {
  const candidates = Array.from({ length: Math.max(0, totalCells - 2) }, (_, idx) => idx + 2)
    .filter((cell) => cell < totalCells && !origins.has(cell));

  for (let attempt = 0; attempt < 80 && candidates.length > 0; attempt++) {
    const start = candidates[Math.floor(random() * candidates.length)];
    const end = type === 'snake'
      ? randomInt(1, start - 1, random)
      : randomInt(start + 1, totalCells, random);

    if (start !== end) {
      origins.add(start);
      return { start, end };
    }
  }

  return null;
}

function randomInt(min: number, max: number, random: () => number): number {
  return Math.floor(random() * (max - min + 1)) + min;
}

function clampInt(value: number, min: number, max: number): number {
  const normalized = Number.isFinite(value) ? Math.trunc(value) : min;
  return Math.max(min, Math.min(max, normalized));
}

function seededRandom(seed: number): () => number {
  let state = Math.abs(Math.trunc(seed)) || 1;
  return () => {
    state = (state * 1664525 + 1013904223) % 4294967296;
    return state / 4294967296;
  };
}
