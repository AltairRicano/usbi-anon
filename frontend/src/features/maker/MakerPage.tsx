import { useMemo, useState, type ComponentType, type ChangeEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '../../shared/components/ui/Button';
import { levelTemplateRegistry } from '../content/maker/registry';
import type { TemplateType } from '../content/types';
import { TEMPLATE_TYPE_LABELS } from '../content/types';

interface MakerMetadata {
  id: string;
  title: string;
  author: string;
  color: string;
  difficulty: number;
  template_type: TemplateType;
}

interface MakerLevelExport {
  metadata: MakerMetadata & { creation_date: string };
  content: unknown;
}

interface MakerFormProps {
  value: unknown;
  onChange: (value: unknown) => void;
  errors?: unknown;
}

interface MakerPreviewProps {
  value: unknown;
}

const generateId = () => {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  return 'id-' + Math.random().toString(36).substr(2, 9);
};

export const MakerPage = () => {
  const navigate = useNavigate();
  const [exportStatus, setExportStatus] = useState<'idle' | 'success' | 'error'>('idle');
  const [metadata, setMetadata] = useState<MakerMetadata>({
    id: generateId(),
    title: '',
    author: '',
    color: '#18529D',
    difficulty: 1,
    template_type: 'trivia' as TemplateType,
  });

  const [content, setContent] = useState<unknown>(() => levelTemplateRegistry['trivia'].getDefaults());
  const [showPreview, setShowPreview] = useState(false);
  const errors = useMemo<unknown>(() => {
    const schema = levelTemplateRegistry[metadata.template_type]?.schema;
    if (schema) {
      try {
        const result = schema.safeParse(content);
        if (!result.success) {
          return result.error.format();
        }
      } catch {
        return { _errors: ["Error interno de validación en la plantilla."] };
      }
    }
    return null;
  }, [content, metadata.template_type]);

  const handleTemplateChange = (e: ChangeEvent<HTMLSelectElement>) => {
    const newType = e.target.value as TemplateType;
    setMetadata({ ...metadata, template_type: newType });
    const registryEntry = levelTemplateRegistry[newType] || levelTemplateRegistry['trivia'];
    setContent(registryEntry.getDefaults());
  };

  const saveLocal = () => {
    if (errors || !metadata.title) {
      setExportStatus('error');
      return;
    }
    
    try {
      const levelExport: MakerLevelExport = { metadata: { ...metadata, creation_date: new Date().toISOString() }, content };
      const stored = localStorage.getItem('usbi_local_levels');
      const parsedLevels = stored ? JSON.parse(stored) : [];
      const levels: MakerLevelExport[] = Array.isArray(parsedLevels) ? parsedLevels : [];
      
      const existingIdx = levels.findIndex((level) => level.metadata.id === metadata.id);
      if (existingIdx >= 0) {
        levels[existingIdx] = levelExport;
      } else {
        levels.push(levelExport);
      }
      localStorage.setItem('usbi_local_levels', JSON.stringify(levels));
      setExportStatus('success');
      setTimeout(() => setExportStatus('idle'), 3000);
    } catch (e) {
      console.error(e);
      setExportStatus('error');
    }
  };

  const onExport = () => {
    setExportStatus('idle');

    if (errors || !metadata.title) {
      setExportStatus('error');
      return;
    }

    const levelExport: MakerLevelExport = { metadata: { ...metadata, creation_date: new Date().toISOString() }, content };
    const jsonStr = JSON.stringify(levelExport, null, 2);
    const fileName = `${metadata.title.replace(/\s+/g, '_') || 'nivel_usbi'}.json`;

    try {
      // Sin empaquetado de escritorio en ningún plan de este proyecto (ver
      // plan/05_Contenido_maker_y_juego.md §5): USBI-Anon es web, así que
      // solo se conserva el camino de descarga vía Blob que usa ../usbi como
      // fallback para navegador puro. Con esto desaparece la dependencia de
      // @tauri-apps/plugin-dialog y @tauri-apps/plugin-fs.
      const blob = new Blob([jsonStr], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = fileName;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      setExportStatus('success');
      setTimeout(() => setExportStatus('idle'), 3000);
    } catch (e) {
      console.error(e);
      setExportStatus('error');
    }
  };

  const registryEntry = levelTemplateRegistry[metadata.template_type] || levelTemplateRegistry['trivia'];
  const FormComponent = registryEntry.FormComponent as ComponentType<MakerFormProps>;
  const PreviewComponent = registryEntry.PreviewComponent as ComponentType<MakerPreviewProps>;
  const hasErrors = errors !== null;

  return (
    <main className="min-h-screen bg-[var(--color-surface)] px-4 py-6 text-[var(--color-foreground)] sm:px-6 lg:px-8">
      <div className="mx-auto max-w-4xl space-y-6">
      <header className="relative flex flex-col items-center justify-center gap-2 text-center sm:px-12">
        <button 
          type="button" 
          onClick={() => navigate(-1)} 
          className="self-start text-sm font-medium text-[var(--color-primary)] hover:underline sm:absolute sm:left-0 sm:top-2"
        >
          ← Volver
        </button>
        <h1 className="text-2xl font-bold text-[var(--color-primary)] sm:text-3xl">Maker Local — Creación de Niveles</h1>
        <p className="max-w-2xl text-sm text-[var(--color-muted)]">
          Crea niveles de prueba. Puedes guardarlos en tu navegador (Crear) o exportarlos como archivo JSON a tu computadora (Exportar).
        </p>
      </header>

      <div className="space-y-6 rounded-lg border border-[var(--color-border)] bg-[var(--color-card)] p-4 text-[var(--color-text-card)] shadow-sm sm:p-6">
        {/* ── Metadata ──────────────────────────────────────────────────── */}
        <fieldset className="space-y-4 rounded-lg border border-[var(--color-border)] p-4">
          <legend className="px-1 text-sm font-semibold text-[var(--color-primary)]">Metadatos del Nivel</legend>

          <div className="space-y-1">
            <label htmlFor="maker-title" className="block text-sm font-medium">Título</label>
            <input 
              id="maker-title" 
              value={metadata.title}
              onChange={(e) => setMetadata({ ...metadata, title: e.target.value })}
              className="w-full rounded border border-[var(--color-border)] bg-[var(--color-background)] p-2 text-[var(--color-foreground)] focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)]" 
              placeholder="Ej: Capitales de América" 
            />
          </div>

          <div className="space-y-1">
            <label htmlFor="maker-author" className="block text-sm font-medium">Autor</label>
            <input 
              id="maker-author" 
              value={metadata.author}
              onChange={(e) => setMetadata({ ...metadata, author: e.target.value })}
              className="w-full rounded border border-[var(--color-border)] bg-[var(--color-background)] p-2 text-[var(--color-foreground)] focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)]" 
              placeholder="Tu nombre o alias" 
            />
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="space-y-1">
              <label htmlFor="maker-difficulty" className="block text-sm font-medium">Dificultad (1–10)</label>
              <input 
                id="maker-difficulty" 
                type="number" 
                min="1" max="10" 
                value={metadata.difficulty}
                onChange={(e) => setMetadata({ ...metadata, difficulty: parseInt(e.target.value, 10) || 1 })}
                className="w-full rounded border border-[var(--color-border)] bg-[var(--color-background)] p-2 text-[var(--color-foreground)] focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)]" 
              />
            </div>
            <div className="space-y-1">
              <label htmlFor="maker-color" className="block text-sm font-medium">Color del nivel</label>
              <input 
                id="maker-color" 
                type="color" 
                value={metadata.color}
                onChange={(e) => setMetadata({ ...metadata, color: e.target.value })}
                className="h-10 w-full cursor-pointer rounded border border-[var(--color-border)] bg-[var(--color-background)] p-1" 
              />
            </div>
          </div>

          <div className="space-y-1">
            <label htmlFor="maker-template" className="block text-sm font-medium">Tipo de Plantilla</label>
            <select 
              id="maker-template" 
              value={metadata.template_type}
              onChange={handleTemplateChange}
              className="w-full rounded border border-[var(--color-border)] bg-[var(--color-background)] p-2 text-[var(--color-foreground)] focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)]"
            >
              {Object.entries(TEMPLATE_TYPE_LABELS).map(([value, label]) => (
                <option key={value} value={value}>{label}</option>
              ))}
            </select>
          </div>
        </fieldset>

        {/* ── Dynamic content ───────────────────────────────────────────── */}
        <fieldset className="space-y-3 rounded-lg border border-[var(--color-border)] p-4">
          <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <legend className="px-1 text-sm font-semibold text-[var(--color-primary)]">Contenido del Nivel</legend>
            <Button 
              type="button"
              onClick={() => setShowPreview(!showPreview)} 
              variant="outline"
              size="sm"
              className="w-full sm:w-auto"
            >
              {showPreview ? 'Ocultar Previsualización' : 'Mostrar Previsualización'}
            </Button>
          </div>

          <div className="min-h-[200px]">
            {showPreview ? (
              <PreviewComponent value={content} />
            ) : (
              <FormComponent value={content} onChange={setContent} errors={errors} />
            )}
          </div>
        </fieldset>

        {/* ── Status & Submit ───────────────────────────────────────────── */}
        {exportStatus === 'success' && (
          <p className="text-sm font-medium text-[var(--color-success)]">✓ Operación realizada exitosamente.</p>
        )}
        {exportStatus === 'error' && (
          <p className="text-sm font-medium text-[var(--color-error)]">✗ Error. Revisa que el título no esté vacío y que no haya errores de validación en el contenido.</p>
        )}
        {hasErrors && (
          <div className="max-h-32 overflow-auto rounded border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-900/30 dark:text-red-200">
            La configuración del nivel contiene errores (revisa los campos).
          </div>
        )}

        <div className="flex flex-col gap-3 pt-2 sm:flex-row sm:gap-4">
          <Button 
            type="button"
            onClick={saveLocal}
            variant="primary"
            className="w-full text-[var(--color-primary-foreground)] sm:flex-1"
          >
            Crear (Guardar en Local)
          </Button>
          
          <Button 
            type="button"
            onClick={onExport}
            variant="secondary"
            className="w-full sm:flex-1"
          >
            Exportar (Descargar JSON)
          </Button>
        </div>
      </div>
      </div>
    </main>
  );
};
