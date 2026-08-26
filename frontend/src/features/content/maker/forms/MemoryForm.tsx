import type { ChangeEvent } from 'react';
import { Button } from '../../../../shared/components/ui/Button';
import { Input } from '../../../../shared/components/ui/Input';
import {
  createMemoryColorOptions,
  createMemoryPairs,
  DEFAULT_MEMORY_BACK_COLOR,
  getMemoryBackCardStyle,
  normalizeMemoryBackColor,
  normalizeMemoryPairs,
  type MemoryPairWithColor,
} from '@usbi/engine';
import type { Memory } from '@usbi/schema';

type MemoryFormValue = Memory;

export function MemoryForm({ value, onChange }: { value: MemoryFormValue; onChange: (val: MemoryFormValue) => void }) {
  const pairs = normalizeMemoryPairs(value.pairs || createMemoryPairs(4));
  const backColor = normalizeMemoryBackColor(value.back_color);
  const palette = createMemoryColorOptions(Math.max(12, pairs.length + 6));

  const updateBackColor = (color: string) => {
    onChange({ ...value, back_color: normalizeMemoryBackColor(color), pairs });
  };
  const addPair = () => {
    const nextColor = palette.find((color) => !pairs.some((pair) => pair.color === color))
      ?? palette[pairs.length % palette.length];
    onChange({ ...value, pairs: [...pairs, { id: crypto.randomUUID(), content1: '', content2: '', color: nextColor }] });
  };
  const removePair = (idx: number) => onChange({ ...value, pairs: pairs.filter((_, i) => i !== idx) });
  const updatePair = (idx: number, field: keyof MemoryPairWithColor, val: string) => {
    const next = [...pairs];
    next[idx] = { ...next[idx], [field]: val } as MemoryPairWithColor;
    onChange({ ...value, pairs: normalizeMemoryPairs(next) });
  };
  const updateColor = (idx: number, color: string) => {
    const next = pairs.map((pair, pairIdx) => (pairIdx === idx ? { ...pair, color } : pair));
    onChange({ ...value, pairs: normalizeMemoryPairs(next) });
  };

  return (
    <div className="space-y-4">
      <div className="grid gap-4 rounded-lg border border-[--color-border] p-4 md:grid-cols-[1fr_180px]">
        <Input
          label="Color del dorso"
          type="color"
          value={backColor || DEFAULT_MEMORY_BACK_COLOR}
          onChange={(event: ChangeEvent<HTMLInputElement>) => updateBackColor(event.target.value)}
          required
        />
        <div
          className="h-24 rounded-lg border border-[--color-border]"
          style={getMemoryBackCardStyle(backColor)}
          aria-label="Vista previa del dorso común"
        />
      </div>
      {pairs.map((p, idx) => (
        <div key={p.id} className="grid gap-4 rounded-lg border border-[--color-border] p-4 md:grid-cols-[1fr_1fr_220px_auto]">
          <div className="flex-1 space-y-2">
            <Input label={`Tarjeta A (Par ${idx + 1})`} value={p.content1} onChange={(event: ChangeEvent<HTMLInputElement>) => updatePair(idx, 'content1', event.target.value)} required />
          </div>
          <div className="flex-1 space-y-2">
            <Input label={`Tarjeta B (Par ${idx + 1})`} value={p.content2} onChange={(event: ChangeEvent<HTMLInputElement>) => updatePair(idx, 'content2', event.target.value)} required />
          </div>
          <label className="flex w-full flex-col gap-1 text-sm font-medium">
            Color del par
            <div className="flex items-center gap-2">
              <span
                className="h-10 w-10 shrink-0 rounded-lg border border-[--color-border]"
                style={{ backgroundColor: p.color }}
                aria-hidden="true"
              />
              <select
                value={p.color}
                onChange={(event) => updateColor(idx, event.currentTarget.value)}
                className="min-h-[44px] w-full rounded-lg border border-[--color-border] bg-[--color-background] px-3"
              >
                {palette
                  .filter((color) => color === p.color || !pairs.some((pair, pairIdx) => pairIdx !== idx && pair.color === color))
                  .map((color) => (
                    <option key={color} value={color}>
                      {color.toUpperCase()}
                    </option>
                  ))}
              </select>
            </div>
            <span className="text-xs text-[--color-muted]">Cada color se reserva para un solo par.</span>
          </label>
          {pairs.length > 4 && (
            <Button type="button" variant="outline" className="mt-6 text-red-500" onClick={() => removePair(idx)}>
              X
            </Button>
          )}
        </div>
      ))}
      <Button type="button" variant="outline" onClick={addPair}>
        + Agregar Par
      </Button>
      <p className="text-xs text-[--color-muted]">Mínimo 4 pares requeridos.</p>
    </div>
  );
}
