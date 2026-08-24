// Package httpproblem centralises the RFC 7807 error responses and JSON success
// writer that were previously copy-pasted verbatim across auth, levels, devices,
// sync and the router (audit finding B8). A single implementation guarantees the
// error envelope can never drift between packages.
package httpproblem

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/altair/usbi-anon-backend/internal/domain"
)

// WriteProblem emits an RFC 7807 application/problem+json response. slug becomes
// the suffix of the type URI.
func WriteProblem(w http.ResponseWriter, r *http.Request, status int, slug, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(domain.ProblemDetails{
		Type:     "https://api.usbi.edu.mx/errors/" + slug,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: r.URL.Path,
	})
}

// WriteJSON serialises v as application/json with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteDecodeProblem maps a request-body decode error to the right problem
// response: 413 when the body exceeded the size cap, 400 otherwise.
func WriteDecodeProblem(w http.ResponseWriter, r *http.Request, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		WriteProblem(w, r, http.StatusRequestEntityTooLarge, "payload-too-large",
			"Payload Too Large", "Request body exceeds the configured size limit")
		return
	}
	WriteProblem(w, r, http.StatusBadRequest, "bad-request",
		"Bad Request", "Invalid JSON body")
}
