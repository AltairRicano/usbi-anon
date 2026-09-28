package legal

import (
	"log/slog"
	"net/http"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetPrivacyNotice maneja GET /api/v1/legal/privacy-notice. Público a
// propósito, sin JWT: el aviso debe poder leerse antes de tener cuenta,
// mismo criterio que ya aplica a /settings.
func (h *Handler) GetPrivacyNotice(w http.ResponseWriter, r *http.Request) {
	httpproblem.WriteJSON(w, http.StatusOK, h.svc.CurrentNotice())
}

// Accept maneja POST /api/v1/legal/accept — el banner de cambio de versión
// (M2.5), no el registro. Requiere sesión: registra qué cuenta aceptó qué
// versión y cuándo.
func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}

	resp, err := h.svc.AcceptCurrent(r.Context(), claims.UserID)
	if err != nil {
		slog.Error("legal accept failed", "error", err)
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
			"Internal Server Error", "Could not record privacy notice acceptance")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}
