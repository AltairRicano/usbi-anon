package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpjson"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
	"github.com/altair/usbi-anon-backend/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler exposes auth HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates an auth.Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ── Registro en 3 pasos ──────────────────────────────────────────────────

// RegisterQuestions handles POST /api/v1/auth/register/questions.
func (h *Handler) RegisterQuestions(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.RegisterQuestions(r.Context())
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
			"Internal Server Error", "Could not load registration questions")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

// RegisterAnswers handles POST /api/v1/auth/register/answers.
func (h *Handler) RegisterAnswers(w http.ResponseWriter, r *http.Request) {
	var req RegisterAnswersRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	resp, err := h.svc.RegisterAnswers(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error",
				"Validation Error", err.Error())
		case errors.Is(err, ErrAuthBusy):
			httpproblem.WriteProblem(w, r, http.StatusTooManyRequests, "auth-busy",
				"Too Many Requests", "Authentication service is busy; retry shortly")
		default:
			slog.Error("register answers failed", "error", err)
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
				"Internal Server Error", "An unexpected error occurred")
		}
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

// RegisterConfirm handles POST /api/v1/auth/register/confirm.
func (h *Handler) RegisterConfirm(w http.ResponseWriter, r *http.Request) {
	var req RegisterConfirmRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	resp, err := h.svc.RegisterConfirm(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error",
				"Validation Error", err.Error())
		case errors.Is(err, ErrRegistrationTokenExpired):
			httpproblem.WriteProblem(w, r, http.StatusGone, "registration-token-expired",
				"Registration Token Expired", "Start the registration flow again")
		case errors.Is(err, ErrRegistrationTokenInvalid):
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "registration-token-invalid",
				"Registration Token Invalid", "The registration token is malformed or was tampered with")
		case errors.Is(err, ErrNicknameConflict):
			httpproblem.WriteProblem(w, r, http.StatusConflict, "conflict",
				"Nickname Already Taken", "Choose a different nickname candidate and try again")
		case errors.Is(err, ErrAuthBusy):
			httpproblem.WriteProblem(w, r, http.StatusTooManyRequests, "auth-busy",
				"Too Many Requests", "Authentication service is busy; retry shortly")
		default:
			slog.Error("register confirm failed", "error", err)
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
				"Internal Server Error", "An unexpected error occurred")
		}
		return
	}
	httpproblem.WriteJSON(w, http.StatusCreated, resp)
}

// ── Login / sesión ────────────────────────────────────────────────────────

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	resp, err := h.svc.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error",
				"Validation Error", err.Error())
		case errors.Is(err, ErrUserNotFound), errors.Is(err, ErrInvalidPassword):
			// Use identical message for both to prevent nickname enumeration.
			httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
				"Authentication Failed", "Invalid nickname or password")
		case errors.Is(err, ErrAccountSuspended):
			httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden",
				"Account Restricted", "This account has been suspended or deleted")
		case errors.Is(err, ErrAuthBusy):
			httpproblem.WriteProblem(w, r, http.StatusTooManyRequests, "auth-busy",
				"Too Many Requests", "Authentication service is busy; retry shortly")
		default:
			slog.Error("login failed", "error", err)
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
				"Internal Server Error", "An unexpected error occurred")
		}
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	resp, err := h.svc.Refresh(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRefresh):
			httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "invalid-refresh-token",
				"Authentication Failed", "Invalid refresh token")
		case errors.Is(err, ErrAccountSuspended):
			httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden",
				"Account Restricted", "This account has been suspended or deleted")
		default:
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
				"Internal Server Error", "An unexpected error occurred")
		}
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

// Logout handles POST /api/v1/auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}

	if err := h.svc.Logout(r.Context(), claims.UserID); err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
			"Internal Server Error", "Could not process logout")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Me handles GET /api/v1/auth/me.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, map[string]any{
		"user_id": claims.UserID,
		"role":    claims.Role,
	})
}

// AgeUp handles POST /api/v1/auth/age-up.
func (h *Handler) AgeUp(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}

	if err := h.svc.AgeUp(r.Context(), claims.UserID, httputil.ClientIP(r), r.UserAgent()); err != nil {
		if errors.Is(err, ErrTooManyAgeUpAttempts) {
			httpproblem.WriteProblem(w, r, http.StatusTooManyRequests, "too-many-requests",
				"Too Many Requests", err.Error())
			return
		}
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
			"Internal Server Error", "Could not process age-up request")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "User adult status updated"})
}

// CancelSelf handles DELETE /api/v1/auth/me.
func (h *Handler) CancelSelf(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}
	if err := h.svc.CancelSelf(r.Context(), claims.UserID); err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
			"Internal Server Error", "Could not cancel account")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── ARCO ──────────────────────────────────────────────────────────────────

