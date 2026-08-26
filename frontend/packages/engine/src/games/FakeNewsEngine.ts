import { FakeNewsItem } from "@usbi/schema";

export class FakeNewsEngine {
  private news: FakeNewsItem[];
  private currentIndex = 0;
  private score = 0;

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
}
