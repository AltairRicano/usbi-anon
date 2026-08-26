import { MemoryPair } from "@usbi/schema";
import { normalizeMemoryPairs } from "./memory/MemoryPalette.js";

export interface MemoryCard {
  id: string;
  pairId: string;
  pairColor: string;
  content: string;
  isFlipped: boolean;
  isMatched: boolean;
}

export class MemoryEngine {
  private cards: MemoryCard[] = [];
  private flippedIndices: number[] = [];
  private matchedPairs = 0;
  private totalPairs = 0;

  constructor(pairs: MemoryPair[]) {
    const normalizedPairs = normalizeMemoryPairs(pairs);
    this.totalPairs = normalizedPairs.length;
    const deck: MemoryCard[] = [];
    normalizedPairs.forEach((pair, index) => {
      deck.push({ id: `c1_${index}`, pairId: pair.id, pairColor: pair.color, content: pair.content1, isFlipped: false, isMatched: false });
      deck.push({ id: `c2_${index}`, pairId: pair.id, pairColor: pair.color, content: pair.content2, isFlipped: false, isMatched: false });
    });
    
    // Fisher-Yates shuffle
    for (let i = deck.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [deck[i], deck[j]] = [deck[j], deck[i]];
    }
    this.cards = deck;
  }

  getCards(): MemoryCard[] {
    return this.cards;
  }

  getFlippedCount(): number {
    return this.flippedIndices.length;
  }

  flipCard(index: number): boolean {
    if (index < 0 || index >= this.cards.length) return false;
    const card = this.cards[index];
    if (card.isFlipped || card.isMatched) return false;
    
    if (this.flippedIndices.length >= 2) return false;

    card.isFlipped = true;
    this.flippedIndices.push(index);
    return true;
  }

  checkMatch(): { match: boolean, gameOver: boolean } | null {
    if (this.flippedIndices.length !== 2) return null;
    
    const [i1, i2] = this.flippedIndices;
    const c1 = this.cards[i1];
    const c2 = this.cards[i2];
    
    let match = false;
    if (c1.pairId === c2.pairId) {
      match = true;
      c1.isMatched = true;
      c2.isMatched = true;
      this.matchedPairs++;
    } else {
      c1.isFlipped = false;
      c2.isFlipped = false;
    }
    this.flippedIndices = [];
    return {
      match,
      gameOver: this.matchedPairs === this.totalPairs
    };
  }
}
