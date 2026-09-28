// Package auditlog implementa la lectura de audit_log para el panel de administración.
// internal/audit es la puerta de escritura (audit.Log); este paquete solo gestiona lecturas.
//
// Solo AdminService: no existe lado jugador. Cualquier consulta exitosa registra
// una entrada audit_log.read con los filtros usados (nunca con los resultados devueltos),
// al tratarse de una lectura privilegiada de un registro forense para trazar quién la realizó.
package auditlog

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/altair/usbi-anon-backend/internal/audit"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

var ErrValidation = errors.New("validation error")

const (
	defaultPageSize = 20
	maxPageSize     = 50
)

type AdminService struct {
	repo *repository.Queries
}

func NewAdminService(repo *repository.Queries) *AdminService {
	return &AdminService{repo: repo}
}

// parseCursor decodifica el cursor opaco "<RFC3339Nano>_<uuid>" que esta
// misma capa generó en la página anterior. Un cursor vacío pide la primera
// página.
func parseCursor(raw string) (time.Time, uuid.UUID, error) {
	if raw == "" {
		return time.Time{}, uuid.Nil, nil
	}
	idx := strings.LastIndexByte(raw, '_')
	if idx < 0 {
		return time.Time{}, uuid.Nil, ErrValidation
	}
	t, err := time.Parse(time.RFC3339Nano, raw[:idx])
	if err != nil {
		return time.Time{}, uuid.Nil, ErrValidation
	}
	id, err := uuid.Parse(raw[idx+1:])
	if err != nil {
		return time.Time{}, uuid.Nil, ErrValidation
	}
	return t, id, nil
}

func encodeCursor(t time.Time, id uuid.UUID) string {
	return t.Format(time.RFC3339Nano) + "_" + id.String()
}

// List filtra y pagina audit_log, y audita la propia consulta
// con los filtros usados —nunca con los resultados devueltos— en la misma
// transacción para que una falla al auditar no deje una lectura sin rastro.
func (s *AdminService) List(ctx context.Context, adminID uuid.UUID, f Filters) (Page, error) {
	cursorTime, cursorID, err := parseCursor(f.Cursor)
	if err != nil {
		return Page{}, err
	}

	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	params := repository.ListAuditLogParams{
		ActorAccountID: uuid.NullUUID{UUID: f.ActorAccountID, Valid: f.ActorAccountID != uuid.Nil},
		Action:         sql.NullString{String: f.Action, Valid: f.Action != ""},
		EntityType:     sql.NullString{String: f.EntityType, Valid: f.EntityType != ""},
		From:           sql.NullTime{Time: f.From, Valid: !f.From.IsZero()},
		To:             sql.NullTime{Time: f.To, Valid: !f.To.IsZero()},
		CursorTime:     sql.NullTime{Time: cursorTime, Valid: !cursorTime.IsZero()},
		CursorID:       uuid.NullUUID{UUID: cursorID, Valid: cursorID != uuid.Nil},
		PageSize:       pageSize + 1,
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return Page{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	rows, err := qtx.ListAuditLog(ctx, params)
	if err != nil {
		return Page{}, err
	}

	page := Page{}
	hasMore := int32(len(rows)) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}
	page.Items = make([]EntryResponse, 0, len(rows))
	for _, row := range rows {
		page.Items = append(page.Items, toResponse(row))
	}
	if hasMore {
		last := rows[len(rows)-1]
		page.NextCursor = encodeCursor(last.CreatedAt, last.ID)
	}

	if err := audit.Log(ctx, qtx, audit.Entry{
		ActorID:    adminID,
		Action:     "audit_log.read",
		EntityType: "audit_log",
		After:      auditReadFilters(f),
	}); err != nil {
		return Page{}, err
	}
	if err := tx.Commit(); err != nil {
		return Page{}, err
	}
	return page, nil
}

// auditReadFilters registra qué se consultó, nunca los resultados obtenidos.
func auditReadFilters(f Filters) map[string]any {
	filters := map[string]any{"page_size": f.PageSize}
	if f.ActorAccountID != uuid.Nil {
		filters["actor_account_id"] = f.ActorAccountID
	}
	if f.Action != "" {
		filters["action"] = f.Action
	}
	if f.EntityType != "" {
		filters["entity_type"] = f.EntityType
	}
	if !f.From.IsZero() {
		filters["from"] = f.From
	}
	if !f.To.IsZero() {
		filters["to"] = f.To
	}
	if f.Cursor != "" {
		filters["cursor"] = f.Cursor
	}
	return filters
}

// parsePageSize traduce el query param page_size con el mismo contrato que
// suggestions/sync: entero entre 1 y maxPageSize.
func parsePageSize(raw string) (int32, error) {
	if raw == "" {
		return defaultPageSize, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > maxPageSize {
		return 0, ErrValidation
	}
	return int32(n), nil
}
