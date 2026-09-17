// Package legal expone HTTP el aviso de privacidad vigente (M2 del plan de
// maduración) y registra la aceptación de cambios de versión por una cuenta
// ya logueada (D-06: informativo, no bloqueante — M2.5). El texto en sí vive
// en el paquete no-internal legaltext (github.com/.../legal), empotrado con
// go:embed; este paquete solo lo sirve y sella su aceptación.
package legal

import (
	"context"
	"time"

	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/altair/usbi-anon-backend/internal/repository"
	legaltext "github.com/altair/usbi-anon-backend/legal"
	"github.com/google/uuid"
)

type Service struct {
	repo       *repository.Queries
	hmacSecret []byte
}

func NewService(repo *repository.Queries, hmacSecret []byte) *Service {
	return &Service{repo: repo, hmacSecret: hmacSecret}
}

// CurrentNotice devuelve el aviso vigente. Público a propósito (igual que
// /settings): debe poder leerse antes de tener cuenta.
func (s *Service) CurrentNotice() PrivacyNoticeResponse {
	n := legaltext.Current()
	return PrivacyNoticeResponse{
		Version:       n.Version,
		EffectiveDate: n.EffectiveDate,
		Simplified:    toSectionResponses(n.Simplified),
		Full:          toSectionResponses(n.Full),
		Checksum:      n.Checksum,
	}
}

func toSectionResponses(in []legaltext.Section) []SectionResponse {
	out := make([]SectionResponse, len(in))
	for i, sec := range in {
		out[i] = SectionResponse{Heading: sec.Heading, Paragraphs: sec.Paragraphs}
	}
	return out
}

// AcceptCurrent registra que userID aceptó la versión vigente del aviso —
// el flujo del banner de cambio de versión (M2.5), no el registro: ahí el
// sello ya se calcula en auth.Service.RegisterConfirm con el mismo
// legaltext.SealPayload.
func (s *Service) AcceptCurrent(ctx context.Context, userID uuid.UUID) (AcceptResponse, error) {
	n := legaltext.Current()
	acceptedAt := time.Now().UTC()
	hash := crypto.GenerateHMAC(legaltext.SealPayload(userID, acceptedAt), s.hmacSecret)

	if err := s.repo.UpdatePrivacyAcceptance(ctx, repository.UpdatePrivacyAcceptanceParams{
		ID:         userID,
		Version:    n.Version,
		AcceptedAt: acceptedAt,
		Hash:       hash,
	}); err != nil {
		return AcceptResponse{}, err
	}

	return AcceptResponse{Version: n.Version, AcceptedAt: acceptedAt}, nil
}
