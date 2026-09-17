import React, { useState } from 'react';
import { MemoryPair } from '@usbi/schema';
import { MemoryEngine, MemoryCard, GameResult } from '@usbi/engine';
import clsx from 'clsx';
import { DEFAULT_MEMORY_BACK_COLOR, getMemoryBackCardStyle, getMemoryFrontCardStyle, getMemoryReadableTextColor, normalizeMemoryBackColor } from '@usbi/engine';

interface MemoryGameProps {
  pairs: MemoryPair[];
  backColor?: string;
  onFinish: (result: GameResult) => void;
}

export const MemoryGame: React.FC<MemoryGameProps> = ({ pairs, backColor = DEFAULT_MEMORY_BACK_COLOR, onFinish }) => {
  const [engine] = useState(() => new MemoryEngine(pairs));
  const [cards, setCards] = useState<MemoryCard[]>(engine.getCards());
  const [isProcessing, setIsProcessing] = useState(false);
  const cardBackColor = normalizeMemoryBackColor(backColor);

  const handleCardClick = (index: number) => {
    if (isProcessing) return;
    
    const flipped = engine.flipCard(index);
    if (!flipped) return;
    
    // We deep clone or force re-render by doing this so React sees the new object references.
    // Actually, just mapping to a new array of objects ensures React sees the changes before checkMatch mutates them.
    setCards(engine.getCards().map(c => ({...c})));

    if (engine.getFlippedCount() === 2) {
      setIsProcessing(true);
      setTimeout(() => {
        const result = engine.checkMatch();
        setCards([...engine.getCards()]);
        setIsProcessing(false);
        if (result && result.gameOver) {
          onFinish(engine.getResult());
        }
      }, 1000);
    }
  };

  return (
    <div className="flex flex-col items-center justify-center p-4">
      {/* grid-cols-3 en pantallas angostas (~350px), grid-cols-4 desde sm: (audit C10) */}
      <div className="grid grid-cols-3 sm:grid-cols-4 gap-2 sm:gap-4 max-w-2xl w-full">
        {cards.map((card, index) => (
          <div
            key={card.id}
            onClick={() => handleCardClick(index)}
            onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') handleCardClick(index); }}
            role="button"
            tabIndex={0}
            className="relative w-full aspect-[3/4] cursor-pointer"
            style={{ perspective: '1000px' }}
          >
              <div
                className={clsx(
                  'w-full h-full transition-transform duration-500 border-2 rounded-xl shadow-md',
                  (card.isFlipped || card.isMatched) ? 'rotate-y-180' : '',
                  card.isMatched ? 'shadow-lg' : '',
                )}
                style={{
                  transformStyle: 'preserve-3d',
                  transform: (card.isFlipped || card.isMatched) ? 'rotateY(180deg)' : 'rotateY(0deg)',
                }}
              >
              {/* Back of the card (visible when not flipped) */}
              <div
                className="absolute w-full h-full rounded-xl backface-hidden border border-white/25"
                style={{
                  ...getMemoryBackCardStyle(cardBackColor),
                  backfaceVisibility: 'hidden',
                }}
              >
                <div className="flex h-full w-full items-center justify-center">
                  <div className="h-14 w-14 rounded-full border border-white/35 bg-white/10 shadow-inner" />
                </div>
              </div>
              
              {/* Front of the card (visible when flipped) */}
              <div
                className={clsx(
                  'absolute w-full h-full rounded-xl backface-hidden flex items-center justify-center p-2 text-center break-words border',
                  card.isMatched ? 'ring-2 ring-white/60' : 'border-slate-300',
                )}
                style={{
                  ...getMemoryFrontCardStyle(card.pairColor),
                  backfaceVisibility: 'hidden',
                  transform: 'rotateY(180deg)',
                  borderColor: card.pairColor,
                }}
              >
                <span
                  className="font-semibold"
                  style={{ color: getMemoryReadableTextColor(card.pairColor) }}
                >
                  {card.content}
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
