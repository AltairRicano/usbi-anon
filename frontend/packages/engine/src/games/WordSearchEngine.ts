

export interface WordSearchState {
  grid: string[][];
  words: string[];
  foundWords: string[];
  isFinished: boolean;
  score: number;
  width: number;
  height: number;
}

export class WordSearchEngine {
  private state: WordSearchState;
  private listeners: Set<(state: WordSearchState) => void> = new Set();
  private seed: number;

  constructor(words: string[], width = 10, height = 10, seed = 12345) {
    this.seed = seed;
    
    if (!words || words.length === 0) {
      throw new Error("Invalid configuration: no words provided");
    }

    const normalizedWords = words.map(this.normalizeWord);
    
    const uniqueWords = Array.from(new Set(normalizedWords));

    this.state = {
      grid: Array.from({ length: height }, () => Array(width).fill('')),
      words: uniqueWords,
      foundWords: [],
      isFinished: false,
      score: 0,
      width,
      height
    };
    
    this.generateGrid();
  }

  private normalizeWord(word: string): string {
    return word.normalize("NFD").replace(/[\u0300-\u036f]/g, "").replace(/[^a-zA-Z]/g, "").toUpperCase();
  }

  private random(): number {
    this.seed = (this.seed * 9301 + 49297) % 233280;
    return this.seed / 233280;
  }

  private generateGrid() {
    // Sort words by length descending
    const sortedWords = [...this.state.words].sort((a, b) => b.length - a.length);

    const placedWords: string[] = [];
    for (const word of sortedWords) {
      let placed = false;
      let attempts = 0;
      while (!placed && attempts < 500) {
        attempts++;
        const dir = Math.floor(this.random() * 3); // 0: horiz, 1: vert, 2: diag
        const dx = dir === 0 ? 1 : dir === 1 ? 0 : 1;
        const dy = dir === 0 ? 0 : dir === 1 ? 1 : 1;
        
        const maxX = this.state.width - word.length * dx;
        const maxY = this.state.height - word.length * dy;
        
        if (maxX < 0 || maxY < 0) continue; // No cabe en esta dirección
        
        const startX = Math.floor(this.random() * (maxX + 1));
        const startY = Math.floor(this.random() * (maxY + 1));

        if (this.canPlaceWord(word, startX, startY, dx, dy)) {
          this.placeWord(word, startX, startY, dx, dy);
          placedWords.push(word);
          placed = true;
        }
      }
      if (!placed) {
        console.warn(`Could not place word: ${word}`);
      }
    }
    
    // Solo requerir las palabras que sí cupieron
    this.state.words = placedWords;

    // Fill empty cells
    const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
    for (let y = 0; y < this.state.height; y++) {
      for (let x = 0; x < this.state.width; x++) {
        if (this.state.grid[y][x] === '') {
          this.state.grid[y][x] = letters[Math.floor(this.random() * letters.length)];
        }
      }
    }
  }

  private canPlaceWord(word: string, x: number, y: number, dx: number, dy: number): boolean {
    for (let i = 0; i < word.length; i++) {
      const cx = x + i * dx;
      const cy = y + i * dy;
      if (cx < 0 || cx >= this.state.width || cy < 0 || cy >= this.state.height) return false;
      const cell = this.state.grid[cy][cx];
      if (cell !== '' && cell !== word[i]) {
        return false; // conflict
      }
    }
    return true; // valid intersection or empty
  }

  private placeWord(word: string, x: number, y: number, dx: number, dy: number) {
    for (let i = 0; i < word.length; i++) {
      this.state.grid[y + i * dy][x + i * dx] = word[i];
    }
  }

  public checkWord(selectedLetters: {x: number, y: number}[]) {
    if (this.state.isFinished) return;
    
    const wordStr = selectedLetters.map(pos => this.state.grid[pos.y][pos.x]).join('');
    const wordStrRev = wordStr.split('').reverse().join('');
    
    let updated = false;
    if (this.state.words.includes(wordStr) && !this.state.foundWords.includes(wordStr)) {
      this.state.foundWords = [...this.state.foundWords, wordStr];
      this.state.score += 100;
      updated = true;
    } else if (this.state.words.includes(wordStrRev) && !this.state.foundWords.includes(wordStrRev)) {
      this.state.foundWords = [...this.state.foundWords, wordStrRev];
      this.state.score += 100;
      updated = true;
    }
    
    if (updated) {
      if (this.state.foundWords.length === this.state.words.length) {
        this.state.isFinished = true;
      }
      this.state = { ...this.state };
      this.notify();
    }
  }

  public getState(): WordSearchState {
    return this.state;
  }

  public subscribe(listener: (state: WordSearchState) => void): () => void {
    this.listeners.add(listener);
    listener(this.state);
    return () => this.listeners.delete(listener);
  }

  private notify() {
    this.listeners.forEach((l) => l(this.state));
  }
  
  public reset() {
    this.state.foundWords = [];
    this.state.score = 0;
    this.state.isFinished = false;
    this.notify();
  }
  
  public destroy() {
    this.listeners.clear();
  }
}
