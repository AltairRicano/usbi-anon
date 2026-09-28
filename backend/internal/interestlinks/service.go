// Package interestlinks implementa la sección de enlaces de interés del sistema.
// Separa las operaciones en PlayerService (solo lectura para jugadores sobre el pool usbi_app)
// y AdminService (CRUD completo para administración sobre el pool usbi_moderador).
package interestlinks

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	ErrValidation       = errors.New("validation error")
	ErrNotFound         = errors.New("not found")
	ErrCategoryHasLinks = errors.New("category still has links")
)

// colorHexPattern espeja el CHECK de la columna interest_links.color —
// validarlo aquí primero da un 422 legible en vez del error crudo del
// driver que devolvería el CHECK de Postgres.
var colorHexPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

const (
	categoryNameMaxLen    = 200
	linkTitleMaxLen       = 50
	linkDescriptionMaxLen = 100
)

func validateCategoryInput(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > categoryNameMaxLen {
		return ErrValidation
	}
	return nil
}

func validateLinkInput(title, description, color, rawURL string) error {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" || len(title) > linkTitleMaxLen {
		return ErrValidation
	}
	if description == "" || len(description) > linkDescriptionMaxLen {
		return ErrValidation
	}
	if !colorHexPattern.MatchString(color) {
		return ErrValidation
	}
	if !isHTTPURL(rawURL) {
		return ErrValidation
	}
	return nil
}

// isHTTPURL espeja el CHECK `url ~* '^https?://'` de interest_links: un
// esquema distinto (javascript:, data:, …) no debe ni siquiera intentar
// guardarse.
func isHTTPURL(rawURL string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}
