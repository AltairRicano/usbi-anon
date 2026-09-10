import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Button } from '../../shared/components/ui/Button';
import { apiClient } from '../../shared/apiClient';
import { errorMessage } from '../../shared/errorMessage';
import { DevicesResponseSchema, SyncHistoryPageSchema, type Device, type SyncEvent } from './schemas';

const ALL_DEVICES = '__all__';

const DEVICE_KIND_LABEL: Record<Device['device_kind'], string> = {
  movil: 'Móvil',
  tablet: 'Tablet',
  laptop: 'Laptop',
  escritorio: 'Escritorio',
  otro: 'Otro',
};

const STATUS_LABEL: Record<string, string> = {
  accepted: 'Aceptado',
  rejected: 'Rechazado',
};

// Copy aprobado por el usuario (bloque C, estado_proyecto.md 2026-09-10).
const REVOKE_WARNING =
  'Si eliminas este dispositivo, dejará de sincronizar tu progreso sin conexión. ' +
  'Lo que ya subió se conserva, pero lo que juegues sin conexión en él no se guardará ' +
  'hasta que vuelvas a iniciar sesión desde ese mismo dispositivo.';

export default function OfflineProcessesPage() {
  const [devices, setDevices] = useState<Device[]>([]);
  const [devicesError, setDevicesError] = useState<string | null>(null);
  const [selectedDeviceID, setSelectedDeviceID] = useState<string>(ALL_DEVICES);
  const [confirmingRevoke, setConfirmingRevoke] = useState<Device | null>(null);
  const [revoking, setRevoking] = useState(false);

  const [events, setEvents] = useState<SyncEvent[]>([]);
  const [eventsCursor, setEventsCursor] = useState<string | undefined>(undefined);
  const [eventsError, setEventsError] = useState<string | null>(null);
  const [eventsLoading, setEventsLoading] = useState(false);

  async function loadDevices() {
    setDevicesError(null);
    try {
      const resp = await apiClient.get('/devices');
      setDevices(DevicesResponseSchema.parse(resp.data).items);
    } catch (err) {
      setDevicesError(errorMessage(err, 'No se pudieron cargar tus dispositivos.'));
    }
  }

  useEffect(() => {
    void loadDevices();
  }, []);

  async function loadEvents(reset: boolean) {
    setEventsLoading(true);
    setEventsError(null);
    try {
      const params: Record<string, string> = {};
      if (selectedDeviceID !== ALL_DEVICES) params.device_id = selectedDeviceID;
      if (!reset && eventsCursor) params.cursor = eventsCursor;
      const resp = await apiClient.get('/sync/events', { params });
      const page = SyncHistoryPageSchema.parse(resp.data);
      setEvents(reset ? page.items : [...events, ...page.items]);
      setEventsCursor(page.next_cursor);
    } catch (err) {
      setEventsError(errorMessage(err, 'No se pudo cargar el historial de sincronización.'));
    } finally {
      setEventsLoading(false);
    }
  }

  useEffect(() => {
    void loadEvents(true);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedDeviceID]);

  async function confirmRevoke() {
    if (!confirmingRevoke) return;
    setRevoking(true);
    try {
      await apiClient.delete(`/devices/${confirmingRevoke.id}`);
      setConfirmingRevoke(null);
      if (selectedDeviceID === confirmingRevoke.id) setSelectedDeviceID(ALL_DEVICES);
      await loadDevices();
    } catch (err) {
      setDevicesError(errorMessage(err, 'No se pudo eliminar el dispositivo.'));
    } finally {
      setRevoking(false);
    }
  }

  return (
    <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
      {confirmingRevoke && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
          <div className="max-w-md rounded-2xl bg-[--color-card] p-6 shadow-2xl border border-[--color-border] space-y-4">
            <h2 className="text-xl font-bold" style={{ color: 'var(--color-error)' }}>¿Eliminar este dispositivo?</h2>
            <p className="text-sm">{REVOKE_WARNING}</p>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setConfirmingRevoke(null)} disabled={revoking}>Cancelar</Button>
              <Button variant="danger" onClick={() => void confirmRevoke()} disabled={revoking}>
                {revoking ? 'Eliminando…' : 'Eliminar dispositivo'}
              </Button>
            </div>
          </div>
        </div>
      )}

      <div className="mx-auto max-w-4xl space-y-6">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">Procesos Offline</h1>
            <p className="text-sm text-[--color-muted]">Tus dispositivos y el historial de sincronización de cada uno.</p>
          </div>
          <Button variant="outline" size="sm">
            <Link to="/">Dashboard</Link>
          </Button>
        </header>

        {devicesError && (
          <p role="alert" className="rounded-lg border p-3 text-sm" style={{ borderColor: 'var(--color-error)', color: 'var(--color-error)' }}>
            {devicesError}
          </p>
        )}

        <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border]">
          <h2 className="mb-4 text-xl font-semibold">Dispositivos</h2>
          <div className="divide-y divide-[--color-border]">
            {devices.map((d) => (
              <div key={d.id} className="flex flex-wrap items-center justify-between gap-3 py-3">
                <div>
                  <p className="font-medium">{DEVICE_KIND_LABEL[d.device_kind]} · {d.platform}</p>
                  <p className="text-xs text-[--color-muted]">
                    Última actividad {new Date(d.last_seen_at).toLocaleString()} · registrado {new Date(d.registered_at).toLocaleDateString()}
                  </p>
                </div>
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" onClick={() => setSelectedDeviceID(d.id)}>
                    Ver historial
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    className="border-[--color-error] text-[--color-error] hover:bg-[--color-error] hover:text-white"
                    onClick={() => setConfirmingRevoke(d)}
                  >
                    Eliminar
                  </Button>
                </div>
              </div>
            ))}
            {devices.length === 0 && <p className="py-4 text-sm text-[--color-muted]">Aún no tienes dispositivos registrados.</p>}
          </div>
        </section>

        <section className="rounded-2xl bg-[--color-card] p-5 shadow-lg border border-[--color-border] space-y-4">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <h2 className="text-xl font-semibold">Historial de sincronización</h2>
            <div className="flex flex-col gap-1">
              <label htmlFor="device-filter" className="text-sm font-medium">Dispositivo</label>
              <select
                id="device-filter"
                value={selectedDeviceID}
                onChange={(e) => setSelectedDeviceID(e.currentTarget.value)}
                className="min-h-[44px] rounded-lg border px-4 py-2 text-base border-[--color-border] bg-[--color-background]"
              >
                <option value={ALL_DEVICES}>Todos los dispositivos</option>
                {devices.map((d) => (
                  <option key={d.id} value={d.id}>{DEVICE_KIND_LABEL[d.device_kind]} · {d.platform}</option>
                ))}
              </select>
            </div>
          </div>

          {eventsError && (
            <p role="alert" className="rounded-lg border p-3 text-sm" style={{ borderColor: 'var(--color-error)', color: 'var(--color-error)' }}>
              {eventsError}
            </p>
          )}

          <div className="divide-y divide-[--color-border]">
            {events.map((ev) => (
              <div key={ev.id} className="py-3 text-sm">
                <p className="font-medium">
                  {STATUS_LABEL[ev.status] ?? ev.status}
                  {!ev.hmac_valid && (
                    <span className="ml-2 rounded-full px-2 py-0.5 text-xs" style={{ backgroundColor: 'var(--color-error)', color: 'white' }}>
                      HMAC inválido
                    </span>
                  )}
                </p>
                <p className="text-xs text-[--color-muted]">
                  {new Date(ev.received_at).toLocaleString()}
                  {ev.processed_at && ` · procesado ${new Date(ev.processed_at).toLocaleString()}`}
                  {ev.rejection_reason && ` · ${ev.rejection_reason}`}
                </p>
              </div>
            ))}
            {events.length === 0 && !eventsLoading && <p className="py-4 text-sm text-[--color-muted]">Sin eventos de sincronización todavía.</p>}
          </div>

          {eventsCursor && (
            <div className="flex justify-center">
              <Button type="button" variant="outline" onClick={() => void loadEvents(false)} disabled={eventsLoading}>
                {eventsLoading ? 'Cargando…' : 'Cargar más'}
              </Button>
            </div>
          )}
        </section>
      </div>
    </main>
  );
}
