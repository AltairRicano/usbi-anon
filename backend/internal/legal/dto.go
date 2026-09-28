package legal

import "time"

// SectionResponse es una sección de aviso en la respuesta HTTP — misma forma
// que legaltext.Section, repetida aquí para no filtrar el tipo de un paquete
// no-internal directo al DTO público (mismo patrón que domain.User en auth).
type SectionResponse struct {
	Heading    string   `json:"heading"`
	Paragraphs []string `json:"paragraphs"`
}

// PrivacyNoticeResponse es la respuesta de GET /api/v1/legal/privacy-notice.
type PrivacyNoticeResponse struct {
	Version       string            `json:"version"`
	EffectiveDate string            `json:"effective_date"`
	Simplified    []SectionResponse `json:"simplified"`
	Full          []SectionResponse `json:"full"`
	Checksum      string            `json:"checksum"`
}

// AcceptResponse es la respuesta de POST /api/v1/legal/accept.
type AcceptResponse struct {
	Version    string    `json:"version"`
	AcceptedAt time.Time `json:"accepted_at"`
}
