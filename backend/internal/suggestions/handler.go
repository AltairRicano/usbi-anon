package suggestions

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpjson"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler mantiene un PlayerService (pool usbi_app, POST /suggestions) y un
// AdminService (pool usbi_moderador, GET/DELETE /admin/suggestions).
type Handler struct {
	player *PlayerService
	admin  *AdminService
}

func NewHandler(player *PlayerService, admin *AdminService) *Handler {
	return &Handler{player: player, admin: admin}
}

func canManageSuggestions(role domain.UserRole) bool {
	return role == domain.RoleAdmin
}

func claimsFromContext(r *http.Request) (*domain.JWTClaims, bool) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	return claims, ok
}

// Submit maneja POST /suggestions. Cualquier cuenta autenticada puede
// mandar una sugerencia, incluida una cuenta admin (no hay razón de negocio
// para excluirla) — no hay guard de rol aquí, solo autenticación (ya
// impuesta por el middleware del router).
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized", "Unauthorized", "Missing JWT claims in context")
		return
	}
	var req SubmitRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}
	resp, err := h.player.Submit(r.Context(), claims.UserID, req)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error", "Validation Error", "Invalid suggestion payload")
		} else {
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not submit suggestion")
		}
		return
	}
	httpproblem.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageSuggestions(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can view the suggestion box")
		return
	}

	q := r.URL.Query()
	cursor := uuid.Nil
	if c := q.Get("cursor"); c != "" {
		parsed, err := uuid.Parse(c)
		if err != nil {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request", "Bad Request", "cursor must be a valid UUID")
			return
		}
		cursor = parsed
	}
	pageSize := int32(defaultPageSize)
	if ps := q.Get("page_size"); ps != "" {
		n, err := strconv.Atoi(ps)
		if err != nil || n < 1 || n > maxPageSize {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request", "Bad Request", "page_size must be between 1 and 50")
			return
		}
		pageSize = int32(n)
	}

	page, err := h.admin.List(r.Context(), cursor, pageSize)
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not list suggestions")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, page)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageSuggestions(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can manage the suggestion box")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "suggestion_id"))
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request", "Bad Request", "suggestion_id must be a valid UUID")
		return
	}
	if err := h.admin.Delete(r.Context(), claims.UserID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpproblem.WriteProblem(w, r, http.StatusNotFound, "not-found", "Not Found", "Suggestion not found")
		} else {
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not delete suggestion")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
