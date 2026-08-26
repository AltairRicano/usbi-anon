import Phaser from 'phaser';
import type { CrosswordEngine, CrosswordState } from '@usbi/engine';

interface SceneData {
  engine?: CrosswordEngine;
  onFinish?: (score: number) => void;
}

export class CrosswordScene extends Phaser.Scene {
  private engine!: CrosswordEngine;
  private onFinish?: (score: number) => void;
  private unsubscribeEngine?: () => void;
  
  private cellSize: number = 50;
  
  private cellRects: Map<string, Phaser.GameObjects.Rectangle> = new Map();
  private cellTexts: Map<string, Phaser.GameObjects.Text> = new Map();
  private highlightGraphics!: Phaser.GameObjects.Graphics;

  // Real, off-screen DOM <input> that mirrors the Phaser canvas's keyboard
  // handling. A canvas cannot receive OS focus, so mobile browsers never open
  // the virtual keyboard on tap — only a real editable HTML element triggers
  // it. Focusing it must happen synchronously inside the pointerdown handler
  // (iOS Safari refuses to open the keyboard from a focus() call outside the
  // original user-gesture call stack).
  private domInput?: HTMLInputElement;
  private handleDomInput = (event: Event) => {
    const target = event.target as HTMLInputElement;
    const value = target.value;
    target.value = '';
    if (!value) return;
    const lastChar = value[value.length - 1];
    if (/[a-zA-ZñÑ]/.test(lastChar)) {
      this.engine.inputChar(lastChar);
    }
  };
  private handleDomKeyDown = (event: KeyboardEvent) => {
    if (event.key === 'Backspace') {
      event.preventDefault();
      this.engine.inputChar('');
      const state = this.engine.getState();
      if (state.orientation === 'horizontal') this.engine.navigate(-1, 0);
      else this.engine.navigate(0, -1);
    } else if (event.key === 'ArrowRight') {
      this.engine.navigate(1, 0);
    } else if (event.key === 'ArrowLeft') {
      this.engine.navigate(-1, 0);
    } else if (event.key === 'ArrowDown') {
      this.engine.navigate(0, 1);
    } else if (event.key === 'ArrowUp') {
      this.engine.navigate(0, -1);
    }
  };

  constructor() {
    super('CrosswordScene');
  }

  init(data: SceneData) {
    this.engine = data.engine ?? this.game.registry.get('crosswordEngine');
    this.onFinish = data.onFinish ?? this.game.registry.get('crosswordOnFinish');
  }

  create() {
    // A nice solid background in case transparent fails
    this.cameras.main.setBackgroundColor('#ffffff');

    if (!this.engine) {
      this.add.text(
        this.cameras.main.width / 2,
        this.cameras.main.height / 2,
        'No se pudo iniciar el crucigrama',
        { fontSize: '20px', color: '#666666', fontFamily: 'sans-serif' }
      ).setOrigin(0.5);
      return;
    }

    this.events.once(Phaser.Scenes.Events.SHUTDOWN, this.shutdown, this);
    this.highlightGraphics = this.add.graphics();
    const cells = this.engine.getGridCells();

    if (cells.length === 0) {
      this.add.text(
        this.cameras.main.width / 2,
        this.cameras.main.height / 2,
        'Sin palabras para mostrar',
        { fontSize: '24px', color: '#ff0000', fontFamily: 'sans-serif' }
      ).setOrigin(0.5);
      return;
    }

    // Find grid bounds
    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
    cells.forEach(c => {
      if (c.x < minX) minX = c.x;
      if (c.x > maxX) maxX = c.x;
      if (c.y < minY) minY = c.y;
      if (c.y > maxY) maxY = c.y;
    });

    const cols = maxX - minX + 1;
    const rows = maxY - minY + 1;
    const padding = 36;
    const availableWidth = this.cameras.main.width - padding * 2;
    const availableHeight = this.cameras.main.height - padding * 2;
    this.cellSize = Math.max(28, Math.min(54, Math.floor(Math.min(availableWidth / cols, availableHeight / rows))));

    const gridW = cols * this.cellSize;
    const gridH = rows * this.cellSize;

    // Use a Container to hold all cells so we can center them easily
    const container = this.add.container(0, 0);

    // Build a set of word-start positions for numbering
    const wordStartLabels = new Map<string, number>();
    const sortedWords = [...this.engine.getPlacedWords()].sort((a, b) => {
      if (a.y === b.y) return a.x - b.x;
      return a.y - b.y;
    });
    
    let currentNum = 1;
    sortedWords.forEach((pw) => {
      const key = `${pw.x},${pw.y}`;
      if (!wordStartLabels.has(key)) {
        wordStartLabels.set(key, currentNum++);
      }
    });

    cells.forEach(c => {
      // Local position inside the grid
      const px = (c.x - minX) * this.cellSize;
      const py = (c.y - minY) * this.cellSize;

      const rect = this.add
        .rectangle(px + this.cellSize / 2, py + this.cellSize / 2, this.cellSize, this.cellSize, 0xffffff)
        .setStrokeStyle(2, 0x000000)
        .setInteractive({ cursor: 'pointer' });

      rect.on('pointerdown', () => {
        this.engine.selectCell(c.x, c.y);
        this.domInput?.focus();
      });

      const text = this.add
        .text(px + this.cellSize / 2, py + this.cellSize / 2, '', {
          fontSize: '26px',
          color: '#000000',
          fontFamily: 'Arial, sans-serif',
          fontStyle: 'bold',
        })
        .setOrigin(0.5);

      const key = `${c.x},${c.y}`;
      this.cellRects.set(key, rect);
      this.cellTexts.set(key, text);

      container.add([rect, text]);

      // Draw small number label
      const labelNum = wordStartLabels.get(key);
      if (labelNum !== undefined) {
        const numText = this.add.text(px + 4, py + 2, String(labelNum), {
          fontSize: '12px',
          color: '#333333',
          fontFamily: 'Arial, sans-serif',
          fontStyle: 'bold'
        });
        container.add(numText);
      }
    });

    // Add highlight graphics to container (behind texts but in front of rects)
    container.addAt(this.highlightGraphics, 0);

    // Center and scale the container
    const scaleX = (this.cameras.main.width - padding * 2) / gridW;
    const scaleY = (this.cameras.main.height - padding * 2) / gridH;
    const scale = Math.min(scaleX, scaleY, 1);
    
    container.setScale(scale);
    
    const scaledW = gridW * scale;
    const scaledH = gridH * scale;
    
    container.x = (this.cameras.main.width - scaledW) / 2;
    container.y = (this.cameras.main.height - scaledH) / 2;

    this.input.keyboard!.on('keydown', this.handleKeyDown, this);
    this.createDomInput();

    this.unsubscribeEngine = this.engine.subscribe((newState) => {
      this.drawState(newState);
      if (newState.isFinished && this.onFinish) {
        this.onFinish(newState.score);
      }
    });
  }

