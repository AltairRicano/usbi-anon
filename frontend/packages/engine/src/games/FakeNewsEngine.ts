import { FakeNewsItem } from "@usbi/schema";
import { GameResult } from "../interfaces/GameResult.js";

const PASS_RATIO = 0.6;

export class FakeNewsEngine {
  private news: FakeNewsItem[];
  private currentIndex = 0;
  private score = 0;
  /** Lo que el jugador marcó como "es falsa" por ítem, en orden — evidencia que el servidor verifica en Fase B. */
  private guesses: boolean[] = [];

  constructor(news: FakeNewsItem[]) {
    this.news = news;
  }

  getCurrentItem(): FakeNewsItem | null {
    if (this.currentIndex >= this.news.length) return null;
    return this.news[this.currentIndex];
  }

  answer(isFake: boolean): boolean {
    const item = this.getCurrentItem();
    if (!item) return false;
    
    const correct = item.isFake === isFake;
    if (correct) {
      this.score++;
    }
    this.guesses.push(isFake);
    this.currentIndex++;
    return correct;
  }

  isGameOver(): boolean {
    return this.currentIndex >= this.news.length;
  }

  getScore(): number {
    return this.score;
  }
  
  getMaxScore(): number {
    return this.news.length;
  }

  getResult(): GameResult {
    const maxScore = this.getMaxScore();
    const completed = maxScore > 0 && this.score / maxScore >= PASS_RATIO;
    return { completed, score: this.score, maxScore, answers: { guesses: this.guesses } };
  }
}
