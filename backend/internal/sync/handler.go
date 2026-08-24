package sync

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpjson"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
)

// Handler exposes the sync HTTP endpoint.
type Handler struct {
	svc *Service
}

// NewHandler creates a sync.Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// SyncData handles POST /api/v1/sync.
//
// Security contract:
//   - The HMAC is verified over a canonical technical payload, not over raw JSON.
//   - hmac_signature in the JSON body is base64-encoded bytes (standard encoding).
//   - The user_id in the body MUST match the JWT claims (checked by middleware).
func (h *Handler) SyncData(w http.ResponseWriter, r *http.Request) {
	// Read the body once. HMAC verification uses the decoded canonical payload.
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