  private drawState(state: CrosswordState) {
    this.highlightGraphics.clear();

    // Reset colors, painting already-correct (locked) cells green
    this.cellRects.forEach((rect, key) => {
      rect.setFillStyle(state.lockedCells.has(key) ? 0xa5d6a7 : 0xffffff);
    });

    // Draw selection (a locked cell keeps its green fill instead)
    if (state.selectedCell) {
      const {x, y} = state.selectedCell;
      const key = `${x},${y}`;
      if (!state.lockedCells.has(key)) {
        const rect = this.cellRects.get(key);
        if (rect) {
          rect.setFillStyle(0xffea00); // Yellow highlight
        }
      }
    }

    // Draw chars
    this.cellTexts.forEach((text, key) => {
      const char = state.userGrid.get(key) || '';
      text.setText(char);
    });
  }

  private createDomInput() {
    const input = document.createElement('input');
    input.type = 'text';
    input.setAttribute('autocomplete', 'off');
    input.setAttribute('autocorrect', 'off');
    input.setAttribute('autocapitalize', 'none');
    input.setAttribute('spellcheck', 'false');
    input.setAttribute('aria-hidden', 'true');
    input.tabIndex = -1;
    // Off-screen but real/focusable — a display:none or 0x0 input is ignored
    // by some mobile browsers' focus handling.
    Object.assign(input.style, {
      position: 'fixed',
      top: '0',
      left: '0',
      width: '1px',
      height: '1px',
      opacity: '0',
      border: 'none',
      padding: '0',
      pointerEvents: 'none',
    });
    input.addEventListener('input', this.handleDomInput);
    input.addEventListener('keydown', this.handleDomKeyDown);
    document.body.appendChild(input);
    this.domInput = input;
  }

  private handleKeyDown(event: KeyboardEvent) {
    if (event.key === 'ArrowRight') {
      this.engine.navigate(1, 0);
    } else if (event.key === 'ArrowLeft') {
      this.engine.navigate(-1, 0);
    } else if (event.key === 'ArrowDown') {
      this.engine.navigate(0, 1);
    } else if (event.key === 'ArrowUp') {
      this.engine.navigate(0, -1);
    } else if (event.key.length === 1 && /[a-zA-ZñÑ\s]/.test(event.key)) {
      this.engine.inputChar(event.key);
    } else if (event.key === 'Backspace') {
      this.engine.inputChar('');
      const state = this.engine.getState();
      if (state.orientation === 'horizontal') this.engine.navigate(-1, 0);
      else this.engine.navigate(0, -1);
    }
  }

  shutdown() {
    this.input.keyboard?.off('keydown', this.handleKeyDown, this);
    if (this.unsubscribeEngine) {
      this.unsubscribeEngine();
      this.unsubscribeEngine = undefined;
    }
    if (this.domInput) {
      this.domInput.removeEventListener('input', this.handleDomInput);
      this.domInput.removeEventListener('keydown', this.handleDomKeyDown);
      this.domInput.remove();
      this.domInput = undefined;
    }
    this.cellRects.forEach(r => r.off('pointerdown'));
    this.cellRects.clear();
    this.cellTexts.clear();
    this.tweens.killAll();
    this.time.removeAllEvents();
  }
}
