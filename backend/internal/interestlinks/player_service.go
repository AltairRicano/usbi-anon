package interestlinks

import (
	"context"

	"github.com/altair/usbi-anon-backend/internal/repository"
)

// PlayerService corre sobre el pool de jugador (usbi_app) con permisos de solo lectura,
// exponiendo la consulta agrupada de categorías y enlaces sin permitir mutaciones.
type PlayerService struct {
	repo *repository.Queries
}

func NewPlayerService(repo *repository.Queries) *PlayerService {
	return &PlayerService{repo: repo}
}

// ListGrouped arma GET /interest-links: cada categoría (en su
// display_order) junto con sus tarjetas (en su created_at). Dos consultas
// en vez de un JOIN porque el volumen esperado es pequeño (contenido
// editorial curado a mano por un admin, no datos de usuario) y así se evita
// tener que deduplicar filas de categoría en Go para el caso JOIN de una
// categoría con muchas tarjetas.
func (s *PlayerService) ListGrouped(ctx context.Context) (InterestLinksResponse, error) {
	categories, err := s.repo.ListInterestLinkCategories(ctx)
	if err != nil {
		return InterestLinksResponse{}, err
	}

	groups := make([]CategoryWithLinks, 0, len(categories))
	for _, category := range categories {
		links, err := s.repo.ListInterestLinksByCategory(ctx, category.ID)
		if err != nil {
			return InterestLinksResponse{}, err
		}
		linkResponses := make([]LinkResponse, 0, len(links))
		for _, link := range links {
			linkResponses = append(linkResponses, linkToResponse(link))
		}
		groups = append(groups, CategoryWithLinks{
			Category: categoryToResponse(category),
			Links:    linkResponses,
		})
	}
	return InterestLinksResponse{Items: groups}, nil
}
