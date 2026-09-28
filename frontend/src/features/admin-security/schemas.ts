import { z } from 'zod';

// Espeja auditlog.EntryResponse/Page en Go (backend/internal/auditlog/dto.go).
// actor_account_id viaja como UUID crudo a propósito: usbi_moderador no
// puede resolverlo a un nickname (ver el comentario del DTO Go).
export const AuditLogEntrySchema = z.object({
  id: z.string().uuid(),
  actor_account_id: z.string().uuid().optional(),
  action: z.string(),
  entity_type: z.string(),
  entity_id: z.string().uuid().optional(),
  before_state: z.unknown().optional(),
  after_state: z.unknown().optional(),
  ip_address: z.string(),
  user_agent: z.string(),
  created_at: z.string(),
});
export type AuditLogEntry = z.infer<typeof AuditLogEntrySchema>;

export const AuditLogPageSchema = z.object({
  items: z.array(AuditLogEntrySchema),
  next_cursor: z.string().optional(),
});

// Espeja incidents.IncidentResponse/IncidentsPage/UpdateIncidentRequest en
// Go (backend/internal/incidents/admin_read.go, service.go).
export const SEVERITIES = ['low', 'medium', 'high', 'critical'] as const;
export const SeveritySchema = z.enum(SEVERITIES);

export const SecurityIncidentSchema = z.object({
  id: z.string().uuid(),
  detected_at: z.string(),
  reported_at: z.string().optional(),
  severity: SeveritySchema,
  affected_scope: z.string(),
  description: z.string(),
  containment_actions: z.string(),
  resolved_at: z.string().optional(),
  reported_to_cutai: z.boolean(),
  notified_to_cutai_at: z.string().optional(),
  notified_to_subjects_at: z.string().optional(),
  evidence_valid: z.boolean(),
});
export type SecurityIncident = z.infer<typeof SecurityIncidentSchema>;

export const SecurityIncidentsPageSchema = z.object({
  items: z.array(SecurityIncidentSchema),
  next_cursor: z.string().optional(),
});

export const CreateIncidentResponseSchema = z.object({
  id: z.string().uuid(),
  severity: SeveritySchema,
  message: z.string(),
});
