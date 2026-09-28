import type { ChangeEvent } from 'react';
import { Button } from '../../../../shared/components/ui/Button';
import { Input } from '../../../../shared/components/ui/Input';
import { canBuildConnectedCrossword, normalizeCrosswordAnswer } from '@usbi/engine';
import type { CrosswordWord } from '@usbi/schema';

export function CrosswordForm({
  value,
  onChange,
}: {
  value: { words?: CrosswordWord[] };
  onChange: (val: { words: CrosswordWord[] }) => void;
}) {
  const words = value.words || [];
  const completeWords = words.filter((word) => word.word.length >= 2 && word.clue.trim().length > 0);
  const hasEnoughWords = completeWords.length >= 2;
  const normalizedAnswers = completeWords.map((word) => normalizeCrosswordAnswer(word.word));
  const hasDuplicateAnswers = new Set(normalizedAnswers).size !== normalizedAnswers.length;
  const canBuildCrossword = !hasEnoughWords || canBuildConnectedCrossword(completeWords);
  
  const addWord = () => onChange({ ...value, words: [...words, { word: '', clue: '' }] });
  const removeWord = (idx: number) => onChange({ ...value, words: words.filter((_, i) => i !== idx) });
  const updateWord = (idx: number, field: 'word'|'clue', val: string) => {
    const next = [...words];
    next[idx] = { ...next[idx], [field]: val };
    onChange({ ...value, words: next });
  };

  return (
    <div className="space-y-4">
      {words.map((w, idx) => (
        <div key={idx} className="flex gap-2 items-end">
          <div className="flex-1">
            <Input
              label="Palabra"
              value={w.word}
              onChange={(event: ChangeEvent<HTMLInputElement>) => updateWord(idx, 'word', normalizeCrosswordAnswer(event.target.value))}
              required
            />
          </div>
          <div className="flex-[2]">
            <Input
              label="Pista"
              value={w.clue}
              onChange={(event: ChangeEvent<HTMLInputElement>) => updateWord(idx, 'clue', event.target.value)}
              required
            />
          </div>
          {words.length > 2 && <Button type="button" variant="outline" className="text-red-500" onClick={() => removeWord(idx)}>X</Button>}
        </div>
      ))}
      {!canBuildCrossword && (
        <p className="rounded-lg border border-[var(--color-error)] bg-red-50 p-3 text-sm text-[var(--color-error)]">
          Las palabras completas no generan un crucigrama conectado. Agrega palabras con letras en común o cambia alguna respuesta.
        </p>
      )}
      {hasDuplicateAnswers && (
        <p className="rounded-lg border border-[var(--color-error)] bg-red-50 p-3 text-sm text-[var(--color-error)]">
          Hay respuestas repetidas. Cada palabra del crucigrama debe ser distinta.
        </p>
      )}
      <Button type="button" variant="outline" onClick={addWord}>+ Agregar Palabra</Button>
    </div>
  );
}
