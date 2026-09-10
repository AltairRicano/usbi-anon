import { z } from 'zod';

// Espeja devices.DeviceResponse/DevicesResponse en Go
// (backend/internal/devices/dto.go).
export const DeviceSchema = z.object({
  id: z.string().uuid(),
  user_id: z.string().uuid(),
  device_kind: z.enum(['movil', 'tablet', 'laptop', 'escritorio', 'otro']),
  platform: z.enum(['web', 'tauri']),
  registered_at: z.string(),
  last_seen_at: z.string(),
  wipe_local_data: z.boolean(),
  revoked_at: z.string().optional(),
});
export type Device = z.infer<typeof DeviceSchema>;

export const DevicesResponseSchema = z.object({
  items: z.array(DeviceSchema),
});

// Espeja sync.SyncEventSummary/SyncHistoryPage en Go
// (backend/internal/sync/history.go).
export const SyncEventSchema = z.object({
  id: z.string().uuid(),
  device_id: z.string().uuid(),
  status: z.string(),
  hmac_valid: z.boolean(),
  received_at: z.string(),
  processed_at: z.string().optional(),
  rejection_reason: z.string().optional(),
});
export type SyncEvent = z.infer<typeof SyncEventSchema>;

export const SyncHistoryPageSchema = z.object({
  items: z.array(SyncEventSchema),
  next_cursor: z.string().optional(),
});
