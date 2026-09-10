package incidents

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpjson"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
	"github.com/altair/usbi-anon-backend/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler expone el endpoint de administración de incidentes de seguridad. (Relleno)
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// CreateIncident maneja POST /api/v1/admin/security-incidents. (Relleno)
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

// List maneja GET /api/v1/admin/security-incidents.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}

	q := r.URL.Query()
	pageSize := int32(defaultPageSize)
	if ps := q.Get("page_size"); ps != "" {
		n, err := strconv.Atoi(ps)
		if err != nil || n < 1 || n > maxPageSize {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request",
				"Bad Request", "page_size must be between 1 and 50")
			return
		}
		pageSize = int32(n)
	}

	page, err := h.svc.List(r.Context(), *claims, q.Get("cursor"), pageSize)
	if err != nil {
		writeIncidentError(w, r, err, "Could not list security incidents")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, page)
}

// Get maneja GET /api/v1/admin/security-incidents/{incident_id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "incident_id"))
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request",
			"Bad Request", "incident_id must be a valid UUID")
		return
	}
	incident, err := h.svc.Get(r.Context(), *claims, id)
	if err != nil {
		writeIncidentError(w, r, err, "Could not retrieve security incident")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, incident)
}

// Update maneja PATCH /api/v1/admin/security-incidents/{incident_id}.
// Nunca expone un DELETE: no hay ruta, ningún rol tiene el privilegio, y el
// esquema lo prohíbe con un trigger BEFORE DELETE (migración 0006) — un
// incidente de seguridad no se borra desde la aplicación bajo ninguna
// circunstancia.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "incident_id"))
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request",
			"Bad Request", "incident_id must be a valid UUID")
		return
	}
	var req UpdateIncidentRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}
	incident, err := h.svc.Update(r.Context(), *claims, id, req, httputil.ClientIP(r), r.UserAgent())
	if err != nil {
		writeIncidentError(w, r, err, "Could not update security incident")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, incident)
}

func writeIncidentError(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	switch {
	case errors.Is(err, ErrForbidden):
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden",
			"Forbidden", "Only admins can manage security incidents")
	case errors.Is(err, ErrValidation):
		httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error",
			"Validation Error", "Invalid security incident payload")
	case errors.Is(err, ErrNotFound):
		httpproblem.WriteProblem(w, r, http.StatusNotFound, "not-found",
			"Not Found", "Security incident not found")
	default:
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
			"Internal Server Error", fallback)
	}
}
