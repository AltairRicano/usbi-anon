package sync

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpjson"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
	"github.com/google/uuid"
)

// Handler expone el endpoint HTTP de sincronización. (Relleno)
type Handler struct {
	svc *Service
}

// NewHandler crea un sync.Handler. (Relleno)
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// SyncData maneja POST /api/v1/sync.
//
// Contrato de seguridad:
//   - El HMAC se verifica sobre un payload técnico canónico, no sobre el JSON en crudo.
//   - hmac_signature en el cuerpo JSON son bytes codificados en base64 (codificación estándar).
//   - El user_id en el cuerpo DEBE coincidir con las claims del JWT (verificado por middleware). (Útil)
func (h *Handler) SyncData(w http.ResponseWriter, r *http.Request) {
	// Leer el cuerpo una vez. La verificación HMAC usa el payload canónico decodificado. (Útil)
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpproblem.WriteProblem(w, r, http.StatusRequestEntityTooLarge, "payload-too-large",
				"Payload Too Large", "Request body exceeds the configured size limit")
			return
		}
		httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request",
			"Bad Request", "Could not read request body")
		return
	}

	var req domain.SyncEventRequest
	if err := httpjson.DecodeStrict(bytes.NewReader(rawBody), &req); err != nil {
		httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request",
			"Bad Request", "Invalid JSON payload")
		return
	}
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok || claims.UserID != req.UserID {
		httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "invalid-sync-user",
			"Invalid Sync Payload", "sync user_id must match the authenticated user")
		return
	}

	resp, err := h.svc.ProcessSync(r.Context(), req, req.HMACSignature)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidPayload):
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error",
				"Validation Error", err.Error())
		case errors.Is(err, ErrInvalidSignature):
			httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "invalid-signature",
				"HMAC Verification Failed", "The payload signature is invalid")
		default:
			httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
				"Internal Server Error", "An unexpected error occurred")
		}
		return
	}

	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

// ListMyHistory maneja GET /api/v1/sync/events: el historial de
// sincronización offline del propio jugador autenticado, sin variante de
// admin (B4, estado_proyecto.md 2026-09-09) — un admin que jugó offline ve
// aquí sus propios eventos, como cualquier jugador, nunca los de otra
// cuenta.
func (h *Handler) ListMyHistory(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	if !ok {
		httpproblem.WriteProblem(w, r, http.StatusUnauthorized, "unauthorized",
			"Unauthorized", "Missing JWT claims in context")
		return
	}

	q := r.URL.Query()

	deviceID := uuid.Nil
	if d := q.Get("device_id"); d != "" {
		parsed, err := uuid.Parse(d)
		if err != nil {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request",
				"Bad Request", "device_id must be a valid UUID")
			return
		}
		deviceID = parsed
	}

	var cursor time.Time
	if c := q.Get("cursor"); c != "" {
		parsed, err := time.Parse(time.RFC3339Nano, c)
		if err != nil {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request",
				"Bad Request", "cursor must be a valid RFC3339 timestamp")
			return
		}
		cursor = parsed
	}

	pageSize := int32(defaultHistoryPageSize)
	if ps := q.Get("page_size"); ps != "" {
		n, err := strconv.Atoi(ps)
		if err != nil || n < 1 || n > maxHistoryPageSize {
			httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request",
				"Bad Request", "page_size must be between 1 and 50")
			return
		}
		pageSize = int32(n)
	}

	page, err := h.svc.ListMySyncEvents(r.Context(), claims.UserID, deviceID, cursor, pageSize)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error",
				"Validation Error", "Invalid sync history query")
			return
		}
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error",
			"Internal Server Error", "Could not retrieve sync history")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, page)
}
