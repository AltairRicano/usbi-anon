package interestlinks

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/altair/usbi-anon-backend/internal/audit"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// AdminService corre sobre el pool de moderador (usbi_moderador), que tiene
// permisos para la gestión de categorías y enlaces de interés.
type AdminService struct {
	repo *repository.Queries
}

func NewAdminService(repo *repository.Queries) *AdminService {
	return &AdminService{repo: repo}
}

func newID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}

func (s *AdminService) ListCategories(ctx context.Context) (CategoriesResponse, error) {
	categories, err := s.repo.ListInterestLinkCategories(ctx)
	if err != nil {
		return CategoriesResponse{}, err
	}
	resp := make([]CategoryResponse, 0, len(categories))
	for _, c := range categories {
		resp = append(resp, categoryToResponse(c))
	}
	return CategoriesResponse{Items: resp}, nil
}

func (s *AdminService) CreateCategory(ctx context.Context, adminID uuid.UUID, req CreateCategoryRequest) (CategoryResponse, error) {
	name := strings.TrimSpace(req.Name)
	if err := validateCategoryInput(name); err != nil {
		return CategoryResponse{}, err
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return CategoryResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	category, err := qtx.CreateInterestLinkCategory(ctx, repository.CreateInterestLinkCategoryParams{
		ID:           newID(),
		Name:         name,
		DisplayOrder: req.DisplayOrder,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return CategoryResponse{}, ErrValidation
		}
		return CategoryResponse{}, err
	}
	resp := categoryToResponse(category)
	if err := audit.Log(ctx, qtx, audit.Entry{
		ActorID:    adminID,
		Action:     "interest_link_category.create",
		EntityType: "interest_link_category",
		EntityID:   category.ID,
		After:      categoryAuditPayload(resp),
	}); err != nil {
		return CategoryResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return CategoryResponse{}, err
	}
	return resp, nil
}

func (s *AdminService) UpdateCategory(ctx context.Context, adminID, categoryID uuid.UUID, req UpdateCategoryRequest) (CategoryResponse, error) {
	name := strings.TrimSpace(req.Name)
	if err := validateCategoryInput(name); err != nil {
		return CategoryResponse{}, err
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return CategoryResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	category, err := qtx.UpdateInterestLinkCategory(ctx, repository.UpdateInterestLinkCategoryParams{
		ID:           categoryID,
		Name:         name,
		DisplayOrder: req.DisplayOrder,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CategoryResponse{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return CategoryResponse{}, ErrValidation
		}
		return CategoryResponse{}, err
	}
	resp := categoryToResponse(category)
	if err := audit.Log(ctx, qtx, audit.Entry{
		ActorID:    adminID,
		Action:     "interest_link_category.update",
		EntityType: "interest_link_category",
		EntityID:   category.ID,
		After:      categoryAuditPayload(resp),
	}); err != nil {
		return CategoryResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return CategoryResponse{}, err
	}
	return resp, nil
}

// DeleteCategory pregunta primero por enlaces vivos en vez de dejar que el ON DELETE
// RESTRICT de interest_links.category_id devuelva un error crudo de Postgres:
// quien administre debe vaciar la categoría o reasignar sus enlaces primero.
func (s *AdminService) DeleteCategory(ctx context.Context, adminID, categoryID uuid.UUID) error {
	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	count, err := qtx.CountInterestLinksByCategory(ctx, categoryID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrCategoryHasLinks
	}

	rows, err := qtx.DeleteInterestLinkCategory(ctx, categoryID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	if err := audit.Log(ctx, qtx, audit.Entry{
		ActorID:    adminID,
		Action:     "interest_link_category.delete",
		EntityType: "interest_link_category",
		EntityID:   categoryID,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *AdminService) ListLinks(ctx context.Context) (LinksResponse, error) {
	links, err := s.repo.ListInterestLinks(ctx)
	if err != nil {
		return LinksResponse{}, err
	}
	resp := make([]LinkResponse, 0, len(links))
	for _, l := range links {
		resp = append(resp, linkToResponse(l))
	}
	return LinksResponse{Items: resp}, nil
}

func (s *AdminService) CreateLink(ctx context.Context, adminID uuid.UUID, req CreateLinkRequest) (LinkResponse, error) {
	title := strings.TrimSpace(req.Title)
	description := strings.TrimSpace(req.Description)
	if req.CategoryID == uuid.Nil {
		return LinkResponse{}, ErrValidation
	}
	if err := validateLinkInput(title, description, req.Color, req.URL); err != nil {
		return LinkResponse{}, err
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return LinkResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	link, err := qtx.CreateInterestLink(ctx, repository.CreateInterestLinkParams{
		ID:          newID(),
		CategoryID:  req.CategoryID,
		Title:       title,
		Description: description,
		Color:       req.Color,
		URL:         req.URL,
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return LinkResponse{}, ErrValidation
		}
		return LinkResponse{}, err
	}
	resp := linkToResponse(link)
	if err := audit.Log(ctx, qtx, audit.Entry{
		ActorID:    adminID,
		Action:     "interest_link.create",
		EntityType: "interest_link",
		EntityID:   link.ID,
		After:      linkAuditPayload(resp),
	}); err != nil {
		return LinkResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return LinkResponse{}, err
	}
	return resp, nil
}

func (s *AdminService) UpdateLink(ctx context.Context, adminID, linkID uuid.UUID, req UpdateLinkRequest) (LinkResponse, error) {
	title := strings.TrimSpace(req.Title)
	description := strings.TrimSpace(req.Description)
	if req.CategoryID == uuid.Nil {
		return LinkResponse{}, ErrValidation
	}
	if err := validateLinkInput(title, description, req.Color, req.URL); err != nil {
		return LinkResponse{}, err
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return LinkResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	link, err := qtx.UpdateInterestLink(ctx, repository.UpdateInterestLinkParams{
		ID:          linkID,
		CategoryID:  req.CategoryID,
		Title:       title,
		Description: description,
		Color:       req.Color,
		URL:         req.URL,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LinkResponse{}, ErrNotFound
		}
		if isForeignKeyViolation(err) {
			return LinkResponse{}, ErrValidation
		}
		return LinkResponse{}, err
	}
	resp := linkToResponse(link)
	if err := audit.Log(ctx, qtx, audit.Entry{
		ActorID:    adminID,
		Action:     "interest_link.update",
		EntityType: "interest_link",
		EntityID:   link.ID,
		After:      linkAuditPayload(resp),
	}); err != nil {
		return LinkResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return LinkResponse{}, err
	}
	return resp, nil
}

func (s *AdminService) DeleteLink(ctx context.Context, adminID, linkID uuid.UUID) error {
	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	rows, err := qtx.DeleteInterestLink(ctx, linkID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	if err := audit.Log(ctx, qtx, audit.Entry{
		ActorID:    adminID,
		Action:     "interest_link.delete",
		EntityType: "interest_link",
		EntityID:   linkID,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func categoryAuditPayload(c CategoryResponse) map[string]any {
	return map[string]any{"id": c.ID, "name": c.Name, "display_order": c.DisplayOrder}
}

func linkAuditPayload(l LinkResponse) map[string]any {
	return map[string]any{
		"id": l.ID, "category_id": l.CategoryID, "title": l.Title, "url": l.URL,
	}
}

// isUniqueViolation y isForeignKeyViolation traducen los códigos de error de
// Postgres (23505 y 23503) para evitar propagar errores crudos del driver:
// el nombre de categoría es UNIQUE y category_id es una clave foránea RESTRICT,
// por lo que violaciones de integridad referencial o unicidad se traducen a errores de validación.
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23503"
}
