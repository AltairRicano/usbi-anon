import { useState, type FormEvent } from 'react';
import { Button } from '../../shared/components/ui/Button';
import { HomeButton } from '../../shared/components/ui/HomeButton';
import { Input } from '../../shared/components/ui/Input';
import { apiClient } from '../../shared/apiClient';
import { errorMessage } from '../../shared/errorMessage';
import {
  AuditLogPageSchema,
  SecurityIncidentsPageSchema,
  SeveritySchema,
  SEVERITIES,
  type AuditLogEntry,
  type SecurityIncident,
} from './schemas';

type Tab = 'audit' | 'incidents';

interface IncidentForm {
  severity: string;
  affected_scope: string;
  description: string;
  containment_actions: string;
  reported_to_cutai: boolean;
}

const EMPTY_INCIDENT_FORM: IncidentForm = {
  severity: 'low',
  affected_scope: '',
  description: '',
  containment_actions: '',
  reported_to_cutai: false,
};

export default function AdminSecurityPage() {
  const [tab, setTab] = useState<Tab>('audit');

  // ── Bitácora (GET /admin/audit-log) ──────────────────────────────────
  const [entries, setEntries] = useState<AuditLogEntry[]>([]);
  const [auditCursor, setAuditCursor] = useState<string | undefined>(undefined);
  const [auditAction, setAuditAction] = useState('');
  const [auditEntityType, setAuditEntityType] = useState('');
  const [auditError, setAuditError] = useState<string | null>(null);
  const [auditLoading, setAuditLoading] = useState(false);

  async function loadAuditLog(reset: boolean) {
    setAuditLoading(true);
    setAuditError(null);
    try {
      const params: Record<string, string> = {};
      if (auditAction.trim()) params.action = auditAction.trim();
      if (auditEntityType.trim()) params.entity_type = auditEntityType.trim();
      if (!reset && auditCursor) params.cursor = auditCursor;
      const resp = await apiClient.get('/admin/audit-log', { params });
      const page = AuditLogPageSchema.parse(resp.data);
      setEntries(reset ? page.items : [...entries, ...page.items]);
      setAuditCursor(page.next_cursor);
    } catch (err) {
      setAuditError(errorMessage(err, 'No se pudo cargar la bitácora.'));
    } finally {
      setAuditLoading(false);
    }
  }

  // ── Incidentes de seguridad (GET/POST/PATCH /admin/security-incidents) ─
  const [incidents, setIncidents] = useState<SecurityIncident[]>([]);
  const [incidentsCursor, setIncidentsCursor] = useState<string | undefined>(undefined);
  const [incidentsError, setIncidentsError] = useState<string | null>(null);
  const [incidentsLoading, setIncidentsLoading] = useState(false);
  const [incidentsLoaded, setIncidentsLoaded] = useState(false);

  const [creating, setCreating] = useState(false);
  const [newIncident, setNewIncident] = useState<IncidentForm>(EMPTY_INCIDENT_FORM);

  const [editingID, setEditingID] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<IncidentForm | null>(null);

  async function loadIncidents(reset: boolean) {
    setIncidentsLoading(true);
    setIncidentsError(null);
    try {
      const params: Record<string, string> = {};
      if (!reset && incidentsCursor) params.cursor = incidentsCursor;
      const resp = await apiClient.get('/admin/security-incidents', { params });
      const page = SecurityIncidentsPageSchema.parse(resp.data);
      setIncidents(reset ? page.items : [...incidents, ...page.items]);
      setIncidentsCursor(page.next_cursor);
      setIncidentsLoaded(true);
    } catch (err) {
      setIncidentsError(errorMessage(err, 'No se pudieron cargar los incidentes.'));
    } finally {
      setIncidentsLoading(false);
    }
  }

  function openIncidentsTab() {
    setTab('incidents');
    if (!incidentsLoaded) void loadIncidents(true);
  }

  async function createIncident(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setCreating(true);
    setIncidentsError(null);
    try {
      const parsedSeverity = SeveritySchema.safeParse(newIncident.severity);
      if (!parsedSeverity.success) {
        setIncidentsError('Severidad inválida.');
        return;
      }
      await apiClient.post('/admin/security-incidents', newIncident);
      setNewIncident(EMPTY_INCIDENT_FORM);
      await loadIncidents(true);
    } catch (err) {
      setIncidentsError(errorMessage(err, 'No se pudo registrar el incidente.'));
    } finally {
      setCreating(false);
    }
  }

  function startEdit(incident: SecurityIncident) {
    setEditingID(incident.id);
    setEditForm({
      severity: incident.severity,
      affected_scope: incident.affected_scope,
      description: incident.description,
      containment_actions: incident.containment_actions,
      reported_to_cutai: incident.reported_to_cutai,
    });
  }

  async function saveEdit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!editingID || !editForm) return;
    setIncidentsError(null);
    try {
      await apiClient.patch(`/admin/security-incidents/${editingID}`, editForm);
      setEditingID(null);
      setEditForm(null);
      await loadIncidents(true);
    } catch (err) {
      setIncidentsError(errorMessage(err, 'No se pudo actualizar el incidente.'));
    }
  }

  return (
    <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
      <div className="mx-auto max-w-4xl space-y-6">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">Seguridad y bitácora</h1>
            <p className="text-sm text-[--color-muted]">Registro de auditoría y bitácora de incidentes de seguridad.</p>
          </div>
          <HomeButton />
        </header>

        <div className="flex bg-[--color-card] rounded-full p-1 border border-[--color-border] w-max shadow-inner">
          <button
            onClick={() => setTab('audit')}
            className={`px-6 py-2 rounded-full font-bold transition-all duration-200 ${tab === 'audit' ? 'bg-[--color-primary] text-[--color-primary-foreground]' : 'text-[--color-muted]'}`}
          >
            Bitácora
          </button>
          <button
            onClick={openIncidentsTab}
            className={`px-6 py-2 rounded-full font-bold transition-all duration-200 ${tab === 'incidents' ? 'bg-[--color-primary] text-[--color-primary-foreground]' : 'text-[--color-muted]'}`}
          >
            Incidentes
          </button>
        </div>

        {tab === 'audit' && (
          <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border] space-y-4">
            <form
              onSubmit={(e) => {
                e.preventDefault();
                void loadAuditLog(true);
              }}
              className="grid gap-3 md:grid-cols-[1fr_1fr_auto] md:items-end"
            >
              <Input id="audit-action" label="Acción (ej. device.revoke)" value={auditAction} onChange={(e) => setAuditAction(e.currentTarget.value)} />
              <Input id="audit-entity-type" label="Tipo de entidad (ej. device)" value={auditEntityType} onChange={(e) => setAuditEntityType(e.currentTarget.value)} />
              <Button type="submit" disabled={auditLoading}>Filtrar</Button>
            </form>

            {auditError && (
              <p role="alert" className="rounded-lg border p-3 text-sm" style={{ borderColor: 'var(--color-error)', color: 'var(--color-error)' }}>
                {auditError}
              </p>
            )}

            <div className="divide-y divide-[--color-border]">
              {entries.map((entry) => (
                <div key={entry.id} className="py-3 text-sm">
                  <p className="font-medium">
                    {entry.action} <span className="text-[--color-muted]">· {entry.entity_type}</span>
                  </p>
                  <p className="text-xs text-[--color-muted]">
                    {new Date(entry.created_at).toLocaleString()} · actor {entry.actor_account_id ?? '—'} · {entry.ip_address}
                  </p>
                </div>
              ))}
              {entries.length === 0 && !auditLoading && <p className="py-4 text-sm text-[--color-muted]">Sin entradas.</p>}
            </div>

            <div className="flex justify-center gap-3">
              {entries.length === 0 && (
                <Button type="button" variant="outline" onClick={() => void loadAuditLog(true)} disabled={auditLoading}>
                  {auditLoading ? 'Cargando…' : 'Cargar bitácora'}
                </Button>
              )}
              {auditCursor && (
                <Button type="button" variant="outline" onClick={() => void loadAuditLog(false)} disabled={auditLoading}>
                  {auditLoading ? 'Cargando…' : 'Cargar más'}
                </Button>
              )}
            </div>
          </section>
        )}

        {tab === 'incidents' && (
          <div className="space-y-6">
            <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border]">
              <h2 className="mb-4 text-xl font-semibold">Registrar incidente</h2>
              <form onSubmit={createIncident} className="grid gap-3 md:grid-cols-2">
                <div className="flex flex-col gap-1">
                  <label htmlFor="new-incident-severity" className="text-sm font-medium">Severidad</label>
                  <select
                    id="new-incident-severity"
                    value={newIncident.severity}
                    onChange={(e) => setNewIncident({ ...newIncident, severity: e.currentTarget.value })}
                    className="min-h-[44px] rounded-lg border px-4 py-2 text-base border-[--color-border] bg-[--color-background]"
                  >
                    {SEVERITIES.map((s) => (
                      <option key={s} value={s}>{s}</option>
                    ))}
                  </select>
                </div>
                <label className="flex items-center gap-2 text-sm mt-6">
                  <input
                    type="checkbox"
                    checked={newIncident.reported_to_cutai}
                    onChange={(e) => setNewIncident({ ...newIncident, reported_to_cutai: e.currentTarget.checked })}
                    className="h-5 w-5"
                  />
                  Reportado al CUTAI
                </label>
                <Input
                  id="new-incident-scope"
                  label="Alcance afectado"
                  value={newIncident.affected_scope}
                  onChange={(e) => setNewIncident({ ...newIncident, affected_scope: e.currentTarget.value })}
                  required
                  className="md:col-span-2"
                />
                <Input
                  id="new-incident-description"
                  label="Descripción"
                  value={newIncident.description}
                  onChange={(e) => setNewIncident({ ...newIncident, description: e.currentTarget.value })}
                  required
                  className="md:col-span-2"
                />
                <Input
                  id="new-incident-containment"
                  label="Acciones de contención"
                  value={newIncident.containment_actions}
                  onChange={(e) => setNewIncident({ ...newIncident, containment_actions: e.currentTarget.value })}
                  required
                  className="md:col-span-2"
                />
                <Button type="submit" disabled={creating} className="md:col-span-2">
                  {creating ? 'Registrando…' : 'Registrar incidente'}
                </Button>
              </form>
            </section>

            <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border] space-y-4">
              <h2 className="text-xl font-semibold">Incidentes registrados</h2>

              {incidentsError && (
                <p role="alert" className="rounded-lg border p-3 text-sm" style={{ borderColor: 'var(--color-error)', color: 'var(--color-error)' }}>
                  {incidentsError}
                </p>
              )}

              <div className="divide-y divide-[--color-border]">
                {incidents.map((inc) => (
                  <div key={inc.id} className="py-3">
                    {editingID === inc.id && editForm ? (
                      <form onSubmit={saveEdit} className="grid gap-3 rounded-lg border border-[--color-border] p-4">
                        <div className="flex flex-col gap-1">
                          <label htmlFor={`edit-severity-${inc.id}`} className="text-sm font-medium">Severidad</label>
                          <select
                            id={`edit-severity-${inc.id}`}
                            value={editForm.severity}
                            onChange={(e) => setEditForm({ ...editForm, severity: e.currentTarget.value })}
                            className="min-h-[44px] rounded-lg border px-4 py-2 text-base border-[--color-border] bg-[--color-background]"
                          >
                            {SEVERITIES.map((s) => (
                              <option key={s} value={s}>{s}</option>
                            ))}
                          </select>
                        </div>
                        <label className="flex items-center gap-2 text-sm">
                          <input
                            type="checkbox"
                            checked={editForm.reported_to_cutai}
                            onChange={(e) => setEditForm({ ...editForm, reported_to_cutai: e.currentTarget.checked })}
                            className="h-5 w-5"
                          />
                          Reportado al CUTAI
                        </label>
                        <Input id={`edit-scope-${inc.id}`} label="Alcance afectado" value={editForm.affected_scope} onChange={(e) => setEditForm({ ...editForm, affected_scope: e.currentTarget.value })} required />
                        <Input id={`edit-description-${inc.id}`} label="Descripción" value={editForm.description} onChange={(e) => setEditForm({ ...editForm, description: e.currentTarget.value })} required />
                        <Input id={`edit-containment-${inc.id}`} label="Acciones de contención" value={editForm.containment_actions} onChange={(e) => setEditForm({ ...editForm, containment_actions: e.currentTarget.value })} required />
                        <div className="flex gap-2">
                          <Button type="submit" size="sm">Guardar</Button>
                          <Button type="button" size="sm" variant="outline" onClick={() => { setEditingID(null); setEditForm(null); }}>Cancelar</Button>
                        </div>
                      </form>
                    ) : (
                      <div className="flex flex-wrap items-start justify-between gap-3">
                        <div>
                          <p className="font-medium">
                            {inc.severity.toUpperCase()} · {inc.affected_scope}{' '}
                            {!inc.evidence_valid && (
                              <span className="ml-2 rounded-full px-2 py-0.5 text-xs" style={{ backgroundColor: 'var(--color-error)', color: 'white' }}>
                                Evidencia alterada
                              </span>
                            )}
                          </p>
                          <p className="text-sm text-[--color-muted]">{inc.description}</p>
                          <p className="text-xs text-[--color-muted]">
                            Detectado {new Date(inc.detected_at).toLocaleString()}
                            {inc.resolved_at ? ` · resuelto ${new Date(inc.resolved_at).toLocaleString()}` : ' · sin resolver'}
                          </p>
                        </div>
                        <Button size="sm" variant="outline" onClick={() => startEdit(inc)}>Editar</Button>
                      </div>
                    )}
                  </div>
                ))}
                {incidents.length === 0 && !incidentsLoading && <p className="py-4 text-sm text-[--color-muted]">Sin incidentes registrados.</p>}
              </div>

              {incidentsCursor && (
                <div className="flex justify-center">
                  <Button type="button" variant="outline" onClick={() => void loadIncidents(false)} disabled={incidentsLoading}>
                    {incidentsLoading ? 'Cargando…' : 'Cargar más'}
                  </Button>
                </div>
              )}
            </section>
          </div>
        )}
      </div>
    </main>
  );
}
