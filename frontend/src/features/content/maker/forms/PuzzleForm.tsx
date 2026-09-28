import type { ChangeEvent } from 'react';
import { Input } from '../../../../shared/components/ui/Input';
import { Button } from '../../../../shared/components/ui/Button';
import type { Puzzle } from '@usbi/schema';

export function PuzzleForm({
  value,
  onChange,
}: {
  value: Partial<Puzzle>;
  onChange: (val: Partial<Puzzle>) => void;
}) {
  return (
    <div className="space-y-4">
      <div>
        <label htmlFor="puzzle-phrase" className="text-sm font-medium block mb-1 text-gray-700">Frase o mensaje secreto</label>
        <textarea
          id="puzzle-phrase"
          className="w-full p-2 border border-gray-300 rounded-md focus:ring-[var(--color-primary)] focus:border-[var(--color-primary)]"
          rows={3}
          value={value.phrase || ''}
          onChange={(e: ChangeEvent<HTMLTextAreaElement>) => onChange({ ...value, phrase: e.target.value })}
          required
          placeholder="Escribe la frase secreta aquí..."
        />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <Input label="Número de piezas (3-20)" type="number" min={3} max={20} value={value.pieces || 3} onChange={(e: ChangeEvent<HTMLInputElement>) => onChange({ ...value, pieces: Number(e.target.value) })} required />
        <div>
           <span className="text-sm font-medium">Semilla (Seed)</span>
           <div className="flex gap-2 mt-1">
             <Input label="Semilla" value={value.seed || 1234} readOnly />
             <Button variant="outline" type="button" onClick={() => onChange({ ...value, seed: Math.floor(Math.random() * 10000) })}>Aleatoria</Button>
           </div>
        </div>
      </div>
    </div>
  );
}