// Arco handles POST /api/v1/arco.
func (h *Handler) Arco(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}

	var req ArcoRequestDTO
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	requestID, err := h.svc.SubmitArcoRequest(r.Context(), claims.UserID, req)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error",
				"Validation Error", err.Error())
		} else {
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
				"Internal Server Error", "Could not submit ARCO request")
		}
		return
	}
	httpproblem.WriteJSON(w, http.StatusCreated, ArcoResponseDTO{
		RequestID: requestID,
		Status:    "pending",
		Message:   "ARCO request submitted successfully",
	})
}

func (h *Handler) ListPendingArco(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}

	limit := int32(50)
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 1 || parsed > 100 {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request", "Bad Request", "limit must be between 1 and 100")
			return
		}
		limit = int32(parsed)
	}

	resp, err := h.svc.ListPendingArcoRequests(r.Context(), *claims, limit)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins or directors can list ARCO requests")
		} else {
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not list ARCO requests")
		}
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) ResolveArco(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}

	requestID, err := uuid.Parse(chi.URLParam(r, "request_id"))
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request", "Bad Request", "request_id must be a valid UUID")
		return
	}

	var req ResolveArcoRequestDTO
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	if err := h.svc.ResolveArcoRequest(r.Context(), *claims, requestID, req, httputil.ClientIP(r), r.UserAgent()); err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins or directors can resolve ARCO requests")
		case errors.Is(err, ErrNotFound):
			httpproblem.WriteProblem(w, r, http.StatusNotFound, "not-found", "Not Found", "ARCO request not found")
		case errors.Is(err, ErrValidation):
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error", "Validation Error", err.Error())
		default:
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not resolve ARCO request")
		}
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

// ── Administración de cuentas ────────────────────────────────────────────

// CreateAdminAccount handles POST /api/v1/admin/accounts.
func (h *Handler) CreateAdminAccount(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}
	var req AdminCreateAccountRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}
	resp, err := h.svc.CreateAdminAccount(r.Context(), *claims, req)
	if err != nil {
		writeAdminServiceError(w, r, err, "Could not create account")
		return
	}
	httpproblem.WriteJSON(w, http.StatusCreated, resp)
}

// DeleteAdminAccount handles DELETE /api/v1/admin/accounts/{account_id}.
func (h *Handler) DeleteAdminAccount(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}
	targetID, ok := parseURLUUID(w, r, "account_id")
	if !ok {
		return
	}
	if err := h.svc.DeleteAccount(r.Context(), *claims, targetID); err != nil {
		writeAdminServiceError(w, r, err, "Could not delete account")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetAccountQuizAnswers handles GET /api/v1/admin/accounts/{account_id}/quiz-answers.
func (h *Handler) GetAccountQuizAnswers(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}
	targetID, ok := parseURLUUID(w, r, "account_id")
	if !ok {
		return
	}
	resp, err := h.svc.GetAccountQuizAnswers(r.Context(), *claims, targetID)
	if err != nil {
		writeAdminServiceError(w, r, err, "Could not load quiz answers")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

// ResetAccountPassword handles POST /api/v1/admin/accounts/{account_id}/reset-password.
func (h *Handler) ResetAccountPassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}
	targetID, ok := parseURLUUID(w, r, "account_id")
	if !ok {
		return
	}
	resp, err := h.svc.ResetAccountPassword(r.Context(), *claims, targetID)
	if err != nil {
		writeAdminServiceError(w, r, err, "Could not reset password")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

// ── Internal helpers ──────────────────────────────────────────────────────

func claimsFromContext(r *http.Request) (*domain.JWTClaims, bool) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	return claims, ok
}

func parseURLUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request", "Bad Request", key+" must be a valid UUID")
		return uuid.Nil, false
	}
	return parsed, true
}

func writeAdminServiceError(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	switch {
	case errors.Is(err, ErrForbidden):
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can perform this action")
	case errors.Is(err, ErrCannotDeleteAdmin):
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "An admin account cannot delete another admin account")
	case errors.Is(err, ErrNotFound):
		httpproblem.WriteProblem(w, r, http.StatusNotFound, "not-found", "Not Found", "Account not found")
	case errors.Is(err, ErrNicknameConflict):
		httpproblem.WriteProblem(w, r, http.StatusConflict, "conflict", "Nickname Already Taken", "Choose a different nickname")
	case errors.Is(err, ErrValidation):
		httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error", "Validation Error", err.Error())
	case errors.Is(err, ErrAuthBusy):
		httpproblem.WriteProblem(w, r, http.StatusTooManyRequests, "auth-busy", "Too Many Requests", "Authentication service is busy; retry shortly")
	default:
		slog.Error("admin account operation failed", "error", err)
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", fallback)
	}
}
