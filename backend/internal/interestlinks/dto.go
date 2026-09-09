package interestlinks

import (
	"time"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

// CategoryResponse es la representación pública de una categoría del
// carrusel de "enlaces de interés" (sección "Más" del frontend). (Relleno)
type CategoryResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	DisplayOrder int16     `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// LinkResponse es una tarjeta del carrusel. (Relleno)
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

// CategoryWithLinks agrupa una categoría con sus tarjetas — es la forma que
// pide GET /interest-links para el jugador (estado_proyecto.md 2026-09-08:
// "GET /interest-links agrupado por categoría"). Links nunca es nil: una
// categoría sin tarjetas aparece con un arreglo vacío para que el frontend
// pueda decidir si oculta el carrusel vacío. (Relleno)
type CategoryWithLinks struct {
	Category CategoryResponse `json:"category"`
	Links    []LinkResponse   `json:"links"`
}

// CreateCategoryRequest/UpdateCategoryRequest son el cuerpo de
// POST/PATCH /admin/interest-link-categories. (Relleno)
type CreateCategoryRequest struct {
	Name         string `json:"name"`
	DisplayOrder int16  `json:"display_order"`
}

type UpdateCategoryRequest struct {
	Name         string `json:"name"`
	DisplayOrder int16  `json:"display_order"`
}

// CreateLinkRequest/UpdateLinkRequest son el cuerpo de
// POST/PATCH /admin/interest-links. (Relleno)
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
