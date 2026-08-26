
import { MultipleChoice } from '@usbi/schema';

export interface TriviaState {
  currentQuestionIndex: number;
  score: number;
  selectedAnswer: number | null;
  isFinished: boolean;
  timeLeft: number;
  questions: MultipleChoice[];
}

export class TriviaEngine {
  private state: TriviaState;
  private listeners: Set<(state: TriviaState) => void> = new Set();
  private timer: number | null = null;
  private readonly defaultTimePerQuestion = 30;

  constructor(questions: MultipleChoice[]) {
    this.state = {
      currentQuestionIndex: 0,
      score: 0,
      selectedAnswer: null,
      isFinished: questions.length === 0,
      timeLeft: this.defaultTimePerQuestion,
      questions,
    };
  }

  public getState(): TriviaState {
    return this.state;
  }

  public subscribe(listener: (state: TriviaState) => void): () => void {
    this.listeners.add(listener);
    listener(this.state);
    return () => this.listeners.delete(listener);
  }

  private notify() {
    this.listeners.forEach((l) => l(this.state));
  }

  public startTimer() {
    this.stopTimer();
    this.timer = window.setInterval(() => {
      if (this.state.timeLeft > 0) {
        this.state.timeLeft--;
        this.notify();
      } else {
        // Time is up, move to next
        this.submitAnswer(-1); // -1 for timeout
      }
    }, 1000);
  }

  public stopTimer() {
    if (this.timer !== null) {
      clearInterval(this.timer);
      this.timer = null;
    }
  }

  public submitAnswer(answerIndex: number) {
    if (this.state.isFinished || this.state.selectedAnswer !== null) return;
    this.stopTimer();

    this.state.selectedAnswer = answerIndex;
    const currentQ = this.state.questions[this.state.currentQuestionIndex];
    if (answerIndex === currentQ.correct_index) {
      this.state.score += 100 + this.state.timeLeft * 2; // Simple scoring
    }
    
    this.notify();

    // Transition to next after a delay
    setTimeout(() => {
      this.nextQuestion();
    }, 1500);
  }

  private nextQuestion() {
    if (this.state.currentQuestionIndex < this.state.questions.length - 1) {
      this.state.currentQuestionIndex++;
      this.state.selectedAnswer = null;
      this.state.timeLeft = this.defaultTimePerQuestion;
      this.notify();
      this.startTimer();
    } else {
      this.state.isFinished = true;
      this.notify();
    }
  }

  public reset() {
    this.stopTimer();
    this.state = {
      ...this.state,
      currentQuestionIndex: 0,
      score: 0,
      selectedAnswer: null,
      isFinished: this.state.questions.length === 0,
      timeLeft: this.defaultTimePerQuestion,
    };
    this.notify();
    this.startTimer();
  }

  public destroy() {
    this.stopTimer();
    this.listeners.clear();
  }
}
