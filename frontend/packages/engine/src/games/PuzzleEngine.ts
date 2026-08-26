export interface PuzzlePiece {
  id: string;
  text: string;
  originalIndex: number;
}

export interface PuzzleState {
  phrase: string;
  piecesCount: number;
  pieces: PuzzlePiece[];
  isFinished: boolean;
  score: number;
  moves: number;
}

export class PuzzleEngine {
  private state: PuzzleState;
  private listeners: Set<(state: PuzzleState) => void> = new Set();
  private seed: number;

  constructor(phrase: string, piecesCount: number, seed = 12345) {
    if (piecesCount < 3 || piecesCount > 20) {
      throw new Error("Invalid configuration: pieces must be between 3 and 20");
    }
    if (!phrase) {
      throw new Error("Invalid configuration: phrase cannot be empty");
    }
    this.seed = seed;
    this.state = {
      phrase,
      piecesCount,
      pieces: [],
      isFinished: false,
      score: 0,
      moves: 0,
    };
    this.generatePieces();
  }

  private random(): number {
    this.seed = (this.seed * 9301 + 49297) % 233280;
    return this.seed / 233280;
  }

  private generatePieces() {
    const { phrase, piecesCount } = this.state;
    const pieces: PuzzlePiece[] = [];
    const baseLength = Math.floor(phrase.length / piecesCount);
    let remainder = phrase.length % piecesCount;
    let startIndex = 0;

    for (let i = 0; i < piecesCount; i++) {
      const length = baseLength + (remainder > 0 ? 1 : 0);
      remainder--;
      pieces.push({
        id: `piece-${i}`,
        text: phrase.substring(startIndex, startIndex + length),
        originalIndex: i,
      });
      startIndex += length;
    }

    // Shuffle
    let isSolved = true;
    let maxAttempts = 100;
    const shuffled = [...pieces];

    while (isSolved && maxAttempts > 0) {
      maxAttempts--;
      for (let i = shuffled.length - 1; i > 0; i--) {
        const j = Math.floor(this.random() * (i + 1));
        [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
      }
      isSolved = shuffled.every((p, i) => p.originalIndex === i);
    }

    this.state.pieces = shuffled;
  }

  public reorderPieces(fromIndex: number, toIndex: number) {
    if (this.state.isFinished) return;
    if (fromIndex < 0 || fromIndex >= this.state.pieces.length || toIndex < 0 || toIndex >= this.state.pieces.length) return;

    const newPieces = [...this.state.pieces];
    const [movedPiece] = newPieces.splice(fromIndex, 1);
    newPieces.splice(toIndex, 0, movedPiece);
    
    this.state.pieces = newPieces;
    this.state.moves++;
    
    this.checkCompletion();
    this.notify();
  }

  private checkCompletion() {
    this.state.isFinished = this.state.pieces.every((p, i) => p.originalIndex === i);
    if (this.state.isFinished) {
      this.state.score = 1000 - (this.state.moves * 10);
      if (this.state.score < 100) this.state.score = 100;
    }
  }

  public getState(): PuzzleState {
    return this.state;
  }

  public subscribe(listener: (state: PuzzleState) => void): () => void {
    this.listeners.add(listener);
    listener(this.state);
    return () => this.listeners.delete(listener);
  }

  private notify() {
    this.listeners.forEach((l) => l(this.state));
  }

  public reset() {
    this.state.moves = 0;
    this.state.score = 0;
    this.state.isFinished = false;
    this.generatePieces();
    this.notify();
  }

  public destroy() {
    this.listeners.clear();
  }
}
