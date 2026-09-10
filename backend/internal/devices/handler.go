package devices

import (
	"errors"
	"net/http"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpjson"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized", "Unauthorized", "Missing JWT claims in context")
		return
	}

	var req RegisterDeviceRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	resp, created, err := h.svc.RegisterDevice(r.Context(), claims.UserID, req)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error", "Validation Error", err.Error())
			return
		}
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not register device")
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	httpproblem.WriteJSON(w, status, resp)
}

func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized", "Unauthorized", "Missing JWT claims in context")
		return
	}
	resp, err := h.svc.ListDevices(r.Context(), claims.UserID)
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not list devices")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

// RevokeDevice maneja DELETE /devices/{device_id}: revocación lógica, no un
// borrado físico — ver el comentario de Service.RevokeDevice.
func (h *Handler) RevokeDevice(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized", "Unauthorized", "Missing JWT claims in context")
		return
	}
	deviceID, err := uuid.Parse(chi.URLParam(r, "device_id"))
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request", "Bad Request", "device_id must be a valid UUID")
		return
	}
	if err := h.svc.RevokeDevice(r.Context(), claims.UserID, deviceID); err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error", "Validation Error", err.Error())
		case errors.Is(err, ErrNotFound):
			httpproblem.WriteProblem(w, r, http.StatusNotFound, "not-found", "Not Found", "Device not found")
		default:
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not revoke device")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
