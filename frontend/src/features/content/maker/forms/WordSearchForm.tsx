import type { ChangeEvent } from 'react';
import { Button } from '../../../../shared/components/ui/Button';
import { Input } from '../../../../shared/components/ui/Input';
import type { WordSearch } from '@usbi/schema';

export function WordSearchForm({
  value,
  onChange,
}: {
  value: Partial<WordSearch>;
  onChange: (val: Partial<WordSearch>) => void;
}) {
  const words = value.words || [];

  const addWord = () => onChange({ ...value, words: [...words, ''] });
  const removeWord = (idx: number) => onChange({ ...value, words: words.filter((_, i) => i !== idx) });
  const updateWord = (idx: number, val: string) => {
    const next = [...words];
    next[idx] = val.toUpperCase().replace(/[^A-Z]/g, '');
    onChange({ ...value, words: next });
  };

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-3 gap-4">
        <Input label="Ancho (5-24)" type="number" min={5} max={24} value={value.width || 12} onChange={(e: ChangeEvent<HTMLInputElement>) => onChange({ ...value, width: Number(e.target.value) })} />
        <Input label="Alto (5-24)" type="number" min={5} max={24} value={value.height || 12} onChange={(e: ChangeEvent<HTMLInputElement>) => onChange({ ...value, height: Number(e.target.value) })} />
        <div>
           <span className="text-sm font-medium">Semilla (Seed)</span>
           <div className="flex gap-2 mt-1">
             <Input label="Semilla" value={value.seed || 1234} readOnly />
             <Button variant="outline" onClick={() => onChange({ ...value, seed: Math.floor(Math.random() * 10000) })}>Aleatoria</Button>
           </div>
        </div>
      </div>
      <div className="space-y-2">
        <h4 className="font-semibold">Palabras a buscar</h4>
        {words.map((w, idx) => (
          <div key={idx} className="flex gap-2">
            <Input label="" className="flex-1" value={w} onChange={(e: ChangeEvent<HTMLInputElement>) => updateWord(idx, e.target.value)} required />
            {words.length > 2 && <Button variant="outline" className="text-red-500" onClick={() => removeWord(idx)}>X</Button>}
          </div>
        ))}
        <Button variant="outline" onClick={addWord}>+ Agregar Palabra</Button>
      </div>
    </div>
  );
}
