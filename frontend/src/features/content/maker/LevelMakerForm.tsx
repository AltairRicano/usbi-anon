import { useState, useEffect, Component, type ComponentType, type ReactNode } from 'react';
import { levelTemplateRegistry } from './registry';
import { LevelMetadataForm } from './LevelMetadataForm';
import { LevelActions } from './LevelActions';
import type { SectionDTO, TemplateType } from '../types';

export interface LevelMakerFormInitialData {
  id?: string;
  section_id?: string;
  title?: string;
  color?: string;
  difficulty?: number;
  template_type?: string;
  content?: unknown;
}

export interface LevelMakerFormSaveData {
  section_id: string;
  title: string;
  color: string;
  difficulty: number;
  template_type: TemplateType;
  content: unknown;
}

interface LevelMakerFormProps {
  initialData?: LevelMakerFormInitialData | null;
  sections: SectionDTO[];
  onSave: (data: LevelMakerFormSaveData) => Promise<void>;
  onCancel: () => void;
}

const VALID_TEMPLATE_TYPES: TemplateType[] = ['trivia', 'crossword', 'word_search', 'puzzle', 'fake_news', 'memory', 'snakes_ladders'];

// Protege contra que una plantilla individual reviente toda la pantalla de
// administración de contenido — un formulario roto queda contenido a su
// propio recuadro en vez de tumbar el resto del panel.
class FormErrorBoundary extends Component<{ children: ReactNode }, { error: string | null }> {
  constructor(props: { children: ReactNode }) {
    super(props);
    this.state = { error: null };
  }
  static getDerivedStateFromError(err: Error) {
    return { error: err.message };
  }
  render() {
    if (this.state.error) {
      return (
        <div className="rounded border border-red-300 bg-red-50 p-4 text-red-700">
          <p className="font-semibold">Error al renderizar el formulario</p>
          <pre className="mt-2 text-xs whitespace-pre-wrap">{this.state.error}</pre>
        </div>
      );
    }
    return this.props.children;
  }
}

function resolveTemplateType(data?: LevelMakerFormInitialData | null): TemplateType {
  const raw = data?.template_type ?? 'trivia';
  return VALID_TEMPLATE_TYPES.includes(raw as TemplateType) ? (raw as TemplateType) : 'trivia';
}

function resolveContent(data: LevelMakerFormInitialData | null | undefined, registryEntry: { getDefaults: () => unknown }): unknown {
  if (!data) return registryEntry.getDefaults();
  const raw = data.content;
  if (raw === null || raw === undefined) return registryEntry.getDefaults();
  if (typeof raw === 'string') {
    try {
      return JSON.parse(raw);
    } catch {
      return registryEntry.getDefaults();
    }
  }
  return raw;
}

function LevelMakerFormInner({ initialData, sections, onSave, onCancel }: LevelMakerFormProps) {
  const isEditing = !!initialData;
  const [title, setTitle] = useState(initialData?.title ?? '');
  const [color, setColor] = useState(initialData?.color ?? '#28AD56');
  const [difficulty, setDifficulty] = useState(initialData?.difficulty ?? 1);
  const [sectionId, setSectionId] = useState(initialData?.section_id ?? (sections[0]?.id ?? ''));
  const [templateType, setTemplateType] = useState<TemplateType>(() => resolveTemplateType(initialData));

  const registryEntry = levelTemplateRegistry[templateType] || levelTemplateRegistry['trivia'];

  const [content, setContent] = useState<unknown>(() => resolveContent(initialData, registryEntry));
  const [errors, setErrors] = useState<unknown>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [showPreview, setShowPreview] = useState(false);

  useEffect(() => {
    const schema = registryEntry.schema;
    if (schema) {
      try {
        const result = schema.safeParse(content);
        setErrors(result.success ? null : result.error.format());
      } catch {
        setErrors({ _errors: ['Error interno de validación en la plantilla.'] });
      }
    }
  }, [content, registryEntry]);

  const handleTemplateChange = (newType: TemplateType) => {
    setTemplateType(newType);
    // Al crear (sin initialData), cambiar de plantilla siempre resetea el
    // contenido a los valores por defecto. Al editar, el metadata form
    // deshabilita el selector de plantilla (isEditing), así que esta rama
    // nunca se alcanza con datos existentes que preservar.
    if (!isEditing) {
      const newEntry = levelTemplateRegistry[newType] || levelTemplateRegistry['trivia'];
      setContent(newEntry.getDefaults());
    }
  };

  const handleSave = async () => {
    if (errors || !title || !sectionId) return;
    setLoading(true);
    setSubmitError(null);
    try {
      await onSave({ section_id: sectionId, title, color, difficulty, template_type: templateType, content });
    } catch (e) {
      const detail = e && typeof e === 'object' && 'response' in e
        ? (e as { response?: { data?: { detail?: string } } }).response?.data?.detail
        : undefined;
      setSubmitError(detail ?? (e instanceof Error ? e.message : 'Error al guardar'));
    } finally {
      setLoading(false);
    }
  };

  const FormComponent = registryEntry.FormComponent as ComponentType<{ value: unknown; onChange: (v: unknown) => void; errors?: unknown }>;
  const PreviewComponent = registryEntry.PreviewComponent as ComponentType<{ value: unknown }>;

  return (
    <div className="bg-[--color-card] text-[--color-text-card] p-6 rounded-lg shadow-sm">
      <h2 className="text-2xl font-bold mb-6">{isEditing ? 'Editar Nivel' : 'Crear Nuevo Nivel'}</h2>

      <LevelMetadataForm
        title={title} setTitle={setTitle}
        color={color} setColor={setColor}
        difficulty={difficulty} setDifficulty={setDifficulty}
        templateType={templateType} setTemplateType={handleTemplateChange}
        sectionId={sectionId} setSectionId={setSectionId}
        sections={sections} isEditing={isEditing}
      />

      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold">Contenido del Nivel</h3>
        <button
          type="button"
          onClick={() => setShowPreview(!showPreview)}
          className="text-sm text-[--color-primary] underline"
        >
          {showPreview ? 'Ocultar Previsualización' : 'Mostrar Previsualización'}
        </button>
      </div>

      <FormErrorBoundary>
        {showPreview ? (
          <PreviewComponent value={content} />
        ) : (
          <FormComponent value={content} onChange={setContent} errors={errors} />
        )}
      </FormErrorBoundary>

      {!!errors && (
        <div className="mt-4 p-3 bg-red-50 text-red-600 rounded text-sm overflow-auto max-h-32">
          La configuración contiene errores y no puede guardarse. Revisa los campos.
        </div>
      )}
      {submitError && (
        <div className="mt-4 p-3 bg-red-500 text-white rounded text-sm overflow-auto max-h-32">
          Error del servidor: {submitError}
        </div>
      )}

      <LevelActions
        onSave={() => void handleSave()}
        onCancel={onCancel}
        loading={loading}
        isValid={!errors && !!title && !!sectionId}
        isEditing={isEditing}
      />
    </div>
  );
}

export function LevelMakerForm(props: LevelMakerFormProps) {
  return (
    <FormErrorBoundary>
      <LevelMakerFormInner {...props} />
    </FormErrorBoundary>
  );
}
