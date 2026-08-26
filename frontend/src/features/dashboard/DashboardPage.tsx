import { useState, useEffect } from 'react';
import { useAuthStore } from '../auth/useAuthStore';
import { Button } from '../../shared/components/ui/Button';
import { UsbiEmblem } from '../../shared/components/ui/Brand';
import { SettingsEntry } from '../../shared/components/SettingsEntry';
import { apiClient } from '../../shared/apiClient';
import { useNavigate, type NavigateFunction } from 'react-router-dom';
import type { SectionDTO, LevelsPageDTO, TemplateType } from '../content/types';
import { templateTypeLabel } from '../content/types';
import { SectionsResponseSchema, LevelsPageDTOSchema } from '../content/schemas';

interface StoredLocalLevel {
  metadata: {
    id: string;
    title: string;
    color?: string;
    difficulty: number;
    template_type: TemplateType;
  };
  content: unknown;
}

function LocalContentSection({ navigate }: { navigate: NavigateFunction }) {
  const [localLevels, setLocalLevels] = useState<StoredLocalLevel[]>([]);

  useEffect(() => {
    const loadLocal = () => {
      const stored = localStorage.getItem('usbi_local_levels');
      if (stored) {
        setLocalLevels(JSON.parse(stored) as StoredLocalLevel[]);
      }
    };
    loadLocal();
    window.addEventListener('storage', loadLocal);
    return () => window.removeEventListener('storage', loadLocal);
  }, []);

  const handleImport = () => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = 'application/json';
    input.onchange = (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (!file) return;
      const reader = new FileReader();
      reader.onload = (re) => {
        if (typeof re.target?.result === 'string') {
          processImportedJSON(re.target.result);
        }
      };
      reader.readAsText(file);
    };
    input.click();
  };

  const processImportedJSON = (jsonStr: string) => {
    try {
      // Validar 5MB de tamaño (string length es buena aproximación para bytes en utf-8 ascii)
      if (jsonStr.length > 5 * 1024 * 1024) {
        alert('El archivo supera el límite de 5MB permitido por seguridad.');
        return;
      }
      const data = JSON.parse(jsonStr) as Partial<StoredLocalLevel>;
      if (!data?.metadata?.id) {
        alert('Archivo inválido. No se encontraron metadatos válidos de un nivel.');
        return;
      }

      const stored = localStorage.getItem('usbi_local_levels');
      const levels: StoredLocalLevel[] = stored ? JSON.parse(stored) : [];
      const existingIdx = levels.findIndex((l) => l.metadata.id === data.metadata!.id);

      if (existingIdx >= 0) {
        levels[existingIdx] = data as StoredLocalLevel;
      } else {
        levels.push(data as StoredLocalLevel);
      }

      localStorage.setItem('usbi_local_levels', JSON.stringify(levels));
      setLocalLevels(levels);
      alert('¡Nivel importado exitosamente!');
    } catch {
      alert('Archivo inválido. No es un JSON correcto.');
    }
  };

  const handleDelete = (id: string) => {
    if (!confirm('¿Estás seguro de eliminar este nivel local?')) return;
    const filtered = localLevels.filter((l) => l.metadata.id !== id);
    localStorage.setItem('usbi_local_levels', JSON.stringify(filtered));
    setLocalLevels(filtered);
  };

  return (
    <section className="rounded-lg bg-[--color-card] text-[--color-text-card] p-6 shadow-sm border border-[--color-border]">
      <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-xl font-semibold text-[--color-primary]">Tus niveles locales (Maker)</h2>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => navigate('/maker')}>
            + Crear nuevo
          </Button>
          <Button variant="outline" size="sm" onClick={handleImport}>
            Importar
          </Button>
        </div>
      </div>

      {localLevels.length === 0 ? (
        <p className="text-[--color-muted] text-sm bg-black/5 dark:bg-white/5 p-4 rounded-lg">
          No tienes niveles locales. Puedes crear uno o importar un archivo JSON.
        </p>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          {localLevels.map((lvl) => (
            <div key={lvl.metadata.id} className="border border-[--color-border] p-4 rounded-lg flex flex-col gap-2 relative bg-black/5 dark:bg-white/5">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => handleDelete(lvl.metadata.id)}
                className="absolute top-2 right-2 text-red-500 hover:text-red-700 font-bold"
                aria-label="Eliminar nivel"
                title="Eliminar nivel"
              >
                ✕
              </Button>
              <div className="flex items-center gap-2 min-w-0">
                <div className="w-8 h-8 shrink-0 rounded-full flex items-center justify-center text-white" style={{ backgroundColor: lvl.metadata.color || '#4caf50' }}>
                  <TemplateIcon type={lvl.metadata.template_type} />
                </div>
                <h3 className="font-bold truncate pr-6" title={lvl.metadata.title}>{lvl.metadata.title}</h3>
              </div>
              <p className="text-xs text-[--color-muted]">Tipo: {templateTypeLabel(lvl.metadata.template_type)} | Dif: {lvl.metadata.difficulty}</p>
              <Button size="sm" variant="primary" className="mt-2 w-full" onClick={() => navigate(`/local-levels/${lvl.metadata.id}/play`)}>
                Jugar local
              </Button>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}

function TemplateIcon({ type }: { type: TemplateType }) {
  switch (type) {
    case 'trivia':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className="w-8 h-8">
          <circle cx="12" cy="12" r="10" />
          <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
          <circle cx="12" cy="17" r="1" />
        </svg>
      );
    case 'puzzle':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className="w-8 h-8">
          <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
          <path d="M3.27 6.96L12 12.01l8.73-5.05" />
          <path d="M12 22.08V12" />
        </svg>
      );
    case 'word_search':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className="w-8 h-8">
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.3-4.3" />
          <path d="M7 11h8M7 7h8M7 15h4" />
        </svg>
      );
    case 'fake_news':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className="w-8 h-8">
          <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
          <path d="M12 8v4" />
          <path d="M12 16h.01" />
        </svg>
      );
    case 'crossword':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className="w-8 h-8">
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
          <path d="M3 9h18M3 15h18M9 3v18M15 3v18" />
        </svg>
      );
    case 'memory':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className="w-8 h-8">
          <rect x="3" y="4" width="7" height="16" rx="1" />
          <rect x="14" y="4" width="7" height="16" rx="1" />
          <path d="M5 8h3M5 12h3M16 8h3M16 12h3" />
        </svg>
      );
    case 'snakes_ladders':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className="w-8 h-8">
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
          <circle cx="8.5" cy="8.5" r="1.5" />
          <circle cx="15.5" cy="15.5" r="1.5" />
          <circle cx="15.5" cy="8.5" r="1.5" />
          <circle cx="8.5" cy="15.5" r="1.5" />
        </svg>
      );
    default:
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className="w-8 h-8">
          <circle cx="12" cy="12" r="10" />
          <path d="M12 8v8M8 12h8" />
        </svg>
      );
  }
}

