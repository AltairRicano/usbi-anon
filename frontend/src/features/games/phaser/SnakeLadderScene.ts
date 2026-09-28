import * as Phaser from 'phaser';
import { SnakeLadderEngine, SnakeLadderConfig, mapToGrid } from '@usbi/engine';
import { useSettingsStore } from '../../../features/settings/useSettingsStore';

interface BoardLayout {
  cols: number;
  rows: number;
  cellSize: number;
  offsetX: number;
  offsetY: number;
  boardWidth: number;
  boardHeight: number;
}

export class SnakeLadderScene extends Phaser.Scene {
  private config!: SnakeLadderConfig;
  private engine!: SnakeLadderEngine;
  private playerToken!: Phaser.GameObjects.Arc;
  private aiToken!: Phaser.GameObjects.Arc;
  public isAnimating = false;
  private messageText!: Phaser.GameObjects.Text;
  private diceText!: Phaser.GameObjects.Text;
  private diceGraphics!: Phaser.GameObjects.Graphics;
  private diceRollTimer?: Phaser.Time.TimerEvent;
  private audioCtx: AudioContext | null = null;
  private layout!: BoardLayout;
  private readonly diceBoxSize = 46;

  constructor() {
    super('SnakeLadderScene');
  }

  init() {
    this.config = this.game.registry.get('config');
    this.engine = this.game.registry.get('engine');
  }

  create() {
    this.events.once(Phaser.Scenes.Events.SHUTDOWN, this.shutdown, this);

    this.layout = this.calculateLayout();
    this.drawHud();
    this.drawBoard();
    this.drawLinks();
    this.drawTokens();

    this.events.on('ROLL_DICE', () => {
      if (this.engine.state.state === 'player_turn' && !this.isAnimating) {
        const startPos = this.engine.state.playerPosition;
        this.engine.rollPlayer();
        const endPos = this.engine.state.playerPosition;
        this.playDiceRollAnimation(this.engine.state.lastRoll, () => {
          this.updateState('player', startPos, endPos);
        });
      }
    });

    // Fired when the player answers a gating question incorrectly and the
    // turn is forfeited straight to the AI, bypassing a player roll.
    this.events.on('AI_PLAY', () => {
      if (this.engine.state.state === 'ai_turn' && !this.isAnimating) {
        this.playAITurnAnimated();
      }
    });
  }

  private playAITurnAnimated() {
    const aiStartPos = this.engine.state.aiPosition;
    this.engine.playAITurn();
    const aiEndPos = this.engine.state.aiPosition;
    this.playDiceRollAnimation(this.engine.state.lastRoll, () => {
      this.updateState('ai', aiStartPos, aiEndPos);
    });
  }

  /**
   * ~1s "tumbling die" animation: cycles random faces with a tick sound,
   * then settles on `finalValue` (already resolved by the engine) with a
   * heavier "thud" before invoking `onComplete` to move the token.
   */
  private playDiceRollAnimation(finalValue: number, onComplete: () => void) {
    const tickIntervalMs = 90;
    const totalTicks = Math.round(1000 / tickIntervalMs);
    this.diceGraphics.setVisible(true);

    let tick = 0;
    this.drawDieFace(1 + Math.floor(Math.random() * 6));
    this.playDiceTick(false);

    this.diceRollTimer?.remove();
    this.diceRollTimer = this.time.addEvent({
      delay: tickIntervalMs,
      repeat: totalTicks - 1,
      callback: () => {
        tick++;
        const isLast = tick >= totalTicks;
        this.drawDieFace(isLast ? finalValue : 1 + Math.floor(Math.random() * 6));
        this.playDiceTick(isLast);
        if (isLast) {
          this.time.delayedCall(150, onComplete);
        }
      },
    });
  }

