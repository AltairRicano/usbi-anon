import { useEffect, useState, type FormEvent } from 'react';
import { Button } from '../../shared/components/ui/Button';
import { HomeButton } from '../../shared/components/ui/HomeButton';
import { LinkButton } from '../../shared/components/ui/LinkButton';
import { Input } from '../../shared/components/ui/Input';
import { apiClient } from '../../shared/apiClient';
import { errorMessage } from '../../shared/errorMessage';
import { CurrentSettingsSchema, RegistrationQuestionsResponseSchema, type RegistrationQuestion } from './schemas';
import { MinQuestionsModal } from './MinQuestionsModal';

interface EditForm {
  id: string;
  question_text: string;
  is_active: boolean;
  display_order: number;
}

export default function AdminQuizBankPage() {
  const [questions, setQuestions] = useState<RegistrationQuestion[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [showMinQuestionsModal, setShowMinQuestionsModal] = useState(false);

  const [newText, setNewText] = useState('');
  const [newOrder, setNewOrder] = useState(0);
  const [editing, setEditing] = useState<EditForm | null>(null);

  const [maxQuestionsShown, setMaxQuestionsShown] = useState<number | null>(null);
  const [maxQuestionsShownInput, setMaxQuestionsShownInput] = useState('5');
  const [savingSettings, setSavingSettings] = useState(false);

  async function loadQuestions() {
    setError(null);
    try {
      const resp = await apiClient.get('/admin/registration-questions');
      setQuestions(RegistrationQuestionsResponseSchema.parse(resp.data).items);
    } catch (err) {
      setError(errorMessage(err, 'No se pudo cargar el banco de preguntas.'));
    }
  }

  async function loadCurrentMaxQuestionsShown() {
    // El backend no expone GET /admin/registration-settings (solo PUT, ver
    // plan/04 §3): el valor vigente se lee del endpoint público de registro,
    // que también lo devuelve.
    try {
      const resp = await apiClient.post('/auth/register/questions');
      const { max_questions_shown } = CurrentSettingsSchema.parse(resp.data);
      setMaxQuestionsShown(max_questions_shown);
      setMaxQuestionsShownInput(String(max_questions_shown));
    } catch {
      // No crítico: si falla, el campo queda editable con el valor por
      // defecto y el admin puede sobrescribirlo igual.
    }
  }

  useEffect(() => {
    void loadQuestions();
    void loadCurrentMaxQuestionsShown();
  }, []);

  async function createQuestion(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await apiClient.post('/admin/registration-questions', {
        question_text: newText,
        is_active: true,
        display_order: newOrder,
      });
      setNewText('');
      setNewOrder(0);
      await loadQuestions();
    } catch (err) {
      setError(errorMessage(err, 'No se pudo crear la pregunta.'));
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
      await apiClient.patch(`/admin/registration-questions/${editing.id}`, {
        question_text: editing.question_text,
        is_active: editing.is_active,
        display_order: editing.display_order,
      });
      setEditing(null);
      await loadQuestions();
    } catch (err) {
      if (isMinActiveQuestionsConflict(err)) {
        setShowMinQuestionsModal(true);
      } else {
        setError(errorMessage(err, 'No se pudo actualizar la pregunta.'));
      }
    } finally {
      setLoading(false);
    }
  }

  async function deleteQuestion(id: string) {
    setError(null);
    try {
      await apiClient.delete(`/admin/registration-questions/${id}`);
      await loadQuestions();
    } catch (err) {
      if (isMinActiveQuestionsConflict(err)) {
        setShowMinQuestionsModal(true);
      } else {
        setError(errorMessage(err, 'No se pudo eliminar la pregunta.'));
      }
    }
  }

  async function toggleActive(q: RegistrationQuestion) {
    setError(null);
    try {
      await apiClient.patch(`/admin/registration-questions/${q.id}`, {
        question_text: q.question_text,
        is_active: !q.is_active,
        display_order: q.display_order,
      });
      await loadQuestions();
    } catch (err) {
      if (isMinActiveQuestionsConflict(err)) {
        setShowMinQuestionsModal(true);
      } else {
        setError(errorMessage(err, 'No se pudo cambiar el estado de la pregunta.'));
      }
    }
  }

  async function saveSettings(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const parsed = Number(maxQuestionsShownInput);
    if (!Number.isInteger(parsed) || parsed < 4 || parsed > 10) {
      setError('El número de preguntas mostradas debe estar entre 4 y 10.');
      return;
    }
    setSavingSettings(true);
    setError(null);
    try {
      await apiClient.put('/admin/registration-settings', { max_questions_shown: parsed });
      setMaxQuestionsShown(parsed);
    } catch (err) {
      setError(errorMessage(err, 'No se pudo actualizar la configuración.'));
    } finally {
      setSavingSettings(false);
    }
  }

  const activeCount = questions.filter((q) => q.is_active).length;

  return (
    <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
      <MinQuestionsModal open={showMinQuestionsModal} onClose={() => setShowMinQuestionsModal(false)} />

      <div className="mx-auto max-w-4xl space-y-6">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">Banco de preguntas de registro</h1>
            <p className="text-sm text-[--color-muted]">
              {activeCount} pregunta{activeCount === 1 ? '' : 's'} activa{activeCount === 1 ? '' : 's'} (mínimo 4).
            </p>
          </div>
          <div className="flex gap-2">
            <HomeButton />
            <LinkButton to="/admin/accounts">Administrar cuentas</LinkButton>
          </div>
        </header>

        {error && (
          <p role="alert" className="rounded-lg border p-3 text-sm" style={{ borderColor: 'var(--color-error)', color: 'var(--color-error)' }}>
            {error}
          </p>
        )}

        <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border]">
          <h2 className="mb-4 text-xl font-semibold">Cuántas preguntas ve cada registro</h2>
          <form onSubmit={saveSettings} className="flex flex-wrap items-end gap-3">
            <Input
              id="max-questions-shown"
              label="Preguntas mostradas por registro (4–10)"
              type="number"
              min={4}
              max={10}
              value={maxQuestionsShownInput}
              onChange={(e) => setMaxQuestionsShownInput(e.currentTarget.value)}
              className="max-w-[10rem]"
            />
            <Button type="submit" disabled={savingSettings}>
              {savingSettings ? 'Guardando…' : 'Guardar'}
            </Button>
            {maxQuestionsShown !== null && (
              <span className="text-sm text-[--color-muted]">Valor actual: {maxQuestionsShown}</span>
            )}
          </form>
        </section>

        <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border]">
          <h2 className="mb-4 text-xl font-semibold">Nueva pregunta</h2>
          <form onSubmit={createQuestion} className="grid gap-3 md:grid-cols-[1fr_140px_auto] md:items-end">
            <Input
              id="new-question-text"
              label="Texto de la pregunta"
              value={newText}
              onChange={(e) => setNewText(e.currentTarget.value)}
              required
              maxLength={280}
            />
            <Input
              id="new-question-order"
              label="Orden"
              type="number"
              value={newOrder}
              onChange={(e) => setNewOrder(Number(e.currentTarget.value))}
            />
            <Button type="submit" disabled={loading}>Crear</Button>
          </form>
        </section>

        <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border]">
          <h2 className="mb-4 text-xl font-semibold">Preguntas existentes</h2>

          {editing && (
            <form onSubmit={saveEdit} className="mb-5 rounded-lg border border-[--color-border] p-4">
              <h3 className="mb-3 font-semibold">Editar pregunta</h3>
              <div className="grid gap-3 md:grid-cols-[1fr_100px_auto] md:items-end">
                <Input
                  id="edit-question-text"
                  label="Texto"
                  value={editing.question_text}
                  onChange={(e) => setEditing({ ...editing, question_text: e.currentTarget.value })}
                  required
                  maxLength={280}
                />
                <Input
                  id="edit-question-order"
                  label="Orden"
                  type="number"
                  value={editing.display_order}
                  onChange={(e) => setEditing({ ...editing, display_order: Number(e.currentTarget.value) })}
                />
                <div className="flex gap-2">
                  <Button type="submit" size="sm" disabled={loading}>Guardar</Button>
                  <Button type="button" size="sm" variant="outline" onClick={() => setEditing(null)}>Cancelar</Button>
                </div>
              </div>
              <label className="mt-3 flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={editing.is_active}
                  onChange={(e) => setEditing({ ...editing, is_active: e.currentTarget.checked })}
                  className="h-5 w-5"
                />
                Activa
              </label>
            </form>
          )}

          <div className="divide-y divide-[--color-border]">
            {questions
              .slice()
              .sort((a, b) => a.display_order - b.display_order)
              .map((q) => (
                <div key={q.id} className="flex flex-wrap items-center justify-between gap-3 py-3">
                  <div>
                    <p className="font-medium">
                      {q.question_text}{' '}
                      <span
                        className="ml-2 rounded-full px-2 py-0.5 text-xs font-normal"
                        style={{
                          backgroundColor: q.is_active ? 'color-mix(in srgb, var(--color-secondary) 20%, transparent)' : 'var(--color-surface)',
                          color: q.is_active ? 'var(--color-secondary-dark)' : 'var(--color-muted)',
                          border: '1px solid var(--color-border)',
                        }}
                      >
                        {q.is_active ? 'Activa' : 'Inactiva'}
                      </span>
                    </p>
                    <p className="text-xs text-[--color-muted]">Orden {q.display_order}</p>
                  </div>
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() =>
                        setEditing({ id: q.id, question_text: q.question_text, is_active: q.is_active, display_order: q.display_order })
                      }
                    >
                      Editar
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => void toggleActive(q)}>
                      {q.is_active ? 'Desactivar' : 'Activar'}
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      className="border-[--color-error] text-[--color-error] hover:bg-[--color-error] hover:text-white"
                      onClick={() => void deleteQuestion(q.id)}
                    >
                      Eliminar
                    </Button>
                  </div>
                </div>
              ))}
            {questions.length === 0 && <p className="py-4 text-sm text-[--color-muted]">No hay preguntas.</p>}
          </div>
        </section>
      </div>
    </main>
  );
}

function isMinActiveQuestionsConflict(err: unknown): boolean {
  return (
    typeof err === 'object' &&
    err !== null &&
    'response' in err &&
    (err as { response?: { status?: number } }).response?.status === 409
  );
}
