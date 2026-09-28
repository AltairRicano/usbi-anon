// interest_link_queries.go es NUEVO en F4 (estado_proyecto.md 2026-09-09,
// "F4 de 4"): acceso a datos de las 3 tablas que F1 detectó sin ningún
// código Go que las tocara — interest_link_categories, interest_links y
// suggestions (migración 0003_enlaces_interes_y_sugerencias). internal/
// interestlinks es el consumidor de las dos primeras; internal/suggestions
// de la tercera.
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type InterestLinkCategory struct {
	ID           uuid.UUID
	Name         string
	DisplayOrder int16
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func scanInterestLinkCategory(row scanner) (InterestLinkCategory, error) {
	var c InterestLinkCategory
	err := row.Scan(&c.ID, &c.Name, &c.DisplayOrder, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

const interestLinkCategoryColumns = `id, name, display_order, created_at, updated_at`

// ListInterestLinkCategories ordena igual que el índice
// interest_link_categories_display_order_idx — es el orden en el que el
// carrusel de categorías se pinta en el frontend.
func (q *Queries) ListInterestLinkCategories(ctx context.Context) ([]InterestLinkCategory, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT `+interestLinkCategoryColumns+`
FROM interest_link_categories
ORDER BY display_order ASC, created_at ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []InterestLinkCategory
	for rows.Next() {
		item, err := scanInterestLinkCategory(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type CreateInterestLinkCategoryParams struct {
	ID           uuid.UUID
	Name         string
	DisplayOrder int16
}

func (q *Queries) CreateInterestLinkCategory(ctx context.Context, arg CreateInterestLinkCategoryParams) (InterestLinkCategory, error) {
	row := q.db.QueryRowContext(ctx, `
INSERT INTO interest_link_categories (id, name, display_order)
VALUES ($1, $2, $3)
RETURNING `+interestLinkCategoryColumns,
		arg.ID, arg.Name, arg.DisplayOrder)
	return scanInterestLinkCategory(row)
}

type UpdateInterestLinkCategoryParams struct {
	ID           uuid.UUID
	Name         string
	DisplayOrder int16
}

func (q *Queries) UpdateInterestLinkCategory(ctx context.Context, arg UpdateInterestLinkCategoryParams) (InterestLinkCategory, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE interest_link_categories
SET name = $2, display_order = $3, updated_at = NOW()
WHERE id = $1
RETURNING `+interestLinkCategoryColumns,
		arg.ID, arg.Name, arg.DisplayOrder)
	return scanInterestLinkCategory(row)
}

// CountInterestLinksByCategory soporta el mismo guard explícito que
// CountLevelsBySection en content_queries.go: en vez de dejar que el
// DELETE choque con el ON DELETE RESTRICT de interest_links.category_id y
// traducir el código de error de Postgres, el servicio pregunta primero y
// decide con un error de dominio propio.
func (q *Queries) CountInterestLinksByCategory(ctx context.Context, categoryID uuid.UUID) (int64, error) {
	var count int64
	err := q.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM interest_links WHERE category_id = $1`, categoryID).Scan(&count)
	return count, err
}

func (q *Queries) DeleteInterestLinkCategory(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, `DELETE FROM interest_link_categories WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

type InterestLink struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	Title       string
	Description string
	Color       string
	URL         string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func scanInterestLink(row scanner) (InterestLink, error) {
	var l InterestLink
	err := row.Scan(&l.ID, &l.CategoryID, &l.Title, &l.Description, &l.Color, &l.URL, &l.CreatedAt, &l.UpdatedAt)
	return l, err
}

const interestLinkColumns = `id, category_id, title, description, color, url, created_at, updated_at`

// ListInterestLinksByCategory ordena por created_at: la migración 0003 deja
// explícito que solo la categoría tiene orden manual, las tarjetas dentro de
// ella no.
func (q *Queries) ListInterestLinksByCategory(ctx context.Context, categoryID uuid.UUID) ([]InterestLink, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT `+interestLinkColumns+`
FROM interest_links
WHERE category_id = $1
ORDER BY created_at ASC
`, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []InterestLink
	for rows.Next() {
		item, err := scanInterestLink(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListInterestLinks trae TODAS las tarjetas sin filtrar por categoría — la
// vista de administración plana (frente a la agrupada que arma el jugador
// combinando esto con ListInterestLinkCategories).
func (q *Queries) ListInterestLinks(ctx context.Context) ([]InterestLink, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT `+interestLinkColumns+`
FROM interest_links
ORDER BY category_id, created_at ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []InterestLink
	for rows.Next() {
		item, err := scanInterestLink(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type CreateInterestLinkParams struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	Title       string
	Description string
	Color       string
	URL         string
}

func (q *Queries) CreateInterestLink(ctx context.Context, arg CreateInterestLinkParams) (InterestLink, error) {
	row := q.db.QueryRowContext(ctx, `
INSERT INTO interest_links (id, category_id, title, description, color, url)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING `+interestLinkColumns,
		arg.ID, arg.CategoryID, arg.Title, arg.Description, arg.Color, arg.URL)
	return scanInterestLink(row)
}

type UpdateInterestLinkParams struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	Title       string
	Description string
	Color       string
	URL         string
}

func (q *Queries) UpdateInterestLink(ctx context.Context, arg UpdateInterestLinkParams) (InterestLink, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE interest_links
SET category_id = $2, title = $3, description = $4, color = $5, url = $6, updated_at = NOW()
WHERE id = $1
RETURNING `+interestLinkColumns,
		arg.ID, arg.CategoryID, arg.Title, arg.Description, arg.Color, arg.URL)
	return scanInterestLink(row)
}

func (q *Queries) DeleteInterestLink(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, `DELETE FROM interest_links WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

type Suggestion struct {
	ID                      uuid.UUID
	Description             string
	LevelsCompletedSnapshot int32
	XPSnapshot              int32
	SubmittedAt             time.Time
}

func scanSuggestion(row scanner) (Suggestion, error) {
	var s Suggestion
	err := row.Scan(&s.ID, &s.Description, &s.LevelsCompletedSnapshot, &s.XPSnapshot, &s.SubmittedAt)
	return s, err
}

const suggestionColumns = `id, description, levels_completed_snapshot, xp_snapshot, submitted_at`

type CreateSuggestionParams struct {
	ID                      uuid.UUID
	Description             string
	LevelsCompletedSnapshot int32
	XPSnapshot              int32
	SubmittedAt             time.Time
}

// CreateSuggestion no recibe ni guarda ningún identificador de cuenta —
// suggestions es anónima por diseño (comentario de la migración 0003), así
// que ni siquiera este paquete conoce quién la mandó más allá del momento en
// que calcula el snapshot.
//
// Deliberadamente sin RETURNING: Postgres exige privilegio SELECT sobre las
// columnas devueltas incluso para un INSERT ... RETURNING (lo mismo que
// exigiría un SELECT normal), y usbi_app tiene únicamente INSERT en
// suggestions (00_roles_unificado.sql, matriz confirmada por el usuario) —
// verificado en vivo contra el contenedor: un INSERT ... RETURNING con ese
// rol falla con "permission denied for table suggestions", un INSERT plano
// no. Otorgar SELECT solo para poder leer la fila recién escrita abriría la
// puerta a que ese mismo rol reconstruya el buzón completo con un SELECT
// suelto más tarde, justo lo que el REVOKE busca impedir. submitted_at lo
// fija esta capa (NOW() en Go, no en el DEFAULT de la columna) para poder
// devolver una Suggestion completa sin leer de vuelta.
func (q *Queries) CreateSuggestion(ctx context.Context, arg CreateSuggestionParams) (Suggestion, error) {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO suggestions (id, description, levels_completed_snapshot, xp_snapshot, submitted_at)
VALUES ($1, $2, $3, $4, $5)
`, arg.ID, arg.Description, arg.LevelsCompletedSnapshot, arg.XPSnapshot, arg.SubmittedAt)
	if err != nil {
		return Suggestion{}, err
	}
	return Suggestion{
		ID:                      arg.ID,
		Description:             arg.Description,
		LevelsCompletedSnapshot: arg.LevelsCompletedSnapshot,
		XPSnapshot:              arg.XPSnapshot,
		SubmittedAt:             arg.SubmittedAt,
	}, nil
}

type ListSuggestionsParams struct {
	// Cursor es el ID de la última sugerencia ya vista; uuid.Nil pide la
	// primera página. Los ID son UUIDv7 (newID(), ordenados en el tiempo),
	// así que ORDER BY id DESC aproxima el mismo orden que el índice
	// suggestions_submitted_at_idx sin necesitar un cursor compuesto.
	Cursor   uuid.UUID
	PageSize int32
}

func (q *Queries) ListSuggestions(ctx context.Context, arg ListSuggestionsParams) ([]Suggestion, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT `+suggestionColumns+`
FROM suggestions
WHERE (id < $1 OR $1 = '00000000-0000-0000-0000-000000000000'::uuid)
ORDER BY id DESC
LIMIT $2
`, arg.Cursor, arg.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Suggestion
	for rows.Next() {
		item, err := scanSuggestion(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (q *Queries) DeleteSuggestion(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, `DELETE FROM suggestions WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