  private drawDieFace(value: number) {
    const size = this.diceBoxSize;
    const x = this.scale.width - size - 20;
    const y = 14;
    const g = this.diceGraphics;

    g.clear();
    g.fillStyle(0xffffff, 1);
    g.lineStyle(2, 0x0f172a, 1);
    g.fillRoundedRect(x, y, size, size, 8);
    g.strokeRoundedRect(x, y, size, size, 8);

    const pad = size * 0.24;
    const mid = size / 2;
    const pipRadius = Math.max(2.5, size * 0.07);
    const facePips: Record<number, [number, number][]> = {
      1: [[mid, mid]],
      2: [[pad, pad], [size - pad, size - pad]],
      3: [[pad, pad], [mid, mid], [size - pad, size - pad]],
      4: [[pad, pad], [size - pad, pad], [pad, size - pad], [size - pad, size - pad]],
      5: [[pad, pad], [size - pad, pad], [mid, mid], [pad, size - pad], [size - pad, size - pad]],
      6: [[pad, pad], [size - pad, pad], [pad, mid], [size - pad, mid], [pad, size - pad], [size - pad, size - pad]],
    };

    g.fillStyle(0x0f172a, 1);
    for (const [px, py] of facePips[value] ?? []) {
      g.fillCircle(x + px, y + py, pipRadius);
    }
  }

