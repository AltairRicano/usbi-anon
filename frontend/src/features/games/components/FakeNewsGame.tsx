import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { FakeNewsItem } from '@usbi/schema';
import { FakeNewsEngine, GameResult } from '@usbi/engine';

interface FakeNewsGameProps {
  news: FakeNewsItem[];
  onFinish: (result: GameResult) => void;
}

export const FakeNewsGame: React.FC<FakeNewsGameProps> = ({ news, onFinish }) => {
  const [engine] = useState(() => new FakeNewsEngine(news));
  const [currentItem, setCurrentItem] = useState<FakeNewsItem | null>(engine.getCurrentItem());

  const handleSwipe = (isFake: boolean) => {
    engine.answer(isFake);
    if (engine.isGameOver()) {
      onFinish(engine.getResult());
      setCurrentItem(null);
    } else {
      setCurrentItem(engine.getCurrentItem());
    }
  };

  if (!currentItem) {
    return (
      <div className="flex flex-col items-center justify-center p-8 text-center text-white bg-slate-800 rounded-xl">
        <h2 className="text-3xl font-bold mb-4">Juego Terminado</h2>
        <p className="text-xl">Puntuación: {engine.getScore()} / {engine.getMaxScore()}</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center justify-center min-h-[400px] w-full max-w-md mx-auto p-4 bg-slate-100 rounded-xl shadow-lg relative overflow-hidden">
      <AnimatePresence mode="wait">
        <motion.div
          key={currentItem.title}
          initial={{ x: 300, opacity: 0 }}
          animate={{ x: 0, opacity: 1 }}
          exit={{ x: -300, opacity: 0 }}
          drag="x"
          dragConstraints={{ left: 0, right: 0 }}
          onDragEnd={(_e, { offset }) => {
            const swipe = offset.x;
            if (swipe < -100) {
              handleSwipe(false); // Swiped left -> VERDADERO
            } else if (swipe > 100) {
              handleSwipe(true); // Swiped right -> FALSO
            }
          }}
          className="bg-[--color-card] p-6 rounded-2xl shadow-xl w-full text-center border-2 border-slate-200 cursor-grab active:cursor-grabbing"
        >
          {currentItem.imageUrl && (
            <img src={currentItem.imageUrl} alt={currentItem.title} className="w-full h-48 object-cover rounded-lg mb-4" />
          )}
          <h3 className="text-2xl font-bold text-slate-800 mb-2">{currentItem.title}</h3>
          <p className="text-slate-600 mb-4">{currentItem.content}</p>
          {/* text-slate-400 on white was 2.56:1 (audit C8); --color-muted is verified AA. */}
          <div className="text-sm text-[--color-muted] italic">Desliza izquierda para VERDADERO, derecha para FALSO</div>
        </motion.div>
      </AnimatePresence>

      <div className="flex justify-between w-full mt-8 px-4 gap-4">
        <button
          onClick={() => handleSwipe(false)}
          className="flex-1 py-3 bg-green-500 text-white rounded-full font-bold shadow-md active:scale-95 transition-transform"
        >
          VERDADERO
        </button>
        <button
          onClick={() => handleSwipe(true)}
          className="flex-1 py-3 bg-red-500 text-white rounded-full font-bold shadow-md active:scale-95 transition-transform"
        >
          FALSO
        </button>
      </div>
    </div>
  );
};
