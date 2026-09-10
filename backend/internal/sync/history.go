// history.go implementa el historial de sincronización offline del propio
// jugador (B4, estado_proyecto.md 2026-09-09) — antes de esto sync_events
// solo tenía escritores (ProcessSync), sin ninguna forma de consultarlo.
//
// Corre sobre el mismo Service que ya usa el pool de jugador (usbi_app):
// no hace falta un paquete nuevo ni un split PlayerService/AdminService,
// porque no existe lado admin — "cada quien ve solo lo suyo, un admin
// también" ya lo impone la matriz de permisos (usbi_moderador no tiene
// ningún GRANT sobre sync_events), no un guard de rol en Go.
package sync

import (
	"context"
	"errors"
	"time"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

const (
	defaultHistoryPageSize = 20
	maxHistoryPageSize     = 50
)

var ErrValidation = errors.New("validation error")

// SyncEventSummary es la representación de un evento de sincronización para
// el propio jugador. Deliberadamente sin el payload JSONB (decisión D3): es
// el delta completo de progreso, crece sin techo, y no aporta nada a un
// listado — solo a la depuración puntual de un evento, que no es el caso de
// uso de esta pantalla.
type SyncEventSummary struct {
	ID              uuid.UUID  `json:"id"`
	DeviceID        uuid.UUID  `json:"device_id"`
	Status          string     `json:"status"`
	HMACValid       bool       `json:"hmac_valid"`
	ReceivedAt      time.Time  `json:"received_at"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty"`
	RejectionReason string     `json:"rejection_reason,omitempty"`
}

// SyncHistoryPage es la respuesta paginada de GET /sync/events — mismo
// shape que suggestions.Page/levels.LevelsPage.
type SyncHistoryPage struct {
	Items      []SyncEventSummary `json:"items"`
	NextCursor string             `json:"next_cursor,omitempty"`
}

func toSyncEventSummary(e repository.SyncEventSummary) SyncEventSummary {
	resp := SyncEventSummary{
		ID:         e.ID,
		DeviceID:   e.DeviceID,
		Status:     e.Status,
		HMACValid:  e.HmacValid,
		ReceivedAt: e.ReceivedAt,
	}
	if e.ProcessedAt.Valid {
		t := e.ProcessedAt.Time
		resp.ProcessedAt = &t
	}
	if e.RejectionReason.Valid {
		resp.RejectionReason = e.RejectionReason.String
	}
	return resp
}

// ListMySyncEvents lista el historial de sincronización del jugador
// autenticado, opcionalmente filtrado por device_id. cursor es el
// received_at (RFC3339) del último elemento ya visto; cero valor pide la
// primera página.
func (s *Service) ListMySyncEvents(ctx context.Context, userID uuid.UUID, deviceID uuid.UUID, cursor time.Time, pageSize int32) (SyncHistoryPage, error) {
	if userID == uuid.Nil {
		return SyncHistoryPage{}, ErrValidation
	}
	if pageSize <= 0 {
		pageSize = defaultHistoryPageSize
	}
	if pageSize > maxHistoryPageSize {
		pageSize = maxHistoryPageSize
	}

	params := repository.ListSyncEventsForUserParams{
		UserID:   userID,
		DeviceID: uuid.NullUUID{UUID: deviceID, Valid: deviceID != uuid.Nil},
		PageSize: pageSize + 1,
	}
	if !cursor.IsZero() {
		params.Cursor.Time = cursor
		params.Cursor.Valid = true
	}

	rows, err := s.queries.ListSyncEventsForUser(ctx, params)
	if err != nil {
		return SyncHistoryPage{}, err
	}

	page := SyncHistoryPage{}
	hasMore := len(rows) > int(pageSize)
	if hasMore {
		rows = rows[:pageSize]
	}
	page.Items = make([]SyncEventSummary, 0, len(rows))
	for _, row := range rows {
		page.Items = append(page.Items, toSyncEventSummary(row))
	}
	if hasMore {
		page.NextCursor = rows[len(rows)-1].ReceivedAt.Format(time.RFC3339Nano)
	}
	return page, nil
}
