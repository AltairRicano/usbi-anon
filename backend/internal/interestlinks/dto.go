package interestlinks

import (
	"time"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

// CategoryResponse es la representación pública de una categoría del
// carrusel de "enlaces de interés" (sección "Más" del frontend).
type CategoryResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	DisplayOrder int16     `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// LinkResponse es una tarjeta del carrusel.
type LinkResponse struct {
	ID          uuid.UUID `json:"id"`
	CategoryID  uuid.UUID `json:"category_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CategoryWithLinks agrupa una categoría con sus enlaces para la vista de jugador.
// Links nunca es nil: una categoría sin tarjetas aparece con un arreglo vacío
// para que el frontend pueda decidir si oculta el carrusel vacío.
type CategoryWithLinks struct {
	Category CategoryResponse `json:"category"`
	Links    []LinkResponse   `json:"links"`
}

// InterestLinksResponse, CategoriesResponse y LinksResponse envuelven a sus
// respectivos listados en `{"items": [...]}` de forma consistente con los demás endpoints.
type InterestLinksResponse struct {
	Items []CategoryWithLinks `json:"items"`
}

type CategoriesResponse struct {
	Items []CategoryResponse `json:"items"`
}

type LinksResponse struct {
	Items []LinkResponse `json:"items"`
}

// CreateCategoryRequest y UpdateCategoryRequest son el cuerpo de
// POST/PATCH /admin/interest-link-categories.
type CreateCategoryRequest struct {
	Name         string `json:"name"`
	DisplayOrder int16  `json:"display_order"`
}

type UpdateCategoryRequest struct {
	Name         string `json:"name"`
	DisplayOrder int16  `json:"display_order"`
}

// CreateLinkRequest y UpdateLinkRequest son el cuerpo de
// POST/PATCH /admin/interest-links.
type CreateLinkRequest struct {
	CategoryID  uuid.UUID `json:"category_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	URL         string    `json:"url"`
}

type UpdateLinkRequest struct {
	CategoryID  uuid.UUID `json:"category_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	URL         string    `json:"url"`
}

func categoryToResponse(c repository.InterestLinkCategory) CategoryResponse {
	return CategoryResponse{
		ID:           c.ID,
		Name:         c.Name,
		DisplayOrder: c.DisplayOrder,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

func linkToResponse(l repository.InterestLink) LinkResponse {
	return LinkResponse{
		ID:          l.ID,
		CategoryID:  l.CategoryID,
		Title:       l.Title,
		Description: l.Description,
		Color:       l.Color,
		URL:         l.URL,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}
