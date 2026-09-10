package auditlog

import (
	"net/http"
	"time"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
	"github.com/google/uuid"
)

// Handler expone la lectura administrativa de audit_log. Solo AdminService
// — no existe lectura de jugador para esta tabla.
type Handler struct {
	admin *AdminService
}

func NewHandler(admin *AdminService) *Handler {
	return &Handler{admin: admin}
}

func canReadAuditLog(role domain.UserRole) bool {
	return role == domain.RoleAdmin
}

// List maneja GET /api/v1/admin/audit-log. Todos los filtros son opcionales
// y combinables: actor_account_id, action, entity_type, from/to (RFC3339),
// cursor (opaco, devuelto por la página anterior) y page_size (1-50).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || !canReadAuditLog(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden",
			"Forbidden", "Only admins can read the audit log")
		return
	}

	q := r.URL.Query()
	var f Filters

	if v := q.Get("actor_account_id"); v != "" {
		parsed, err := uuid.Parse(v)
		if err != nil {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request",
				"Bad Request", "actor_account_id must be a valid UUID")
			return
		}
		f.ActorAccountID = parsed
	}
	f.Action = q.Get("action")
	f.EntityType = q.Get("entity_type")
	if v := q.Get("from"); v != "" {
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request",
				"Bad Request", "from must be a valid RFC3339 timestamp")
			return
		}
		f.From = parsed
	}
	if v := q.Get("to"); v != "" {
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request",
				"Bad Request", "to must be a valid RFC3339 timestamp")
			return
		}
		f.To = parsed
	}
	f.Cursor = q.Get("cursor")

	pageSize, err := parsePageSize(q.Get("page_size"))
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request",
			"Bad Request", "page_size must be between 1 and 50")
		return
	}
	f.PageSize = pageSize

	page, err := h.admin.List(r.Context(), claims.UserID, f)
	if err != nil {
		if err == ErrValidation {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request",
				"Bad Request", "Invalid cursor")
			return
		}
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
			"Internal Server Error", "Could not retrieve audit log")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, page)
}
