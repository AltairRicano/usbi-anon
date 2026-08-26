import type { SnakeLadderContent, SnakeLadderLink } from '../snakesLayout';
import { cellAt, cellCenterPercent } from '../snakesLayout';

export function SnakeLadderPreview({ value }: { value: SnakeLadderContent }) {
  return <SnakeLadderBoardPreview value={value} />;
}

export function SnakeLadderBoardPreview({ value }: { value: SnakeLadderContent }) {
  const boardWidth = Math.max(2, value.board_width || 6);
  const boardHeight = Math.max(2, value.board_height || 6);
  const totalCells = boardWidth * boardHeight;
  const snakes = value.snakes ?? [];
  const ladders = value.ladders ?? [];

  return (
    <div className="space-y-3">
      <div
        className="relative mx-auto w-full max-w-[520px] overflow-hidden rounded-lg border border-[--color-border] bg-[--color-card] shadow-sm"
        style={{ aspectRatio: `${boardWidth} / ${boardHeight}` }}
        aria-label="Vista previa del tablero de serpientes y escaleras"
      >
        <svg className="pointer-events-none absolute inset-0 z-10 h-full w-full" aria-hidden="true">
          {ladders.map((ladder, idx) => (
            <BoardLink key={`ladder-${idx}`} link={ladder} boardWidth={boardWidth} boardHeight={boardHeight} color="#16a34a" />
          ))}
          {snakes.map((snake, idx) => (
            <BoardLink key={`snake-${idx}`} link={snake} boardWidth={boardWidth} boardHeight={boardHeight} color="#dc2626" curved />
          ))}
        </svg>

        <div
          className="grid h-full w-full"
          style={{ gridTemplateColumns: `repeat(${boardWidth}, minmax(0, 1fr))` }}
        >
          {Array.from({ length: totalCells }, (_, visualIdx) => {
            const row = Math.floor(visualIdx / boardWidth);
            const col = visualIdx % boardWidth;
            const cell = cellAt(row, col, boardWidth, boardHeight);
            const isStart = cell === value.start_position;
            const isEnd = cell === value.end_position;
            const hasSnake = snakes.some((snake) => snake.start === cell);
            const hasLadder = ladders.some((ladder) => ladder.start === cell);

            return (
              <div
                key={cell}
                className={[
                  'relative min-h-9 border border-slate-300 p-1 text-xs font-bold text-slate-900 sm:min-h-12',
                  isStart ? 'bg-[var(--color-primary)]/10' : '',
                  isEnd ? 'bg-green-100' : '',
                  !isStart && !isEnd ? 'bg-white' : '',
                ].join(' ')}
              >
                <span>{cell}</span>
                {isStart && <span className="absolute bottom-1 left-1 text-[10px] text-[--color-primary]">Inicio</span>}
                {isEnd && <span className="absolute bottom-1 left-1 text-[10px] text-green-700">Meta</span>}
                {hasSnake && <span className="absolute right-1 top-1 h-2 w-2 rounded-full bg-red-600" />}
                {hasLadder && <span className="absolute right-1 top-1 h-2 w-2 rounded-full bg-green-600" />}
              </div>
            );
          })}
        </div>
      </div>

      <div className="flex flex-wrap justify-center gap-3 text-xs text-[--color-muted]">
        <span>Serpientes: {snakes.length}</span>
        <span>Escaleras: {ladders.length}</span>
        <span>Seed: {value.seed ?? 'sin definir'}</span>
      </div>
    </div>
  );
}

function BoardLink({
  link,
  boardWidth,
  boardHeight,
  color,
  curved = false,
}: {
  link: SnakeLadderLink;
  boardWidth: number;
  boardHeight: number;
  color: string;
  curved?: boolean;
}) {
  const from = cellCenterPercent(link.start, boardWidth, boardHeight);
  const to = cellCenterPercent(link.end, boardWidth, boardHeight);
  const path = curved
    ? `M ${from.x} ${from.y} Q ${(from.x + to.x) / 2 + 8} ${(from.y + to.y) / 2} ${to.x} ${to.y}`
    : `M ${from.x} ${from.y} L ${to.x} ${to.y}`;

  return (
    <>
      <path d={path} fill="none" stroke={color} strokeLinecap="round" strokeWidth="2.8" opacity="0.88" />
      <circle cx={from.x} cy={from.y} r="1.8" fill={color} />
      <circle cx={to.x} cy={to.y} r="1.3" fill={color} opacity="0.75" />
    </>
  );
}
