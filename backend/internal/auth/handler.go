package auth

import (
	"errors"
	"html"
	"log/slog"
	"net"
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

// Register handles POST /api/v1/auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error",
				"Validation Error", err.Error())
		case errors.Is(err, ErrEmailConflict):
			httpproblem.WriteProblem(w, r, http.StatusConflict, "conflict",
				"Email Already Registered", "An active account with this email already exists")
		case errors.Is(err, ErrAuthBusy):
			httpproblem.WriteProblem(w, r, http.StatusTooManyRequests, "auth-busy",
				"Too Many Requests", "Authentication service is busy; retry shortly")
		default:
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
				"Internal Server Error", "An unexpected error occurred")
		}
		return
	}

	httpproblem.WriteJSON(w, http.StatusCreated, resp)
}

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
			// Use identical message for both to prevent user enumeration.
			httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
				"Authentication Failed", "Invalid email or password")
		case errors.Is(err, ErrAccountSuspended):
			httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden",
				"Account Restricted", "This account has been suspended or deleted")
		case errors.Is(err, ErrPendingTutor):
			httpproblem.WriteProblem(w, r, http.StatusForbidden, "pending-tutor-consent",
				"Tutor Consent Required", "Tutor consent is required before login")
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
		case errors.Is(err, ErrPendingTutor):
			httpproblem.WriteProblem(w, r, http.StatusForbidden, "pending-tutor-consent",
				"Tutor Consent Required", "Tutor consent is required before login")
		default:
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
				"Internal Server Error", "An unexpected error occurred")
		}
		return
	}

	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) TutorConsent(w http.ResponseWriter, r *http.Request) {
	var req TutorConsentRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}

	ip := net.ParseIP("0.0.0.0")
	if parsed := net.ParseIP(httputil.ClientIP(r)); parsed != nil {
		ip = parsed
	}

	if err := h.svc.SubmitTutorConsent(r.Context(), req, ip, r.UserAgent()); err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error",
				"Validation Error", err.Error())
		case errors.Is(err, ErrMailSend):
			slog.Error("tutor consent email send failed", "error", err)
			httpproblem.WriteProblem(w, r, http.StatusBadGateway, "email-delivery-failed",
				"Email Delivery Failed",
				"No se pudo enviar el correo de verificación al tutor. Intente nuevamente más tarde.")
		default:
			slog.Error("tutor consent submit failed", "error", err)
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
				"Internal Server Error", "Could not register tutor consent")
		}
		return
	}

	// 202: the tutor must still click the emailed link before the account is
	// activated. The message is intentionally uniform so it does not reveal
	// whether the referenced account exists or was pending.
	httpproblem.WriteJSON(w, http.StatusAccepted, map[string]string{
		"status":  "pending_verification",
		"message": "Si la cuenta requiere consentimiento, se envió un correo de verificación al tutor.",
	})
}

// VerifyTutorConsent handles GET /api/v1/auth/tutor-consent/verify?token=...
// It is browser-facing (the tutor clicks a link in their email), so it renders a
// small self-contained HTML page rather than JSON. On success the minor's account
// is activated and the click is recorded as legal consent evidence.
func (h *Handler) VerifyTutorConsent(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	ip := net.ParseIP("0.0.0.0")
	if parsed := net.ParseIP(httputil.ClientIP(r)); parsed != nil {
		ip = parsed
	}

	err := h.svc.VerifyTutorConsent(r.Context(), token, ip, r.UserAgent())
	switch {
	case err == nil:
		writeTutorConsentPage(w, http.StatusOK, true,
			"Consentimiento confirmado",
			"La cuenta del menor ha sido activada. Ya puede iniciar sesión en la plataforma USBI.")
	case errors.Is(err, ErrTutorTokenExpired):
		writeTutorConsentPage(w, http.StatusGone, false,
			"El enlace ha caducado",
			"Este enlace de verificación superó su vigencia. Solicite al menor que genere una nueva solicitud de consentimiento.")
	case errors.Is(err, ErrTutorTokenUsed):
		writeTutorConsentPage(w, http.StatusConflict, false,
			"El enlace ya fue utilizado",
			"Este consentimiento ya había sido confirmado anteriormente. La cuenta se encuentra activa.")
	case errors.Is(err, ErrTutorTokenInvalid):
		writeTutorConsentPage(w, http.StatusBadRequest, false,
			"Enlace no válido",
			"El enlace de verificación no es válido. Verifique que lo copió completo desde el correo.")
	default:
		slog.Error("tutor consent verify failed", "error", err)
		writeTutorConsentPage(w, http.StatusInternalServerError, false,
			"Ocurrió un error",
			"No fue posible procesar la verificación en este momento. Intente nuevamente más tarde.")
	}
}

// Logout handles POST /api/v1/auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
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

// AgeUp handles POST /api/v1/auth/age-up.
func (h *Handler) AgeUp(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}

	if err := h.svc.AgeUp(r.Context(), claims.UserID, httputil.ClientIP(r), r.UserAgent()); err != nil {
		if err.Error() == "maximum age-up attempts exceeded" {
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

// Arco handles POST /api/v1/arco.
func (h *Handler) Arco(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
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
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "validation-error",
				"Validation Error", "Invalid ARCO request")
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
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
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
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
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
		case errors.Is(err, ErrValidation):
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error", "Validation Error", err.Error())
		default:
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not resolve ARCO request")
		}
		return
	}

	httpproblem.WriteJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// writeTutorConsentPage renders a minimal, self-contained, accessible HTML page
// for the tutor's browser after clicking the verification link. heading/message
// are server-controlled constants but are HTML-escaped defensively.
func writeTutorConsentPage(w http.ResponseWriter, status int, ok bool, heading, message string) {
	accent := "#18529D" // institutional blue (UV)
	if !ok {
		accent = "#b00020" // error red — deliberately non-institutional to read as an error
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	h := html.EscapeString(heading)
	m := html.EscapeString(message)
	page := `<!doctype html><html lang="es"><head>` +
		`<meta charset="utf-8">` +
		`<meta name="viewport" content="width=device-width, initial-scale=1">` +
		`<title>` + h + ` — USBI</title><style>` +
		`*{box-sizing:border-box}` +
		`body{margin:0;font-family:system-ui,-apple-system,"Segoe UI",Roboto,sans-serif;` +
		`background:#f1f5f9;color:#0f172a;display:flex;min-height:100vh;align-items:center;` +
		`justify-content:center;padding:1rem}` +
		`main{background:#fff;max-width:32rem;width:100%;border-radius:12px;padding:2rem;` +
		`box-shadow:0 10px 30px rgba(0,0,0,.08);border-top:6px solid ` + accent + `}` +
		`h1{color:` + accent + `;font-size:1.4rem;margin:0 0 .75rem;line-height:1.3}` +
		`p{font-size:1rem;line-height:1.55;margin:0}` +
		`.brand{margin-top:1.5rem;font-size:.85rem;color:#475569}` +
		`</style></head><body><main>` +
		`<h1>` + h + `</h1><p>` + m + `</p>` +
		`<p class="brand">Plataforma USBI · Universidad Veracruzana</p>` +
		`</main></body></html>`
	_, _ = w.Write([]byte(page))
}