  /** Lazily creates (and resumes) a WebAudio context — no audio asset files needed. */
  private ensureAudioContext(): AudioContext | null {
    const AudioCtor = window.AudioContext ?? (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
    if (!AudioCtor) return null;
    if (!this.audioCtx) {
      this.audioCtx = new AudioCtor();
    }
    if (this.audioCtx.state === 'suspended') {
      void this.audioCtx.resume();
    }
    return this.audioCtx;
  }

  /** Short synthesized "tick" for each tumble, or a lower "thud" on the final settle. */
  private playDiceTick(isSettle: boolean) {
    // Audit finding C7: this was the only way to play the dice sound, with no
    // way to disable it — relevant in shared computer labs / with headphones.
    if (useSettingsStore.getState().muteGameSounds) return;
    const ctx = this.ensureAudioContext();
    if (!ctx) return;

    const now = ctx.currentTime;
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.type = isSettle ? 'triangle' : 'square';
    osc.frequency.setValueAtTime(isSettle ? 180 : 320 + Math.random() * 120, now);
    gain.gain.setValueAtTime(isSettle ? 0.12 : 0.06, now);
    gain.gain.exponentialRampToValueAtTime(0.001, now + (isSettle ? 0.18 : 0.06));
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.start(now);
    osc.stop(now + (isSettle ? 0.2 : 0.07));
  }

  updateState(who: 'player' | 'ai', startPos: number, endPos: number) {
    this.messageText.setText(this.engine.state.message);
    this.diceText.setText(`Tiro: ${this.engine.state.lastRoll}`);

    const token = who === 'player' ? this.playerToken : this.aiToken;

    this.animateToken(token, startPos, endPos, () => {
      if (who === 'player' && this.engine.state.state === 'ai_turn') {
        this.time.delayedCall(500, () => {
          if (!this.isAnimating) {
            this.playAITurnAnimated();
          }
        });
      } else if (this.engine.state.state === 'game_over') {
        this.events.emit('GAME_OVER', { winner: this.engine.state.winner });
      }
    });
  }

  animateToken(token: Phaser.GameObjects.Arc, startPos: number, endPos: number, onComplete: () => void) {
    this.isAnimating = true;

    const roll = this.engine.state.lastRoll;
    const winPosition = this.config.endPosition ?? this.config.boardSize;
    let intermediatePos = startPos + roll;
    if (intermediatePos > winPosition) {
      const excess = intermediatePos - winPosition;
      intermediatePos = winPosition - excess;
    }

    const isPlayer = token === this.playerToken;
    const positions: { x: number; y: number }[] = [];

    if (intermediatePos > startPos) {
      for (let i = startPos + 1; i <= intermediatePos; i++) {
        positions.push(this.tokenPoint(i, isPlayer));
      }
    } else if (intermediatePos < startPos) {
      for (let i = startPos + 1; i <= winPosition; i++) {
        positions.push(this.tokenPoint(i, isPlayer));
      }
      for (let i = winPosition - 1; i >= intermediatePos; i--) {
        positions.push(this.tokenPoint(i, isPlayer));
      }
    } else {
      positions.push(this.tokenPoint(endPos, isPlayer));
    }

    if (intermediatePos !== endPos) {
      positions.push(this.tokenPoint(endPos, isPlayer));
    }

    if (positions.length === 0) {
      this.isAnimating = false;
      onComplete();
      return;
    }

    this.tweens.chain({
      tweens: positions.map((pos, idx) => {
        const isJump = idx === positions.length - 1 && intermediatePos !== endPos;
        return {
          targets: token,
          x: pos.x,
          y: pos.y,
          duration: isJump ? 520 : 180,
          ease: isJump ? 'Sine.easeInOut' : 'Linear',
        };
      }),
      onComplete: () => {
        this.isAnimating = false;
        onComplete();
      },
    });
  }

  private calculateLayout(): BoardLayout {
    const cols = this.config.boardWidth ?? Math.ceil(Math.sqrt(this.config.boardSize));
    const rows = this.config.boardHeight ?? Math.ceil(this.config.boardSize / cols);
    const width = this.scale.width;
    const height = this.scale.height;
    const hudHeight = 82;
    const padding = 24;
    const availableWidth = width - padding * 2;
    const availableHeight = height - hudHeight - padding * 2;
    const cellSize = Math.floor(Math.min(availableWidth / cols, availableHeight / rows));
    const boardWidth = cellSize * cols;
    const boardHeight = cellSize * rows;

    return {
      cols,
      rows,
      cellSize,
      boardWidth,
      boardHeight,
      offsetX: Math.round((width - boardWidth) / 2),
      offsetY: Math.round(hudHeight + (height - hudHeight - boardHeight) / 2),
    };
  }

  private drawHud() {
    this.add.rectangle(0, 0, this.scale.width, 74, 0xf8fafc).setOrigin(0);
    this.messageText = this.add.text(20, 16, this.engine.state.message, {
      color: '#0f172a',
      fontSize: '22px',
      fontStyle: 'bold',
    });
    this.diceText = this.add.text(20, 44, '', {
      color: '#334155',
      fontSize: '18px',
    });
    this.diceGraphics = this.add.graphics().setVisible(false);
  }

  private drawBoard() {
    const { cols, rows, cellSize, offsetX, offsetY } = this.layout;
    const totalCells = cols * rows;
    const startPosition = this.config.startPosition ?? 1;
    const endPosition = this.config.endPosition ?? totalCells;

    for (let i = 1; i <= totalCells; i++) {
      const { x, y } = mapToGrid(i, cols, rows);
      const cx = offsetX + x * cellSize;
      const cy = offsetY + y * cellSize;
      const isStart = i === startPosition;
      const isEnd = i === endPosition;
      const fill = isStart ? 0xdbeafe : isEnd ? 0xdcfce7 : 0xffffff;

      this.add.rectangle(cx, cy, cellSize, cellSize, fill)
        .setStrokeStyle(1, 0x334155)
        .setOrigin(0);

      this.add.text(cx + cellSize * 0.08, cy + cellSize * 0.08, `${i}`, {
        color: '#0f172a',
        fontSize: `${Math.max(12, Math.floor(cellSize * 0.22))}px`,
        fontStyle: 'bold',
      });

      if (isStart || isEnd) {
        this.add.text(cx + cellSize / 2, cy + cellSize * 0.65, isStart ? 'INICIO' : 'META', {
          color: isStart ? '#1d4ed8' : '#15803d',
          fontSize: `${Math.max(10, Math.floor(cellSize * 0.16))}px`,
          fontStyle: 'bold',
        }).setOrigin(0.5);
      }
    }
  }

  private drawLinks() {
    const graphics = this.add.graphics();
    this.config.ladders.forEach((ladder) => this.drawLadder(graphics, ladder.start, ladder.end));
    this.config.snakes.forEach((snake) => this.drawSnake(graphics, snake.start, snake.end));
  }

  private drawSnake(graphics: Phaser.GameObjects.Graphics, start: number, end: number) {
    const from = this.cellCenter(start);
    const to = this.cellCenter(end);
    const curve = new Phaser.Curves.QuadraticBezier(
      new Phaser.Math.Vector2(from.x, from.y),
      new Phaser.Math.Vector2((from.x + to.x) / 2 + this.layout.cellSize * 0.35, (from.y + to.y) / 2),
      new Phaser.Math.Vector2(to.x, to.y),
    );

    graphics.lineStyle(Math.max(4, this.layout.cellSize * 0.08), 0xdc2626, 0.78);
    curve.draw(graphics, 32);
    graphics.fillStyle(0xdc2626, 0.95);
    graphics.fillCircle(from.x, from.y, Math.max(5, this.layout.cellSize * 0.1));
  }

  private drawLadder(graphics: Phaser.GameObjects.Graphics, start: number, end: number) {
    const from = this.cellCenter(start);
    const to = this.cellCenter(end);
    const dx = to.x - from.x;
    const dy = to.y - from.y;
    const length = Math.hypot(dx, dy) || 1;
    const normalX = (-dy / length) * Math.max(5, this.layout.cellSize * 0.08);
    const normalY = (dx / length) * Math.max(5, this.layout.cellSize * 0.08);

    graphics.lineStyle(Math.max(3, this.layout.cellSize * 0.05), 0x16a34a, 0.9);
    graphics.lineBetween(from.x + normalX, from.y + normalY, to.x + normalX, to.y + normalY);
    graphics.lineBetween(from.x - normalX, from.y - normalY, to.x - normalX, to.y - normalY);

    for (let i = 1; i <= 4; i++) {
      const t = i / 5;
      const rungX = from.x + dx * t;
      const rungY = from.y + dy * t;
      graphics.lineBetween(rungX + normalX, rungY + normalY, rungX - normalX, rungY - normalY);
    }
  }

  private drawTokens() {
    const playerPoint = this.tokenPoint(this.engine.state.playerPosition, true);
    const aiPoint = this.tokenPoint(this.engine.state.aiPosition, false);
    const radius = Math.max(7, Math.min(13, this.layout.cellSize * 0.16));
    this.playerToken = this.add.circle(playerPoint.x, playerPoint.y, radius, 0x2563eb)
      .setStrokeStyle(2, 0xffffff);
    this.aiToken = this.add.circle(aiPoint.x, aiPoint.y, radius, 0xc026d3)
      .setStrokeStyle(2, 0xffffff);
  }

  private cellCenter(position: number): { x: number; y: number } {
    const safePosition = Phaser.Math.Clamp(position, 1, this.layout.cols * this.layout.rows);
    const { x, y } = mapToGrid(safePosition, this.layout.cols, this.layout.rows);
    return {
      x: this.layout.offsetX + x * this.layout.cellSize + this.layout.cellSize / 2,
      y: this.layout.offsetY + y * this.layout.cellSize + this.layout.cellSize / 2,
    };
  }

  private tokenPoint(position: number, isPlayer: boolean): { x: number; y: number } {
    const center = this.cellCenter(position);
    const spread = Math.max(7, this.layout.cellSize * 0.16);
    return {
      x: center.x + (isPlayer ? -spread : spread),
      y: center.y + spread * 0.35,
    };
  }

  private shutdown() {
    this.diceRollTimer?.remove();
    this.diceRollTimer = undefined;
    if (this.audioCtx) {
      void this.audioCtx.close();
      this.audioCtx = null;
    }
  }
}
