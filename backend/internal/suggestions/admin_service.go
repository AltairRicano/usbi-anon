package suggestions

import (
	"context"
	"database/sql"
	"errors"

	"github.com/altair/usbi-anon-backend/internal/audit"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

const defaultPageSize = 20
const maxPageSize = 50

// AdminService corre sobre el pool de moderador (usbi_moderador), con permisos de
// lectura y eliminación (las sugerencias no se editan).
type AdminService struct {
	repo *repository.Queries
}

func NewAdminService(repo *repository.Queries) *AdminService {
	return &AdminService{repo: repo}
}

// List pagina por id (UUIDv7, ordenados en el tiempo) en vez de por
// submitted_at directamente — evita un cursor compuesto y da el mismo
// orden "más recientes primero" que ya pedía el índice
// suggestions_submitted_at_idx.
func (s *AdminService) List(ctx context.Context, cursor uuid.UUID, pageSize int32) (Page, error) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	items, err := s.repo.ListSuggestions(ctx, repository.ListSuggestionsParams{
		Cursor:   cursor,
		PageSize: pageSize + 1,
	})
	if err != nil {
		return Page{}, err
	}

	page := Page{}
	hasMore := int32(len(items)) > pageSize
	if hasMore {
		items = items[:pageSize]
	}
	page.Items = make([]SuggestionResponse, 0, len(items))
	for _, item := range items {
		page.Items = append(page.Items, toResponse(item))
	}
	if hasMore {
		page.NextCursor = items[len(items)-1].ID.String()
	}
	return page, nil
}

// Delete no registra un audit.Log con el contenido de la sugerencia: hacerlo
// dejaría en audit_log el texto libre que un admin acaba de borrar, lo cual
// derrota el propósito de que un admin pueda depurar el buzón. Se audita
// solo el ID, igual que incidents/levels auditan operaciones sin exponer el
// cuerpo sensible en el "after" cuando no aplica.
func (s *AdminService) Delete(ctx context.Context, adminID, suggestionID uuid.UUID) error {
	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	rows, err := qtx.DeleteSuggestion(ctx, suggestionID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	if err := audit.Log(ctx, qtx, audit.Entry{
		ActorID:    adminID,
		Action:     "suggestion.delete",
		EntityType: "suggestion",
		EntityID:   suggestionID,
	}); err != nil {
		return err
	}
	return tx.Commit()
}
