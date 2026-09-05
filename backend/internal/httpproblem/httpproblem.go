// Package httpproblem centraliza las respuestas de error RFC 7807 y el escritor de JSON
// de éxito que previamente estaban copiados tal cual entre auth, levels, devices,
// sync y el router (hallazgo de auditoría B8). Una sola implementación garantiza que la
// envoltura de error nunca varíe entre paquetes. (Útil)
package httpproblem

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/altair/usbi-anon-backend/internal/domain"
)

// WriteProblem emite una respuesta RFC 7807 application/problem+json. slug se convierte
// en el sufijo del URI type. (Relleno)
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

// WriteJSON serializa v como application/json con el estado dado. (Relleno)
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteDecodeProblem mapea un error de decodificación de cuerpo de petición a la respuesta
// problem correcta: 413 cuando el cuerpo excede el límite de tamaño, 400 de lo contrario. (Útil)
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
