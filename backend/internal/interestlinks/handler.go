package interestlinks

import (
	"errors"
	"net/http"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpjson"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler mantiene un PlayerService (pool usbi_app, GET /interest-links) y
// un AdminService (pool usbi_moderador, CRUD /admin/interest-link-*) —
// mismo patrón que internal/levels desde F3.
type Handler struct {
	player *PlayerService
	admin  *AdminService
}

func NewHandler(player *PlayerService, admin *AdminService) *Handler {
	return &Handler{player: player, admin: admin}
}

func canManageInterestLinks(role domain.UserRole) bool {
	return role == domain.RoleAdmin
}

func claimsFromContext(r *http.Request) (*domain.JWTClaims, bool) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	return claims, ok
}

// ListInterestLinks maneja GET /interest-links: cualquier cuenta
// autenticada (jugador o admin) ve el mismo carrusel agrupado, sin
// distinción de rol — a diferencia de levels.ListLevels no hay contenido
// "no publicado" que ocultarle a un jugador.
func (h *Handler) ListInterestLinks(w http.ResponseWriter, r *http.Request) {
	groups, err := h.player.ListGrouped(r.Context())
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not retrieve interest links")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, groups)
}

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageInterestLinks(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can manage interest link categories")
		return
	}
	resp, err := h.admin.ListCategories(r.Context())
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not list categories")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageInterestLinks(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can manage interest link categories")
		return
	}
	var req CreateCategoryRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}
	resp, err := h.admin.CreateCategory(r.Context(), claims.UserID, req)
	if err != nil {
		writeServiceError(w, r, err, "Could not create category")
		return
	}
	httpproblem.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageInterestLinks(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can manage interest link categories")
		return
	}
	id, ok := parseURLUUID(w, r, "category_id")
	if !ok {
		return
	}
	var req UpdateCategoryRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}
	resp, err := h.admin.UpdateCategory(r.Context(), claims.UserID, id, req)
	if err != nil {
		writeServiceError(w, r, err, "Could not update category")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageInterestLinks(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can manage interest link categories")
		return
	}
	id, ok := parseURLUUID(w, r, "category_id")
	if !ok {
		return
	}
	if err := h.admin.DeleteCategory(r.Context(), claims.UserID, id); err != nil {
		writeServiceError(w, r, err, "Could not delete category")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListLinks(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageInterestLinks(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can manage interest links")
		return
	}
	resp, err := h.admin.ListLinks(r.Context())
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not list interest links")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageInterestLinks(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can manage interest links")
		return
	}
	var req CreateLinkRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}
	resp, err := h.admin.CreateLink(r.Context(), claims.UserID, req)
	if err != nil {
		writeServiceError(w, r, err, "Could not create interest link")
		return
	}
	httpproblem.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) UpdateLink(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageInterestLinks(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can manage interest links")
		return
	}
	id, ok := parseURLUUID(w, r, "link_id")
	if !ok {
		return
	}
	var req UpdateLinkRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}
	resp, err := h.admin.UpdateLink(r.Context(), claims.UserID, id, req)
	if err != nil {
		writeServiceError(w, r, err, "Could not update interest link")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageInterestLinks(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can manage interest links")
		return
	}
	id, ok := parseURLUUID(w, r, "link_id")
	if !ok {
		return
	}
	if err := h.admin.DeleteLink(r.Context(), claims.UserID, id); err != nil {
		writeServiceError(w, r, err, "Could not delete interest link")
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
		httpproblem.WriteProblem(w, r, http.StatusNotFound, "not-found", "Not Found", "Resource not found")
	case errors.Is(err, ErrCategoryHasLinks):
		httpproblem.WriteProblem(w, r, http.StatusConflict, "category-has-links", "Conflict", "Category still has interest links; move or delete them first")
	default:
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", fallback)
	}
}
