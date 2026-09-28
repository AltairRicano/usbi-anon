import type { ChangeEvent } from 'react';
import { Input } from '../../../shared/components/ui/Input';
import type { SectionDTO, TemplateType } from '../types';
import { TEMPLATE_TYPE_LABELS } from '../types';

interface LevelMetadataFormProps {
  title: string;
  setTitle: (t: string) => void;
  color: string;
  setColor: (c: string) => void;
  difficulty: number;
  setDifficulty: (d: number) => void;
  templateType: string;
  setTemplateType: (t: TemplateType) => void;
  sectionId: string;
  setSectionId: (id: string) => void;
  sections: SectionDTO[];
  isEditing: boolean;
}

export function LevelMetadataForm({
  title, setTitle, color, setColor, difficulty, setDifficulty, 
  templateType, setTemplateType, sectionId, setSectionId, sections, isEditing
}: LevelMetadataFormProps) {
  return (
    <div className="space-y-4 mb-6">
      <div className="grid grid-cols-2 gap-4">
        <label className="flex flex-col gap-1 text-sm font-medium">
          Sección
          <select
            className="min-h-[44px] rounded-lg border border-[var(--color-border)] bg-[var(--color-card)] px-3"
            value={sectionId}
            onChange={(e) => {
              const newId = e.currentTarget.value;
              setSectionId(newId);
              const sec = sections.find(s => s.id === newId);
              if (sec) setColor(sec.color); // Regla de color heredada
            }}
            required
            disabled={isEditing}
          >
            <option value="">Selecciona una sección</option>
            {sections.map(s => <option key={s.id} value={s.id}>{s.title}</option>)}
          </select>
        </label>
        <Input label="Título del nivel" value={title} onChange={(e: ChangeEvent<HTMLInputElement>) => setTitle(e.currentTarget.value)} required />
      </div>
      <div className="grid grid-cols-3 gap-4">
        <Input label="Dificultad (1-10)" type="number" min={1} max={10} value={difficulty} onChange={(e: ChangeEvent<HTMLInputElement>) => setDifficulty(Number(e.currentTarget.value))} required />
        <Input label="Color" type="color" value={color} onChange={(e: ChangeEvent<HTMLInputElement>) => setColor(e.currentTarget.value)} required />
        <label className="flex flex-col gap-1 text-sm font-medium">
          Plantilla
          <select
            className="min-h-[44px] rounded-lg border border-[var(--color-border)] bg-[var(--color-card)] px-3"
            value={templateType}
            onChange={(e: ChangeEvent<HTMLSelectElement>) => setTemplateType(e.currentTarget.value as TemplateType)}
            required
            disabled={isEditing}
          >
            {Object.entries(TEMPLATE_TYPE_LABELS).map(([value, label]) => (
              <option key={value} value={value}>{label}</option>
            ))}
          </select>
        </label>
      </div>
    </div>
  );
}
