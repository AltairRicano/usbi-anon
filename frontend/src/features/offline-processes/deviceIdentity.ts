import { apiClient } from '../../shared/apiClient';
import { DeviceSchema } from './schemas';

const STORAGE_KEY = 'usbi_device_id';

export function getStoredDeviceID(): string | null {
  return localStorage.getItem(STORAGE_KEY);
}

export function storeDeviceID(id: string): void {
  localStorage.setItem(STORAGE_KEY, id);
}

// Heurística de device_kind (C1, estado_proyecto.md 2026-09-10): el CHECK
// de devices.device_kind solo acepta estos cinco valores — "otro" es el
// escape razonable cuando la detección por user-agent/ancho de pantalla no
// concluye. platform siempre "web": el cliente Tauri, cuando exista, usará
// su propia heurística nativa.
export function guessDeviceKind(): 'movil' | 'tablet' | 'laptop' | 'escritorio' | 'otro' {
  const ua = navigator.userAgent.toLowerCase();
  const width = window.innerWidth;
  if (/mobile|android|iphone/.test(ua)) return 'movil';
  if (/ipad|tablet/.test(ua)) return 'tablet';
  if (width >= 1280) return 'escritorio';
  if (width >= 768) return 'laptop';
  return 'otro';
}

// registerCurrentDevice implementa C1 (estado_proyecto.md 2026-09-10): se
// llama justo después de un login exitoso, nunca detrás de un botón manual.
// Es un upsert en el backend — con device_id ya guardado, solo toca
// last_seen_at; sin él (o si ya no es válido), crea uno nuevo y lo persiste
// aquí para el siguiente login. Falla en silencio a propósito: que
// "Procesos Offline" no reciba este dispositivo no debe bloquear el login.
export async function registerCurrentDevice(): Promise<void> {
  try {
    const storedID = getStoredDeviceID();
    const resp = await apiClient.post('/devices', {
      device_id: storedID ?? undefined,
      device_kind: guessDeviceKind(),
      platform: 'web',
    });
    const device = DeviceSchema.parse(resp.data);
    storeDeviceID(device.id);
  } catch (err) {
    console.warn('[registerCurrentDevice] No se pudo registrar el dispositivo:', err);
  }
}
