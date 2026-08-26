import { useEffect, useState, useMemo, type ChangeEvent, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { Button } from '../../shared/components/ui/Button';
import { Input } from '../../shared/components/ui/Input';
import { apiClient } from '../../shared/apiClient';
import { errorMessage } from '../../shared/errorMessage';
import type { LevelDTO, LevelSummaryDTO, SectionDTO } from './types';
import { templateTypeLabel } from './types';
import {
  LevelsPageDTOSchema,
  SectionsResponseSchema,
  ArchivedLevelsResponseSchema,
  ArchivedSectionsResponseSchema,
} from './schemas';
import { LevelMakerForm, type LevelMakerFormInitialData, type LevelMakerFormSaveData } from './maker/LevelMakerForm';

interface SectionEditForm {
  id: string;
  title: string;
  description: string;
  color: string;
}

// Formato del archivo que exporta el maker local (frontend/src/features/
// maker/MakerPage.tsx). NO es el mismo formato que acepta POST /levels — ver
// plan/05_Contenido_maker_y_juego.md §7. author/creation_date/id se
// descartan a propósito: `levels` no tiene esas columnas y el id lo asigna
// el servidor.
interface MakerLevelExportFile {
  metadata: {
    title?: string;
    color?: string;
    difficulty?: number;
    template_type?: string;
  };
  content: unknown;
}

export function AdminContentPage() {
  const [sections, setSections] = useState<SectionDTO[]>([]);
  const [levels, setLevels] = useState<LevelSummaryDTO[]>([]);
  const [archivedSections, setArchivedSections] = useState<SectionDTO[]>([]);
  const [archivedLevels, setArchivedLevels] = useState<LevelSummaryDTO[]>([]);
  const [showArchived, setShowArchived] = useState(false);

  const [sectionTitle, setSectionTitle] = useState('');
  const [sectionDescription, setSectionDescription] = useState('');
  const [sectionColor, setSectionColor] = useState('#18529D');
  const [editingSection, setEditingSection] = useState<SectionEditForm | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const [showMaker, setShowMaker] = useState(false);
  const [makerInitialData, setMakerInitialData] = useState<LevelMakerFormInitialData | null>(null);
  const [loadingLevelID, setLoadingLevelID] = useState<string | null>(null);
  const [importedFileWarning, setImportedFileWarning] = useState<string | null>(null);

  const [expandedSections, setExpandedSections] = useState<Record<string, boolean>>({});

  const levelsBySection = useMemo(() => {
    const groups: Record<string, LevelSummaryDTO[]> = {};
    for (const lvl of levels) {
      if (!groups[lvl.section_id]) groups[lvl.section_id] = [];
      groups[lvl.section_id].push(lvl);
    }
    return groups;
  }, [levels]);

  const toggleSection = (id: string) => {
    setExpandedSections((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  async function loadContent() {
    setError(null);
    try {
      const [sectionResp, levelResp] = await Promise.all([
        apiClient.get('/sections?include_unpublished=true'),
        apiClient.get('/levels?include_unpublished=true&page_size=50'),
      ]);
      setSections(SectionsResponseSchema.parse(sectionResp.data).items);
      setLevels(LevelsPageDTOSchema.parse(levelResp.data).items);
    } catch (err) {
      setError(errorMessage(err, 'No se pudo cargar el contenido.'));
    }
  }

  async function loadArchived() {
    setError(null);
    try {
      const [sectionResp, levelResp] = await Promise.all([
        apiClient.get('/sections/archived'),
        apiClient.get('/levels/archived'),
      ]);
      setArchivedSections(ArchivedSectionsResponseSchema.parse(sectionResp.data).items);
      setArchivedLevels(ArchivedLevelsResponseSchema.parse(levelResp.data).items);
    } catch (err) {
      setError(errorMessage(err, 'No se pudo cargar el contenido archivado.'));
    }
  }

  useEffect(() => {
    void loadContent();
  }, []);

  useEffect(() => {
    if (showArchived) void loadArchived();
  }, [showArchived]);

  async function createSection(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await apiClient.post('/sections', { title: sectionTitle, description: sectionDescription, color: sectionColor });
      setSectionTitle('');
      setSectionDescription('');
      await loadContent();
    } catch (err) {
      setError(errorMessage(err, 'No se pudo crear la sección.'));
    } finally {
      setLoading(false);
    }
  }

  async function saveSectionEdit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!editingSection) return;
    setLoading(true);
    setError(null);
    try {
      await apiClient.patch(`/sections/${editingSection.id}`, {
        title: editingSection.title,
        description: editingSection.description,
        color: editingSection.color,
      });
      setEditingSection(null);
      await loadContent();
    } catch (err) {
      setError(errorMessage(err, 'No se pudo actualizar la sección.'));
    } finally {
      setLoading(false);
    }
  }

  async function startLevelEdit(levelID: string) {
    setLoadingLevelID(levelID);
    setError(null);
    try {
      const { data } = await apiClient.get<LevelDTO>(`/levels/${levelID}`);
      setMakerInitialData({
        id: data.id,
        section_id: data.section_id,
        title: data.title,
        color: data.color,
        difficulty: data.difficulty,
        template_type: data.template_type,
        content: data.content,
      });
      setShowMaker(true);
    } catch (err) {
      setError(errorMessage(err, 'No se pudo cargar el nivel para edición.'));
    } finally {
      setLoadingLevelID(null);
    }
  }

  async function handleMakerSave(data: LevelMakerFormSaveData) {
    try {
      if (makerInitialData?.id) {
        await apiClient.patch(`/levels/${makerInitialData.id}`, data);
      } else {
        await apiClient.post('/levels', data);
      }
      setShowMaker(false);
      setMakerInitialData(null);
      setImportedFileWarning(null);
      await loadContent();
    } catch (err) {
      setError(errorMessage(err, 'No se pudo guardar el nivel.'));
      throw err;
    }
  }

  function handleImportCommunityLevel() {
    document.getElementById('import-community-level')?.click();
  }

  // Cierra el ciclo admin → archivo → maker → admin en la otra dirección
  // (plan/05 §7.3): un nivel ya guardado se exporta al mismo formato que
  // produce y acepta el maker local. `author` queda vacío — no hay dónde
  // leerlo, `levels` no tiene esa columna — y `creation_date` se rellena con
  // `created_at`, lo más cercano que existe en el servidor.
  async function exportLevel(levelID: string) {
    setError(null);
    try {
      const { data } = await apiClient.get<LevelDTO>(`/levels/${levelID}`);
      const exportPayload = {
        metadata: {
          id: data.id,
          title: data.title,
          author: '',
          color: data.color,
          difficulty: data.difficulty,
          template_type: data.template_type,
          creation_date: data.created_at,
        },
        content: data.content,
      };
      const jsonStr = JSON.stringify(exportPayload, null, 2);
      const fileName = `${data.title.replace(/\s+/g, '_') || 'nivel_usbi'}.json`;
      const blob = new Blob([jsonStr], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = fileName;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (err) {
      setError(errorMessage(err, 'No se pudo exportar el nivel.'));
    }
  }

  async function onFileSelected(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;

    try {
      const text = await file.text();
      const data = JSON.parse(text) as Partial<MakerLevelExportFile>;

      if (!data?.metadata || data.content === undefined) {
        setError('El archivo seleccionado no tiene el formato correcto (faltan metadata o content).');
        return;
      }

      // El maker local exporta metadata.author y metadata.creation_date, que
      // no existen como columnas en `levels` — se descartan aquí, no en el
      // backend, para que quede a la vista qué se pierde (plan/05 §7.3).
      setImportedFileWarning('Se descartaron autor y fecha de creación del archivo: no se guardan en el servidor. Falta elegir la sección destino.');
      setMakerInitialData({
        section_id: '',
        title: data.metadata.title || '',
        color: data.metadata.color || '#18529D',
        difficulty: data.metadata.difficulty || 1,
        template_type: data.metadata.template_type || 'trivia',
        content: data.content,
      });
      setShowMaker(true);
    } catch {
      setError('El archivo seleccionado no es un JSON válido.');
    } finally {
      e.target.value = '';
    }
  }

  async function publishSection(id: string) { await apiClient.post(`/sections/${id}/publish`); await loadContent(); }
  async function archiveSection(id: string) {
    await apiClient.post(`/sections/${id}/archive`);
    await loadContent();
    if (showArchived) await loadArchived();
  }
  async function publishLevel(id: string) { await apiClient.post(`/levels/${id}/publish`); await loadContent(); }
  async function unpublishSection(id: string) { await apiClient.post(`/sections/${id}/unpublish`); await loadContent(); }
  async function unpublishLevel(id: string) { await apiClient.post(`/levels/${id}/unpublish`); await loadContent(); }
  async function archiveLevel(id: string) {
    await apiClient.post(`/levels/${id}/archive`);
    await loadContent();
    if (showArchived) await loadArchived();
  }

  async function unarchiveSection(id: string) {
    setError(null);
    try {
      await apiClient.post(`/sections/${id}/unarchive`);
      await Promise.all([loadContent(), loadArchived()]);
    } catch (err) {
      setError(errorMessage(err, 'No se pudo restaurar la sección.'));
    }
  }
  async function unarchiveLevel(id: string) {
    setError(null);
    try {
      await apiClient.post(`/levels/${id}/unarchive`);
      await Promise.all([loadContent(), loadArchived()]);
    } catch (err) {
      setError(errorMessage(err, 'No se pudo restaurar el nivel.'));
    }
  }
  async function purgeSection(id: string) {
    setError(null);
    try {
      await apiClient.delete(`/sections/${id}`);
      await loadArchived();
    } catch (err) {
      setError(errorMessage(err, 'No se pudo purgar la sección.'));
    }
  }
  async function purgeLevel(id: string) {
    setError(null);
    try {
      await apiClient.delete(`/levels/${id}`);
      await loadArchived();
    } catch (err) {
      setError(errorMessage(err, 'No se pudo purgar el nivel.'));
    }
  }

  return (
    <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
      <div className="mx-auto max-w-6xl space-y-6">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">Administración de Contenido</h1>
            <p className="text-sm text-[--color-muted]">Secciones y niveles oficiales.</p>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" size="sm">
              <Link to="/">Inicio</Link>
            </Button>
            <Button variant="outline" size="sm">
              <Link to="/maker">Maker local</Link>
            </Button>
          </div>
        </header>

        {error && <p className="rounded border border-[--color-error] bg-[--color-card] p-3 text-[--color-error]">{error}</p>}

        {showMaker ? (
          <LevelMakerForm
            sections={sections}
            initialData={makerInitialData}
            onSave={handleMakerSave}
            onCancel={() => {
              setShowMaker(false);
              setMakerInitialData(null);
              setImportedFileWarning(null);
            }}
          />
        ) : (
          <div className="grid gap-6 lg:grid-cols-2">
            <section className="rounded-lg bg-[--color-card] p-5 shadow-sm">
              <h2 className="mb-4 text-xl font-semibold">Nueva sección</h2>
              <form onSubmit={createSection} className="space-y-4">
                <Input label="Título de sección" value={sectionTitle} onChange={(e) => setSectionTitle(e.currentTarget.value)} required />
                <Input label="Descripción (opcional)" value={sectionDescription} onChange={(e) => setSectionDescription(e.currentTarget.value)} />
                <Input label="Color" type="color" value={sectionColor} onChange={(e) => setSectionColor(e.currentTarget.value)} required />
                <Button type="submit" disabled={loading} className="w-full">Crear sección</Button>
              </form>
            </section>

            <section className="rounded-lg bg-[--color-card] p-5 shadow-sm flex flex-col justify-center items-center">
              <h2 className="mb-4 text-xl font-semibold">Nuevo Nivel Oficial</h2>
              <p className="text-sm text-[--color-muted] mb-4 text-center">Usa el editor visual para configurar niveles con validación completa.</p>
              <div className="flex gap-4 mb-6">
                <Button onClick={() => setShowMaker(true)}>Abrir Creador de Niveles</Button>
                <Button variant="outline" onClick={handleImportCommunityLevel}>Importar Nivel (JSON)</Button>
              </div>
              <div className="text-xs text-[--color-muted] text-center max-w-sm">
                <p className="mb-1"><strong>Aviso:</strong> Los niveles que sean expuestos al público son responsabilidad de la institución.</p>
                <p>Se copiarán los datos válidos del archivo del maker local y se te pedirá añadir la sección a la que corresponde.</p>
              </div>
              <input type="file" accept=".json" style={{ display: 'none' }} id="import-community-level" onChange={(e) => void onFileSelected(e)} />
            </section>
          </div>
        )}

        {!showMaker && importedFileWarning && (
          <p className="rounded border border-[--color-warning] bg-[--color-card] p-3 text-sm text-[--color-warning]">{importedFileWarning}</p>
        )}

        {!showMaker && (
          <section className="rounded-lg bg-[--color-card] p-5 shadow-sm">
            <h2 className="mb-4 text-xl font-semibold">Secciones y Niveles</h2>
            {editingSection && (
              <form onSubmit={saveSectionEdit} className="mb-5 rounded-lg border border-[--color-border] p-4">
                <h3 className="mb-3 font-semibold">Editar sección</h3>
                <div className="grid gap-3 md:grid-cols-[1fr_140px_auto] md:items-end">
                  <Input
                    label="Título de sección"
                    value={editingSection.title}
                    onChange={(e) => setEditingSection({ ...editingSection, title: e.currentTarget.value })}
                    required
                  />
                  <Input
                    label="Descripción"
                    value={editingSection.description}
                    onChange={(e) => setEditingSection({ ...editingSection, description: e.currentTarget.value })}
                  />
                  <Input
                    label="Color"
                    type="color"
                    value={editingSection.color}
                    onChange={(e) => setEditingSection({ ...editingSection, color: e.currentTarget.value })}
                    required
                  />
                  <div className="flex gap-2">
                    <Button type="submit" size="sm" disabled={loading}>Guardar</Button>
                    <Button type="button" size="sm" variant="outline" onClick={() => setEditingSection(null)}>
                      Cancelar
                    </Button>
                  </div>
                </div>
              </form>
            )}
            <div className="divide-y">
              {sections.map((section) => {
                const sectionLevels = levelsBySection[section.id] || [];
                const isExpanded = expandedSections[section.id];
                return (
                  <div key={section.id} className="py-3">
                    <div className="flex flex-wrap items-center justify-between gap-3 p-2 rounded">
                      <button
                        type="button"
                        className="flex items-center gap-2 text-left hover:bg-black/5 dark:hover:bg-white/5 rounded transition-colors -m-1 p-1"
                        onClick={() => toggleSection(section.id)}
                        aria-expanded={isExpanded}
                      >
                        <span className="text-[--color-muted] w-5 text-center text-xs" aria-hidden="true">
                          {isExpanded ? '▼' : '▶'}
                        </span>
                        <span>
                          <p className="font-semibold">{section.title} <span className="text-xs font-normal text-[--color-muted] bg-[--color-surface] px-2 py-0.5 rounded-full ml-2 border border-[--color-border]">{sectionLevels.length} niveles</span></p>
                          <p className="text-sm text-[--color-muted]">{section.description || 'Sin descripción'}</p>
                        </span>
                      </button>
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => setEditingSection({ id: section.id, title: section.title, description: section.description, color: section.color })}
                        >
                          Editar
                        </Button>
                        {!section.is_published && <Button size="sm" onClick={() => void publishSection(section.id)}>Publicar</Button>}
                        {section.is_published && <Button size="sm" variant="outline" onClick={() => void unpublishSection(section.id)}>Ocultar</Button>}
                        <Button size="sm" variant="outline" onClick={() => void archiveSection(section.id)}>Archivar</Button>
                      </div>
                    </div>
                    {isExpanded && (
                      <div className="mt-3 pl-6 pr-2 border-l-2 border-[--color-border] ml-4 bg-[--color-surface]/30 rounded-r-lg">
                        <div className="divide-y divide-dashed border-t border-[--color-border] mt-2">
                          {sectionLevels.map((level) => (
                            <div key={level.id} className="flex flex-wrap items-center justify-between gap-3 py-3">
                              <div>
                                <p className="font-medium text-sm">{level.title}</p>
                                <p className="text-xs text-[--color-muted]">
                                  {templateTypeLabel(level.template_type)} · dificultad {level.difficulty} · {level.is_published ? 'Publicado' : 'Borrador'}
                                </p>
                              </div>
                              <div className="flex gap-2 items-center">
                                <Button
                                  size="sm"
                                  variant="outline"
                                  className="h-8 px-2 text-xs"
                                  disabled={loadingLevelID === level.id}
                                  onClick={() => void startLevelEdit(level.id)}
                                >
                                  {loadingLevelID === level.id ? 'Cargando' : 'Editar'}
                                </Button>
                                <Button size="sm" variant="outline" className="h-8 px-2 text-xs" onClick={() => void exportLevel(level.id)}>
                                  Exportar
                                </Button>
                                {!level.is_published && <Button size="sm" className="h-8 px-2 text-xs" onClick={() => void publishLevel(level.id)}>Publicar</Button>}
                                {level.is_published && <Button size="sm" variant="outline" className="h-8 px-2 text-xs" onClick={() => void unpublishLevel(level.id)}>Ocultar</Button>}
                                <Button size="sm" variant="outline" className="h-8 px-2 text-xs border-[--color-error] text-[--color-error] hover:bg-[--color-error] hover:text-white" onClick={() => void archiveLevel(level.id)}>Archivar</Button>
                              </div>
                            </div>
                          ))}
                          {sectionLevels.length === 0 && <p className="py-3 text-sm text-[--color-muted]">No hay niveles en esta sección.</p>}
                        </div>
                      </div>
                    )}
                  </div>
                );
              })}
              {sections.length === 0 && <p className="py-4 text-sm text-[--color-muted]">No hay secciones.</p>}
            </div>
          </section>
        )}

        {!showMaker && (
          <section className="rounded-lg bg-[--color-card] p-5 shadow-sm">
            <div className="mb-4 flex items-center justify-between">
              <div>
                <h2 className="text-xl font-semibold">Contenido archivado</h2>
                <p className="text-sm text-[--color-muted]">
                  Archivar es reversible y no libera espacio. Purgar es irreversible y sí lo libera —
                  la experiencia (XP) de cada jugador se conserva siempre.
                </p>
              </div>
              <Button variant="outline" size="sm" onClick={() => setShowArchived((v) => !v)}>
                {showArchived ? 'Ocultar' : 'Mostrar archivados'}
              </Button>
            </div>

            {showArchived && (
              <div className="space-y-6">
                <div>
                  <h3 className="mb-2 text-sm font-semibold text-[--color-muted]">Secciones archivadas</h3>
                  {archivedSections.length === 0 && <p className="text-sm text-[--color-muted]">No hay secciones archivadas.</p>}
                  <div className="divide-y">
                    {archivedSections.map((section) => (
                      <ArchivedRow
                        key={section.id}
                        title={section.title}
                        subtitle={section.description || 'Sin descripción'}
                        entityLabel="la sección"
                        onRestore={() => void unarchiveSection(section.id)}
                        onPurge={() => void purgeSection(section.id)}
                      />
                    ))}
                  </div>
                </div>

                <div>
                  <h3 className="mb-2 text-sm font-semibold text-[--color-muted]">Niveles archivados</h3>
                  {archivedLevels.length === 0 && <p className="text-sm text-[--color-muted]">No hay niveles archivados.</p>}
                  <div className="divide-y">
                    {archivedLevels.map((level) => (
                      <ArchivedRow
                        key={level.id}
                        title={level.title}
                        subtitle={templateTypeLabel(level.template_type)}
                        entityLabel="el nivel"
                        onRestore={() => void unarchiveLevel(level.id)}
                        onPurge={() => void purgeLevel(level.id)}
                      />
                    ))}
                  </div>
                </div>
              </div>
            )}
          </section>
        )}
      </div>
    </main>
  );
}

// Doble confirmación real (plan/05 §6.2): un aviso de irreversibilidad
// primero, y el botón de purgar solo se habilita al escribir el título
// exacto. Nada de un confirm() del navegador.
function ArchivedRow({
  title,
  subtitle,
  entityLabel,
  onRestore,
  onPurge,
}: {
  title: string;
  subtitle: string;
  entityLabel: string;
  onRestore: () => void;
  onPurge: () => void;
}) {
  const [confirming, setConfirming] = useState(false);
  const [typedTitle, setTypedTitle] = useState('');
  const canPurge = typedTitle.trim() === title.trim() && title.trim() !== '';

  return (
    <div className="py-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="font-medium text-sm">{title}</p>
          <p className="text-xs text-[--color-muted]">{subtitle}</p>
        </div>
        <div className="flex gap-2 items-center">
          <Button size="sm" variant="outline" className="h-8 px-2 text-xs" onClick={onRestore}>
            Restaurar
          </Button>
          <Button
            size="sm"
            variant="outline"
            className="h-8 px-2 text-xs border-[--color-error] text-[--color-error] hover:bg-[--color-error] hover:text-white"
            onClick={() => setConfirming((v) => !v)}
          >
            Purgar definitivamente
          </Button>
        </div>
      </div>
      {confirming && (
        <div className="mt-2 rounded border border-[--color-error] bg-[--color-surface] p-3 space-y-2">
          <p className="text-sm text-[--color-error]">
            Esta acción es <strong>irreversible</strong> y libera el almacenamiento de {entityLabel}. No se
            puede deshacer. Escribe el título exacto para confirmar.
          </p>
          <div className="flex flex-wrap gap-2 items-end">
            <Input
              label={`Título de confirmación (${title})`}
              value={typedTitle}
              onChange={(e) => setTypedTitle(e.currentTarget.value)}
            />
            <Button
              size="sm"
              disabled={!canPurge}
              className="bg-[var(--color-error)] text-white"
              onClick={() => {
                onPurge();
                setConfirming(false);
                setTypedTitle('');
              }}
            >
              Confirmar purga
            </Button>
            <Button size="sm" variant="outline" onClick={() => { setConfirming(false); setTypedTitle(''); }}>
              Cancelar
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
