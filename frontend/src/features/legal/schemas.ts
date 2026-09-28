import { z } from 'zod';

// Espeja legal.PrivacyNoticeResponse en Go (backend/internal/legal/dto.go).
export const PrivacyNoticeSectionSchema = z.object({
  heading: z.string(),
  paragraphs: z.array(z.string()),
});

export const PrivacyNoticeResponseSchema = z.object({
  version: z.string(),
  effective_date: z.string(),
  simplified: z.array(PrivacyNoticeSectionSchema),
  full: z.array(PrivacyNoticeSectionSchema),
  checksum: z.string(),
});
export type PrivacyNoticeSection = z.infer<typeof PrivacyNoticeSectionSchema>;
export type PrivacyNoticeResponse = z.infer<typeof PrivacyNoticeResponseSchema>;

// Espeja auth.MeResponse en Go. Solo los campos de privacidad — user_id/role
// existen en la respuesta pero esta feature no los necesita.
export const MyPrivacyStatusSchema = z.object({
  privacy_notice_version: z.string(),
  current_privacy_notice_version: z.string(),
});
export type MyPrivacyStatus = z.infer<typeof MyPrivacyStatusSchema>;
