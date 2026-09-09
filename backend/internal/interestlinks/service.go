// Package interestlinks implementa la sección "Más" / "Enlaces de interés"
// del frontend (migración 0003_enlaces_interes_y_sugerencias, F4 —
// estado_proyecto.md 2026-09-09). Antes de F4, interest_link_categories e
// interest_links existían en el esquema sin ningún código Go que las
// tocara — F1 las marcó como huérfanas.
//
// Sigue el mismo patrón PlayerService/AdminService que F3 dejó establecido
// para internal/levels e internal/quiz: un archivo de lógica compartida
// (este) más dos Service, cada uno con su propio *repository.Queries — el
// del jugador sobre el pool usbi_app (solo lectura, 00_roles_unificado.sql),
// el de administración sobre usbi_moderador (CRUD completo).
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
