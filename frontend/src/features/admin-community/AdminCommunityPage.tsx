import { useEffect, useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { Button } from '../../shared/components/ui/Button';
import { Input } from '../../shared/components/ui/Input';
import { apiClient } from '../../shared/apiClient';
import { errorMessage } from '../../shared/errorMessage';
import {
  CategoriesResponseSchema,
  LinksResponseSchema,
  SuggestionsPageSchema,
  type InterestLinkCategory,
  type InterestLink,
  type Suggestion,
} from './schemas';

type Tab = 'links' | 'suggestions';

interface CategoryForm {
  id: string;
  name: string;
  display_order: number;
}

interface LinkForm {
  id: string;
  category_id: string;
  title: string;
  description: string;
  color: string;
  url: string;
}

const EMPTY_LINK_FORM = { category_id: '', title: '', description: '', color: '#18529D', url: '' };

export default function AdminCommunityPage() {
  const [tab, setTab] = useState<Tab>('links');

  // ── Enlaces de interés: categorías ──────────────────────────────────
  const [categories, setCategories] = useState<InterestLinkCategory[]>([]);
  const [links, setLinks] = useState<InterestLink[]>([]);
  const [linksError, setLinksError] = useState<string | null>(null);

  const [newCategoryName, setNewCategoryName] = useState('');
  const [newCategoryOrder, setNewCategoryOrder] = useState(0);
  const [editingCategory, setEditingCategory] = useState<CategoryForm | null>(null);

  const [newLink, setNewLink] = useState(EMPTY_LINK_FORM);
  const [editingLink, setEditingLink] = useState<LinkForm | null>(null);

  async function loadCategories() {
    try {
      const resp = await apiClient.get('/admin/interest-link-categories');
      setCategories(CategoriesResponseSchema.parse(resp.data).items);
    } catch (err) {
      setLinksError(errorMessage(err, 'No se pudieron cargar las categorías.'));
    }
  }

  async function loadLinks() {
    try {
      const resp = await apiClient.get('/admin/interest-links');
      setLinks(LinksResponseSchema.parse(resp.data).items);
    } catch (err) {
      setLinksError(errorMessage(err, 'No se pudieron cargar los enlaces.'));
    }
  }

  useEffect(() => {
    void loadCategories();
    void loadLinks();
  }, []);

  async function createCategory(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLinksError(null);
    try {
      await apiClient.post('/admin/interest-link-categories', { name: newCategoryName, display_order: newCategoryOrder });
      setNewCategoryName('');
      setNewCategoryOrder(0);
      await loadCategories();
    } catch (err) {
      setLinksError(errorMessage(err, 'No se pudo crear la categoría.'));
    }
  }

  async function saveCategory(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!editingCategory) return;
    setLinksError(null);
    try {
      await apiClient.patch(`/admin/interest-link-categories/${editingCategory.id}`, {
        name: editingCategory.name,
        display_order: editingCategory.display_order,
      });
      setEditingCategory(null);
      await loadCategories();
    } catch (err) {
      setLinksError(errorMessage(err, 'No se pudo actualizar la categoría.'));
    }
  }

  async function deleteCategory(id: string) {
    setLinksError(null);
    try {
      await apiClient.delete(`/admin/interest-link-categories/${id}`);
      await loadCategories();
    } catch (err) {
      // 409 category-has-links: hay que vaciar o reasignar sus tarjetas primero.
      setLinksError(errorMessage(err, 'No se pudo eliminar la categoría.'));
    }
  }

  async function createLink(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLinksError(null);
    try {
      await apiClient.post('/admin/interest-links', newLink);
      setNewLink(EMPTY_LINK_FORM);
      await loadLinks();
    } catch (err) {
      setLinksError(errorMessage(err, 'No se pudo crear el enlace.'));
    }
  }

  async function saveLink(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!editingLink) return;
    setLinksError(null);
    try {
      await apiClient.patch(`/admin/interest-links/${editingLink.id}`, {
        category_id: editingLink.category_id,
        title: editingLink.title,
        description: editingLink.description,
        color: editingLink.color,
        url: editingLink.url,
      });
      setEditingLink(null);
      await loadLinks();
    } catch (err) {
      setLinksError(errorMessage(err, 'No se pudo actualizar el enlace.'));
    }
  }

  async function deleteLink(id: string) {
    setLinksError(null);
    try {
      await apiClient.delete(`/admin/interest-links/${id}`);
      await loadLinks();
    } catch (err) {
      setLinksError(errorMessage(err, 'No se pudo eliminar el enlace.'));
    }
  }

  function categoryName(id: string): string {
    return categories.find((c) => c.id === id)?.name ?? id;
  }

  // ── Buzón de sugerencias (GET/DELETE /admin/suggestions) ────────────
  const [suggestions, setSuggestions] = useState<Suggestion[]>([]);
  const [suggestionsCursor, setSuggestionsCursor] = useState<string | undefined>(undefined);
  const [suggestionsError, setSuggestionsError] = useState<string | null>(null);
  const [suggestionsLoading, setSuggestionsLoading] = useState(false);
  const [suggestionsLoaded, setSuggestionsLoaded] = useState(false);

  async function loadSuggestions(reset: boolean) {
    setSuggestionsLoading(true);
    setSuggestionsError(null);
    try {
      const params: Record<string, string> = {};
      if (!reset && suggestionsCursor) params.cursor = suggestionsCursor;
      const resp = await apiClient.get('/admin/suggestions', { params });
      const page = SuggestionsPageSchema.parse(resp.data);
      setSuggestions(reset ? page.items : [...suggestions, ...page.items]);
      setSuggestionsCursor(page.next_cursor);
      setSuggestionsLoaded(true);
    } catch (err) {
      setSuggestionsError(errorMessage(err, 'No se pudo cargar el buzón de sugerencias.'));
    } finally {
      setSuggestionsLoading(false);
    }
  }

  function openSuggestionsTab() {
    setTab('suggestions');
    if (!suggestionsLoaded) void loadSuggestions(true);
  }

  async function deleteSuggestion(id: string) {
    setSuggestionsError(null);
    try {
      await apiClient.delete(`/admin/suggestions/${id}`);
      setSuggestions((prev) => prev.filter((s) => s.id !== id));
    } catch (err) {
      setSuggestionsError(errorMessage(err, 'No se pudo eliminar la sugerencia.'));
    }
  }

  return (
    <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
      <div className="mx-auto max-w-4xl space-y-6">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">Comunidad</h1>
            <p className="text-sm text-[--color-muted]">Enlaces de interés del carrusel y buzón de sugerencias anónimo.</p>
          </div>
          <Button variant="outline" size="sm">
            <Link to="/">Dashboard</Link>
          </Button>
        </header>

        <div className="flex bg-[--color-card] rounded-full p-1 border border-[--color-border] w-max shadow-inner">
          <button
            onClick={() => setTab('links')}
            className={`px-6 py-2 rounded-full font-bold transition-all duration-200 ${tab === 'links' ? 'bg-[--color-primary] text-[--color-primary-foreground]' : 'text-[--color-muted]'}`}
          >
            Enlaces de interés
          </button>
          <button
            onClick={openSuggestionsTab}
            className={`px-6 py-2 rounded-full font-bold transition-all duration-200 ${tab === 'suggestions' ? 'bg-[--color-primary] text-[--color-primary-foreground]' : 'text-[--color-muted]'}`}
          >
            Buzón de sugerencias
          </button>
        </div>

        {tab === 'links' && (
          <div className="space-y-6">
            {linksError && (
              <p role="alert" className="rounded-lg border p-3 text-sm" style={{ borderColor: 'var(--color-error)', color: 'var(--color-error)' }}>
                {linksError}
              </p>
            )}

            <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border]">
              <h2 className="mb-4 text-xl font-semibold">Categorías</h2>
              <form onSubmit={createCategory} className="mb-4 grid gap-3 md:grid-cols-[1fr_120px_auto] md:items-end">
                <Input id="new-category-name" label="Nombre" value={newCategoryName} onChange={(e) => setNewCategoryName(e.currentTarget.value)} required maxLength={200} />
                <Input id="new-category-order" label="Orden" type="number" value={newCategoryOrder} onChange={(e) => setNewCategoryOrder(Number(e.currentTarget.value))} />
                <Button type="submit">Crear</Button>
              </form>

              {editingCategory && (
                <form onSubmit={saveCategory} className="mb-4 grid gap-3 rounded-lg border border-[--color-border] p-4 md:grid-cols-[1fr_120px_auto] md:items-end">
                  <Input id="edit-category-name" label="Nombre" value={editingCategory.name} onChange={(e) => setEditingCategory({ ...editingCategory, name: e.currentTarget.value })} required />
                  <Input id="edit-category-order" label="Orden" type="number" value={editingCategory.display_order} onChange={(e) => setEditingCategory({ ...editingCategory, display_order: Number(e.currentTarget.value) })} />
                  <div className="flex gap-2">
                    <Button type="submit" size="sm">Guardar</Button>
                    <Button type="button" size="sm" variant="outline" onClick={() => setEditingCategory(null)}>Cancelar</Button>
                  </div>
                </form>
              )}

              <div className="divide-y divide-[--color-border]">
                {categories.map((c) => (
                  <div key={c.id} className="flex flex-wrap items-center justify-between gap-3 py-2">
                    <p className="font-medium">{c.name} <span className="text-xs text-[--color-muted]">(orden {c.display_order})</span></p>
                    <div className="flex gap-2">
                      <Button size="sm" variant="outline" onClick={() => setEditingCategory({ id: c.id, name: c.name, display_order: c.display_order })}>Editar</Button>
                      <Button size="sm" variant="outline" className="border-[--color-error] text-[--color-error]" onClick={() => void deleteCategory(c.id)}>Eliminar</Button>
                    </div>
                  </div>
                ))}
                {categories.length === 0 && <p className="py-4 text-sm text-[--color-muted]">No hay categorías.</p>}
              </div>
            </section>

            <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border]">
              <h2 className="mb-4 text-xl font-semibold">Tarjetas del carrusel</h2>
              <form onSubmit={createLink} className="mb-4 grid gap-3 md:grid-cols-2">
                <div className="flex flex-col gap-1">
                  <label htmlFor="new-link-category" className="text-sm font-medium">Categoría</label>
                  <select
                    id="new-link-category"
                    value={newLink.category_id}
                    onChange={(e) => setNewLink({ ...newLink, category_id: e.currentTarget.value })}
                    required
                    className="min-h-[44px] rounded-lg border px-4 py-2 text-base border-[--color-border] bg-[--color-background]"
                  >
                    <option value="" disabled>Elige una categoría</option>
                    {categories.map((c) => (
                      <option key={c.id} value={c.id}>{c.name}</option>
                    ))}
                  </select>
                </div>
                <Input id="new-link-color" label="Color (#RRGGBB)" value={newLink.color} onChange={(e) => setNewLink({ ...newLink, color: e.currentTarget.value })} required />
                <Input id="new-link-title" label="Título" value={newLink.title} onChange={(e) => setNewLink({ ...newLink, title: e.currentTarget.value })} required maxLength={50} className="md:col-span-2" />
                <Input id="new-link-description" label="Descripción" value={newLink.description} onChange={(e) => setNewLink({ ...newLink, description: e.currentTarget.value })} required maxLength={100} className="md:col-span-2" />
                <Input id="new-link-url" label="URL (https://…)" value={newLink.url} onChange={(e) => setNewLink({ ...newLink, url: e.currentTarget.value })} required className="md:col-span-2" />
                <Button type="submit" className="md:col-span-2">Crear tarjeta</Button>
              </form>

              {editingLink && (
                <form onSubmit={saveLink} className="mb-4 grid gap-3 rounded-lg border border-[--color-border] p-4 md:grid-cols-2">
                  <div className="flex flex-col gap-1">
                    <label htmlFor="edit-link-category" className="text-sm font-medium">Categoría</label>
                    <select
                      id="edit-link-category"
                      value={editingLink.category_id}
                      onChange={(e) => setEditingLink({ ...editingLink, category_id: e.currentTarget.value })}
                      required
                      className="min-h-[44px] rounded-lg border px-4 py-2 text-base border-[--color-border] bg-[--color-background]"
                    >
                      {categories.map((c) => (
                        <option key={c.id} value={c.id}>{c.name}</option>
                      ))}
                    </select>
                  </div>
                  <Input id="edit-link-color" label="Color" value={editingLink.color} onChange={(e) => setEditingLink({ ...editingLink, color: e.currentTarget.value })} required />
                  <Input id="edit-link-title" label="Título" value={editingLink.title} onChange={(e) => setEditingLink({ ...editingLink, title: e.currentTarget.value })} required className="md:col-span-2" />
                  <Input id="edit-link-description" label="Descripción" value={editingLink.description} onChange={(e) => setEditingLink({ ...editingLink, description: e.currentTarget.value })} required className="md:col-span-2" />
                  <Input id="edit-link-url" label="URL" value={editingLink.url} onChange={(e) => setEditingLink({ ...editingLink, url: e.currentTarget.value })} required className="md:col-span-2" />
                  <div className="flex gap-2 md:col-span-2">
                    <Button type="submit" size="sm">Guardar</Button>
                    <Button type="button" size="sm" variant="outline" onClick={() => setEditingLink(null)}>Cancelar</Button>
                  </div>
                </form>
              )}

              <div className="divide-y divide-[--color-border]">
                {links.map((l) => (
                  <div key={l.id} className="flex flex-wrap items-center justify-between gap-3 py-2">
                    <div>
                      <p className="font-medium">{l.title} <span className="text-xs text-[--color-muted]">({categoryName(l.category_id)})</span></p>
                      <p className="text-xs text-[--color-muted]">{l.description} · {l.url}</p>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => setEditingLink({ id: l.id, category_id: l.category_id, title: l.title, description: l.description, color: l.color, url: l.url })}
                      >
                        Editar
                      </Button>
                      <Button size="sm" variant="outline" className="border-[--color-error] text-[--color-error]" onClick={() => void deleteLink(l.id)}>Eliminar</Button>
                    </div>
                  </div>
                ))}
                {links.length === 0 && <p className="py-4 text-sm text-[--color-muted]">No hay tarjetas.</p>}
              </div>
            </section>
          </div>
        )}

        {tab === 'suggestions' && (
          <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border] space-y-4">
            <p className="text-sm text-[--color-muted]">
              Sugerencias enviadas de forma anónima: no llevan identidad del autor, solo un snapshot de su progreso al momento de enviarla.
            </p>

            {suggestionsError && (
              <p role="alert" className="rounded-lg border p-3 text-sm" style={{ borderColor: 'var(--color-error)', color: 'var(--color-error)' }}>
                {suggestionsError}
              </p>
            )}

            <div className="divide-y divide-[--color-border]">
              {suggestions.map((s) => (
                <div key={s.id} className="flex flex-wrap items-start justify-between gap-3 py-3">
                  <div>
                    <p>{s.description}</p>
                    <p className="text-xs text-[--color-muted]">
                      {new Date(s.submitted_at).toLocaleString()} · {s.levels_completed_snapshot} niveles completados · {s.xp_snapshot} XP
                    </p>
                  </div>
                  <Button size="sm" variant="outline" className="border-[--color-error] text-[--color-error]" onClick={() => void deleteSuggestion(s.id)}>Eliminar</Button>
                </div>
              ))}
              {suggestions.length === 0 && !suggestionsLoading && <p className="py-4 text-sm text-[--color-muted]">Buzón vacío.</p>}
            </div>

            {suggestionsCursor && (
              <div className="flex justify-center">
                <Button type="button" variant="outline" onClick={() => void loadSuggestions(false)} disabled={suggestionsLoading}>
                  {suggestionsLoading ? 'Cargando…' : 'Cargar más'}
                </Button>
              </div>
            )}
          </section>
        )}
      </div>
    </main>
  );
}
