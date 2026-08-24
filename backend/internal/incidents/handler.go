package incidents

import (
	"errors"
	"net/http"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpjson"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
	"github.com/altair/usbi-anon-backend/internal/httputil"
)

// Handler exposes the security-incident admin endpoint.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// CreateIncident handles POST /api/v1/admin/security-incidents.
func (h *Handler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}

	var req CreateIncidentRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	resp, err := h.svc.CreateIncident(r.Context(), *claims, req, httputil.ClientIP(r), r.UserAgent())
	if err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden",
				"Forbidden", "Only admins or directors can record security incidents")
		case errors.Is(err, ErrValidation):
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error",
				"Validation Error", "Invalid security incident payload")
		default:
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
				"Internal Server Error", "Could not record security incident")
		}
		return
	}

	httpproblem.WriteJSON(w, http.StatusCreated, resp)
}
