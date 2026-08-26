import { useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { z } from 'zod';
import { Button } from '../../shared/components/ui/Button';
import { Input } from '../../shared/components/ui/Input';
import { apiClient } from '../../shared/apiClient';
import { errorMessage } from '../../shared/errorMessage';
import { NicknameSchema } from '../auth/schemas';
import { UserRoleSchema } from '../../shared/schemas';
import {
  AdminAccountResponseSchema,
  AdminQuizAnswersResponseSchema,
  AdminResetPasswordResponseSchema,
  type AdminAccountResponse,
} from './schemas';

type UserRole = z.infer<typeof UserRoleSchema>;

const STAFF_ROLES: UserRole[] = ['player', 'admin'];

interface QuizAnswerRow {
  question_text_snapshot: string;
  answer_text: string;
  created_at: string;
}

export default function AdminAccountsPage() {
  // ── Crear cuenta de staff ──────────────────────────────────────────────
  const [newNickname, setNewNickname] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [newRole, setNewRole] = useState<UserRole>('player');
  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const [created, setCreated] = useState<AdminAccountResponse | null>(null);

  // ── Gestionar cuenta existente por UUID ────────────────────────────────
  // El contrato de API (plan/04 §3) no incluye un endpoint de listado de
  // cuentas — solo acciones puntuales sobre un account_id conocido. No hay
  // forma de "deshabilitar de antemano" el borrado de un admin sin ese
  // listado: el backend lo rechaza con 403 (ErrCannotDeleteAdmin) y esta
  // página lo refleja tal cual llega, en vez de adivinar el rol antes de
  // intentarlo.
  const [targetID, setTargetID] = useState('');
  const [manageError, setManageError] = useState<string | null>(null);
  const [managing, setManaging] = useState(false);
  const [quizAnswers, setQuizAnswers] = useState<QuizAnswerRow[] | null>(null);
  const [newIssuedPassword, setNewIssuedPassword] = useState<string | null>(null);
  const [deleteConfirmed, setDeleteConfirmed] = useState(false);

  async function handleCreate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setCreateError(null);
    setCreated(null);

    const nicknameCheck = NicknameSchema.safeParse(newNickname);
    if (!nicknameCheck.success) {
      setCreateError(nicknameCheck.error.issues[0]?.message ?? 'Nickname inválido.');
      return;
    }
    if (newPassword.length < 8) {
      setCreateError('La contraseña debe tener al menos 8 caracteres.');
      return;
    }

    setCreating(true);
    try {
      const resp = await apiClient.post('/admin/accounts', {
        nickname: nicknameCheck.data,
        password: newPassword,
        role: newRole,
      });
      setCreated(AdminAccountResponseSchema.parse(resp.data));
      setNewNickname('');
      setNewPassword('');
    } catch (err) {
      setCreateError(errorMessage(err, 'No se pudo crear la cuenta.'));
    } finally {
      setCreating(false);
    }
  }

  function resetManagePanels() {
    setManageError(null);
    setQuizAnswers(null);
    setNewIssuedPassword(null);
    setDeleteConfirmed(false);
  }

  async function handleViewQuizAnswers() {
    resetManagePanels();
    if (!targetID) return;
    setManaging(true);
    try {
      const resp = await apiClient.get(`/admin/accounts/${targetID}/quiz-answers`);
      setQuizAnswers(AdminQuizAnswersResponseSchema.parse(resp.data).items);
    } catch (err) {
      setManageError(errorMessage(err, 'No se pudieron leer las respuestas del cuestionario.'));
    } finally {
      setManaging(false);
    }
  }

  async function handleResetPassword() {
    resetManagePanels();
    if (!targetID) return;
    setManaging(true);
    try {
      const resp = await apiClient.post(`/admin/accounts/${targetID}/reset-password`);
      setNewIssuedPassword(AdminResetPasswordResponseSchema.parse(resp.data).new_password);
    } catch (err) {
      setManageError(errorMessage(err, 'No se pudo resetear la contraseña.'));
    } finally {
      setManaging(false);
    }
  }

  async function handleDelete() {
    resetManagePanels();
    if (!targetID) return;
    setManaging(true);
    try {
      await apiClient.delete(`/admin/accounts/${targetID}`);
      setDeleteConfirmed(true);
      setTargetID('');
    } catch (err) {
      setManageError(errorMessage(err, 'No se pudo eliminar la cuenta.'));
    } finally {
      setManaging(false);
    }
  }

  return (
    <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
      <div className="mx-auto max-w-4xl space-y-6">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">Administración de cuentas</h1>
            <p className="text-sm text-[--color-muted]">Alta de staff, borrado, respuestas del cuestionario y reseteo de contraseña.</p>
          </div>
          <Button variant="outline" size="sm">
            <Link to="/admin/registration-questions">Banco de preguntas</Link>
          </Button>
        </header>

        <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border]">
          <h2 className="mb-4 text-xl font-semibold">Crear cuenta de staff</h2>
          <p className="mb-4 text-sm text-[--color-muted]">
            Sin cuestionario de gustos: nickname y contraseña se fijan aquí directamente.
          </p>

          {createError && (
            <p role="alert" className="mb-4 rounded-lg border p-3 text-sm" style={{ borderColor: 'var(--color-error)', color: 'var(--color-error)' }}>
              {createError}
            </p>
          )}

          {created && (
            <dl className="mb-4 space-y-1 rounded-lg border p-4 text-sm" style={{ borderColor: 'var(--color-secondary-dark)' }}>
              <p className="font-semibold" style={{ color: 'var(--color-secondary-dark)' }}>Cuenta creada.</p>
              <div><dt className="inline text-[--color-muted]">ID: </dt><dd className="inline font-mono">{created.id}</dd></div>
              <div><dt className="inline text-[--color-muted]">Nickname: </dt><dd className="inline font-mono">{created.nickname}</dd></div>
              <div><dt className="inline text-[--color-muted]">Rol: </dt><dd className="inline">{created.role}</dd></div>
              <div><dt className="inline text-[--color-muted]">Alias: </dt><dd className="inline">{created.display_alias}</dd></div>
            </dl>
          )}

          <form onSubmit={handleCreate} className="grid gap-3 md:grid-cols-2">
            <Input
              id="new-account-nickname"
              label="Nickname (6-20, minúsculas y dígitos)"
              value={newNickname}
              onChange={(e) => setNewNickname(e.currentTarget.value)}
              required
              maxLength={20}
            />
            <Input
              id="new-account-password"
              label="Contraseña (mínimo 8 caracteres)"
              type="text"
              value={newPassword}
              onChange={(e) => setNewPassword(e.currentTarget.value)}
              required
              minLength={8}
            />
            <div className="flex flex-col gap-1">
              <label htmlFor="new-account-role" className="text-sm font-medium text-[--color-foreground]">Rol</label>
              <select
                id="new-account-role"
                value={newRole}
                onChange={(e) => setNewRole(e.currentTarget.value as UserRole)}
                className="min-h-[44px] rounded-lg border px-4 py-2 text-base border-[--color-border] bg-[--color-background]"
              >
                {STAFF_ROLES.map((role) => (
                  <option key={role} value={role}>{role}</option>
                ))}
              </select>
            </div>
            <div className="flex items-end">
              <Button type="submit" className="w-full" disabled={creating}>
                {creating ? 'Creando…' : 'Crear cuenta'}
              </Button>
            </div>
          </form>
        </section>

        <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border]">
          <h2 className="mb-4 text-xl font-semibold">Gestionar cuenta existente</h2>
          <Input
            id="target-account-id"
            label="ID de la cuenta (UUID)"
            value={targetID}
            onChange={(e) => setTargetID(e.currentTarget.value)}
            placeholder="00000000-0000-0000-0000-000000000000"
          />

          <div className="mt-3 flex flex-wrap gap-2">
            <Button type="button" variant="outline" disabled={!targetID || managing} onClick={() => void handleViewQuizAnswers()}>
              Ver respuestas del cuestionario
            </Button>
            <Button type="button" variant="outline" disabled={!targetID || managing} onClick={() => void handleResetPassword()}>
              Resetear contraseña
            </Button>
            <Button
              type="button"
              variant="danger"
              disabled={!targetID || managing}
              onClick={() => void handleDelete()}
            >
              Eliminar cuenta
            </Button>
          </div>
          <p className="mt-2 text-xs text-[--color-muted]">
            Una cuenta con rol admin no puede eliminarse desde aquí ni desde ningún otro camino — el servidor lo
            rechaza siempre, sin excepción.
          </p>

          {manageError && (
            <p role="alert" className="mt-4 rounded-lg border p-3 text-sm" style={{ borderColor: 'var(--color-error)', color: 'var(--color-error)' }}>
              {manageError}
            </p>
          )}

          {deleteConfirmed && (
            <p className="mt-4 rounded-lg border p-3 text-sm" style={{ borderColor: 'var(--color-secondary-dark)', color: 'var(--color-secondary-dark)' }}>
              Cuenta eliminada (seudonimizada, progreso purgado).
            </p>
          )}

          {newIssuedPassword && (
            <p className="mt-4 rounded-lg border p-3 text-sm" style={{ borderColor: 'var(--color-warning)', color: 'var(--color-warning)' }}>
              Nueva contraseña (se muestra una sola vez): <span className="font-mono">{newIssuedPassword}</span>
            </p>
          )}

          {quizAnswers && (
            <div className="mt-4 space-y-2">
              <h3 className="font-semibold">Respuestas del cuestionario</h3>
              {quizAnswers.length === 0 && <p className="text-sm text-[--color-muted]">Sin respuestas registradas.</p>}
              {quizAnswers.map((row, idx) => (
                <div key={idx} className="rounded-lg border p-3 text-sm" style={{ borderColor: 'var(--color-border)' }}>
                  <p className="font-medium">{row.question_text_snapshot}</p>
                  <p className="text-[--color-muted]">{row.answer_text}</p>
                </div>
              ))}
            </div>
          )}
        </section>
      </div>
    </main>
  );
}