function SectionAccordionItem({ section, navigate }: { section: SectionDTO; navigate: NavigateFunction }) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [levels, setLevels] = useState<LevelsPageDTO['items']>([]);
  const [loading, setLoading] = useState(false);
  const [loadError, setLoadError] = useState(false);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    if (!isExpanded || loaded) return;
    setLoading(true);
    setLoadError(false);
    apiClient
      .get(`/levels?section_id=${section.id}&page_size=50`)
      .then((resp) => {
        setLevels(LevelsPageDTOSchema.parse(resp.data).items);
        setLoaded(true);
      })
      .catch(() => setLoadError(true))
      .finally(() => setLoading(false));
  }, [isExpanded, loaded, section.id]);

  return (
    <article className="rounded-xl bg-[--color-card] text-[--color-text-card] overflow-hidden shadow-md transition-all duration-300">
      <div
        role="button"
        tabIndex={0}
        aria-expanded={isExpanded}
        className="flex items-center justify-between p-5 cursor-pointer hover:bg-black/5 dark:hover:bg-white/5 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[--color-primary] focus-visible:ring-inset"
        onClick={() => setIsExpanded(!isExpanded)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            setIsExpanded((v) => !v);
          }
        }}
      >
        <div className="flex items-center gap-4 min-w-0">
          <div className="w-14 h-14 shrink-0 rounded-full shadow-inner flex items-center justify-center text-white font-bold text-2xl" style={{ backgroundColor: section.color || '#4caf50' }}>
            {section.title.charAt(0).toUpperCase()}
          </div>
          <div className="min-w-0">
            <h3 className="text-2xl font-bold truncate">{section.title}</h3>
            {section.description && (
              <p className="text-sm text-[--color-muted] mt-1 truncate">{section.description}</p>
            )}
          </div>
        </div>
        <div className="text-[--color-muted]">
          <svg className={`w-8 h-8 transition-transform duration-300 ${isExpanded ? 'rotate-180' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </div>

      {isExpanded && (
        <div className="p-8 pb-12 bg-[--color-card]">
          {loading ? (
            <p className="text-center text-[--color-muted]">Cargando niveles...</p>
          ) : loadError ? (
            <p className="text-center text-[--color-error]">Error al cargar los niveles.</p>
          ) : levels.length === 0 ? (
            <p className="text-center text-[--color-muted]">No hay niveles publicados en esta sección.</p>
          ) : (
            <div className="flex flex-col gap-10 items-center w-full">
              {levels.map((level, index) => {
                // Generar un pequeño desplazamiento lateral para crear un camino (zigzag leve)
                const offset = index % 2 === 0 ? '-translate-x-8' : 'translate-x-8';

                return (
                  <div key={level.id} className={`flex flex-col items-center gap-3 group relative ${offset} transform transition-transform`}>
                    <button
                      onClick={() => navigate(`/levels/${level.id}/play`)}
                      className="w-20 h-20 rounded-full shadow-lg flex items-center justify-center text-white transition-transform hover:scale-110 active:scale-95 z-10"
                      style={{ backgroundColor: level.color || section.color || '#4caf50' }}
                    >
                      <TemplateIcon type={level.template_type} />
                    </button>
                    <span className="text-sm font-semibold text-center w-32 break-words">{level.title}</span>
                    <div className="absolute -top-10 scale-0 group-hover:scale-100 transition-transform bg-black/80 backdrop-blur text-white text-xs rounded-lg py-1.5 px-3 pointer-events-none whitespace-nowrap z-20 shadow-xl">
                      {templateTypeLabel(level.template_type)} • Dificultad {level.difficulty}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}
    </article>
  );
}

export default function DashboardPage() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const navigate = useNavigate();
  const isAdmin = user?.role === 'admin';

  const [activeTab, setActiveTab] = useState<'public' | 'local'>('public');
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [sections, setSections] = useState<SectionDTO[]>([]);
  const [loadError, setLoadError] = useState<string | null>(null);

  // Cerrar el menú lateral con Escape (patrón modal-escape / accesibilidad).
  useEffect(() => {
    if (!isMenuOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setIsMenuOpen(false);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [isMenuOpen]);

  useEffect(() => {
    apiClient
      .get('/sections')
      .then((resp) => setSections(SectionsResponseSchema.parse(resp.data).items))
      .catch(() => setLoadError('No se pudo cargar el contenido oficial.'));
  }, []);

  async function handleLogout() {
    try {
      await apiClient.post('/auth/logout');
    } finally {
      logout();
    }
  }

  return (
    <main className="min-h-screen p-8" style={{ backgroundColor: 'var(--color-surface)' }}>
      <div className="max-w-6xl mx-auto space-y-6">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsMenuOpen(true)}
                className="!px-2 !min-w-[44px]"
                aria-label="Menú principal"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className="w-5 h-5">
                  <line x1="4" x2="20" y1="12" y2="12" />
                  <line x1="4" x2="20" y1="6" y2="6" />
                  <line x1="4" x2="20" y1="18" y2="18" />
                </svg>
              </Button>

              {/* aria-hidden: el cierre por teclado lo cubren el listener Escape y el botón ✕ del panel */}
              <button
                type="button"
                aria-hidden="true"
                tabIndex={-1}
                className={`fixed inset-0 z-40 bg-black/40 backdrop-blur-sm transition-opacity duration-300 ${isMenuOpen ? 'opacity-100' : 'opacity-0 pointer-events-none'}`}
                onClick={() => setIsMenuOpen(false)}
              />

              <div
                className={`fixed top-0 left-0 bottom-0 w-72 border-r border-[--color-border] shadow-2xl z-50 p-6 flex flex-col gap-4 transform transition-transform duration-300 ease-in-out overflow-y-auto ${isMenuOpen ? 'translate-x-0' : '-translate-x-full'}`}
                style={{ backgroundColor: 'var(--color-card)', color: 'var(--theme-text-card)' }}
              >
                <div className="flex items-center justify-between mb-8">
                  <h2 className="text-2xl font-bold text-[--color-primary]">Menú</h2>
                  <Button variant="ghost" size="sm" onClick={() => setIsMenuOpen(false)} className="!px-2" aria-label="Cerrar menú">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className="w-6 h-6">
                      <line x1="18" x2="6" y1="6" y2="18" />
                      <line x1="6" x2="18" y1="6" y2="18" />
                    </svg>
                  </Button>
                </div>

                <div className="flex flex-col gap-2">
                  <Button variant="outline" onClick={() => { setIsMenuOpen(false); navigate('/perfil'); }} className="justify-start w-full text-lg py-6">
                    Perfil
                  </Button>
                  <Button variant="outline" onClick={() => { setIsMenuOpen(false); navigate('/maker'); }} className="justify-start w-full text-lg py-6">
                    Maker
                  </Button>
                  <Button variant="outline" onClick={() => { setIsMenuOpen(false); navigate('/settings'); }} className="justify-start w-full text-lg py-6">
                    Configuración
                  </Button>
                  {isAdmin && (
                    <>
                      <p className="mt-2 px-1 text-xs font-semibold uppercase text-[--color-muted]">Administración</p>
                      <Button variant="primary" onClick={() => { setIsMenuOpen(false); navigate('/admin/content'); }} className="justify-start w-full text-lg py-6">
                        Contenido
                      </Button>
                      <Button variant="outline" onClick={() => { setIsMenuOpen(false); navigate('/admin/accounts'); }} className="justify-start w-full text-lg py-6">
                        Cuentas
                      </Button>
                      <Button variant="outline" onClick={() => { setIsMenuOpen(false); navigate('/admin/registration-questions'); }} className="justify-start w-full text-lg py-6">
                        Banco de preguntas
                      </Button>
                    </>
                  )}
                </div>

                <div className="mt-auto">
                  <Button variant="outline" onClick={() => void handleLogout()} className="justify-start w-full text-lg py-6 text-red-500 border-red-200 hover:bg-red-50 dark:hover:bg-red-900/20 dark:border-red-900/50">
                    Cerrar sesión
                  </Button>
                </div>
              </div>
            </div>
            <UsbiEmblem size={40} className="hidden sm:inline-flex" />
            <h1 className="text-3xl font-bold" style={{ color: 'var(--color-primary)' }}>
              Bienvenido, {user?.display_alias ?? user?.nickname}
            </h1>
          </div>
          <SettingsEntry />
        </header>

        {loadError && <p className="rounded border border-[--color-error] bg-[--color-card] p-3 text-[--color-error]">{loadError}</p>}

        <div className="flex bg-[--color-surface] rounded-full p-1 border border-[--color-border] w-max shadow-inner mx-auto mb-8">
          <button
            onClick={() => setActiveTab('public')}
            className={`px-8 py-2 rounded-full font-bold transition-all duration-200 ${
              activeTab === 'public'
                ? 'bg-[--color-primary] text-[--color-primary-foreground] shadow-md'
                : 'text-[--color-muted] hover:text-[--color-foreground]'
            }`}
          >
            Públicos
          </button>
          <button
            onClick={() => setActiveTab('local')}
            className={`px-8 py-2 rounded-full font-bold transition-all duration-200 ${
              activeTab === 'local'
                ? 'bg-[--color-primary] text-[--color-primary-foreground] shadow-md'
                : 'text-[--color-muted] hover:text-[--color-foreground]'
            }`}
          >
            Míos
          </button>
        </div>

        {activeTab === 'public' ? (
          <section className="rounded-lg bg-[--color-card] text-[--color-text-card] p-6 shadow-sm" aria-label="Secciones oficiales">
            <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
              <h2 className="text-xl font-semibold">Secciones oficiales</h2>
              {isAdmin && (
                <Button variant="outline" size="sm" onClick={() => navigate('/admin/content')}>
                  Crear contenido
                </Button>
              )}
            </div>
            <div className="flex flex-col gap-6">
              {sections.map((section) => (
                <SectionAccordionItem key={section.id} section={section} navigate={navigate} />
              ))}
              {sections.length === 0 && (
                <p className="rounded-lg border border-[--color-border] bg-[--color-card] p-5 text-[--color-muted]">
                  No hay secciones publicadas.
                </p>
              )}
            </div>
          </section>
        ) : (
          <LocalContentSection navigate={navigate} />
        )}
      </div>
    </main>
  );
}
