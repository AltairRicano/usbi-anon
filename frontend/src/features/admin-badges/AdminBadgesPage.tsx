import { useEffect, useState, type FormEvent } from 'react';
import { Button } from '../../shared/components/ui/Button';
import { HomeButton } from '../../shared/components/ui/HomeButton';
import { Input } from '../../shared/components/ui/Input';
import { apiClient } from '../../shared/apiClient';
import { errorMessage } from '../../shared/errorMessage';
import { BadgesResponseSchema, type Badge } from './schemas';

interface EditForm {
  id: string;
  name: string;
  xp_threshold: number;
  icon_key: string;
}

export default function AdminBadgesPage() {
  const [badges, setBadges] = useState<Badge[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const [newName, setNewName] = useState('');
  const [newThreshold, setNewThreshold] = useState(0);
  const [newIconKey, setNewIconKey] = useState('');
  const [editing, setEditing] = useState<EditForm | null>(null);

  async function loadBadges() {
    setError(null);
    try {
      const resp = await apiClient.get('/admin/badges');
      setBadges(BadgesResponseSchema.parse(resp.data).items);
    } catch (err) {
      setError(errorMessage(err, 'No se pudo cargar el catálogo de insignias.'));
    }
  }

  useEffect(() => {
    void loadBadges();
  }, []);

  async function createBadge(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await apiClient.post('/admin/badges', {
        name: newName,
        xp_threshold: newThreshold,
        icon_key: newIconKey,
      });
      setNewName('');
      setNewThreshold(0);
      setNewIconKey('');
      await loadBadges();
    } catch (err) {
      setError(errorMessage(err, 'No se pudo crear la insignia.'));
    } finally {
      setLoading(false);
    }
  }

  async function saveEdit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!editing) return;
    setLoading(true);
    setError(null);
    try {
      await apiClient.patch(`/admin/badges/${editing.id}`, {
        name: editing.name,
        xp_threshold: editing.xp_threshold,
        icon_key: editing.icon_key,
      });
      setEditing(null);
      await loadBadges();
    } catch (err) {
      setError(errorMessage(err, 'No se pudo actualizar la insignia.'));
    } finally {
      setLoading(false);
    }
  }

  async function deleteBadge(id: string) {
    setError(null);
    try {
      await apiClient.delete(`/admin/badges/${id}`);
      await loadBadges();
    } catch (err) {
      // 409 badge-has-holder: al menos una cuenta ya ganó esta insignia, no
      // se puede borrar (backend/internal/badges/handler.go). errorMessage
      // ya toma el `detail` del RFC 7807, así que basta mostrarlo tal cual.
      setError(errorMessage(err, 'No se pudo eliminar la insignia.'));
    }
  }

  return (
    <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
      <div className="mx-auto max-w-4xl space-y-6">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">Catálogo de insignias</h1>
            <p className="text-sm text-[--color-muted]">{badges.length} insignia{badges.length === 1 ? '' : 's'} definida{badges.length === 1 ? '' : 's'}.</p>
          </div>
          <HomeButton />
        </header>

        {error && (
          <p role="alert" className="rounded-lg border p-3 text-sm" style={{ borderColor: 'var(--color-error)', color: 'var(--color-error)' }}>
            {error}
          </p>
        )}

        <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border]">
          <h2 className="mb-4 text-xl font-semibold">Nueva insignia</h2>
          <form onSubmit={createBadge} className="grid gap-3 md:grid-cols-[1fr_140px_1fr_auto] md:items-end">
            <Input
              id="new-badge-name"
              label="Nombre"
              value={newName}
              onChange={(e) => setNewName(e.currentTarget.value)}
              required
              maxLength={100}
            />
            <Input
              id="new-badge-threshold"
              label="Umbral de XP"
              type="number"
              min={0}
              value={newThreshold}
              onChange={(e) => setNewThreshold(Number(e.currentTarget.value))}
            />
            <Input
              id="new-badge-icon"
              label="Clave de icono"
              value={newIconKey}
              onChange={(e) => setNewIconKey(e.currentTarget.value)}
              required
              maxLength={100}
            />
            <Button type="submit" disabled={loading}>Crear</Button>
          </form>
        </section>

        <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border]">
          <h2 className="mb-4 text-xl font-semibold">Insignias existentes</h2>

          {editing && (
            <form onSubmit={saveEdit} className="mb-5 rounded-lg border border-[--color-border] p-4">
              <h3 className="mb-3 font-semibold">Editar insignia</h3>
              <div className="grid gap-3 md:grid-cols-[1fr_140px_1fr_auto] md:items-end">
                <Input
                  id="edit-badge-name"
                  label="Nombre"
                  value={editing.name}
                  onChange={(e) => setEditing({ ...editing, name: e.currentTarget.value })}
                  required
                  maxLength={100}
                />
                <Input
                  id="edit-badge-threshold"
                  label="Umbral de XP"
                  type="number"
                  min={0}
                  value={editing.xp_threshold}
                  onChange={(e) => setEditing({ ...editing, xp_threshold: Number(e.currentTarget.value) })}
                />
                <Input
                  id="edit-badge-icon"
                  label="Clave de icono"
                  value={editing.icon_key}
                  onChange={(e) => setEditing({ ...editing, icon_key: e.currentTarget.value })}
                  required
                  maxLength={100}
                />
                <div className="flex gap-2">
                  <Button type="submit" size="sm" disabled={loading}>Guardar</Button>
                  <Button type="button" size="sm" variant="outline" onClick={() => setEditing(null)}>Cancelar</Button>
                </div>
              </div>
            </form>
          )}

          <div className="divide-y divide-[--color-border]">
            {badges.map((b) => (
              <div key={b.id} className="flex flex-wrap items-center justify-between gap-3 py-3">
                <div>
                  <p className="font-medium">{b.name}</p>
                  <p className="text-xs text-[--color-muted]">Umbral {b.xp_threshold} XP · icono <code>{b.icon_key}</code></p>
                </div>
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setEditing({ id: b.id, name: b.name, xp_threshold: b.xp_threshold, icon_key: b.icon_key })}
                  >
                    Editar
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    className="border-[--color-error] text-[--color-error] hover:bg-[--color-error] hover:text-white"
                    onClick={() => void deleteBadge(b.id)}
                  >
                    Eliminar
                  </Button>
                </div>
              </div>
            ))}
            {badges.length === 0 && <p className="py-4 text-sm text-[--color-muted]">No hay insignias.</p>}
          </div>
        </section>
      </div>
    </main>
  );
}
