package badges

import (
	"errors"
	"net/http"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpjson"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler expone el CRUD administrativo del catálogo de insignias. Solo
// AdminService — la lectura del jugador se gestiona en internal/levels.
type Handler struct {
	admin *AdminService
}

func NewHandler(admin *AdminService) *Handler {
	return &Handler{admin: admin}
}

func canManageBadges(role domain.UserRole) bool {
	return role == domain.RoleAdmin
}

func claimsFromContext(r *http.Request) (*domain.JWTClaims, bool) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	return claims, ok
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageBadges(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can manage the badge catalog")
		return
	}
	resp, err := h.admin.List(r.Context())
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not list badges")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageBadges(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can manage the badge catalog")
		return
	}
	var req CreateBadgeRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}
	resp, err := h.admin.Create(r.Context(), claims.UserID, req)
	if err != nil {
		writeServiceError(w, r, err, "Could not create badge")
		return
	}
	httpproblem.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageBadges(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can manage the badge catalog")
		return
	}
	id, ok := parseURLUUID(w, r, "badge_id")
	if !ok {
		return
	}
	var req UpdateBadgeRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}
	resp, err := h.admin.Update(r.Context(), claims.UserID, id, req)
	if err != nil {
		writeServiceError(w, r, err, "Could not update badge")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageBadges(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can manage the badge catalog")
		return
	}
	id, ok := parseURLUUID(w, r, "badge_id")
	if !ok {
		return
	}
	if err := h.admin.Delete(r.Context(), claims.UserID, id); err != nil {
		writeServiceError(w, r, err, "Could not delete badge")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseURLUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request", "Bad Request", key+" must be a valid UUID")
		return uuid.Nil, false
	}
	return parsed, true
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	switch {
	case errors.Is(err, ErrValidation):
		httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error", "Validation Error", err.Error())
	case errors.Is(err, ErrNotFound):
		httpproblem.WriteProblem(w, r, http.StatusNotFound, "not-found", "Not Found", "Badge not found")
	case errors.Is(err, ErrBadgeHasHolder):
		httpproblem.WriteProblem(w, r, http.StatusConflict, "badge-has-holder", "Conflict", "Badge has already been earned by at least one account and cannot be deleted")
	default:
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", fallback)
	}
}
