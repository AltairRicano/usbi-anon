// Package audit centralises writes to admin_audit_log so every sensitive
// operation records No-Repudio evidence through a single, consistent path.
//
// Before this package the only caller was internal/levels; the most sensitive
// actions of all — approving/rejecting an ARCO request, aging a minor up, and
// bootstrapping the first admin — wrote nothing (audit finding A3).
package audit

import (
	"context"
	"encoding/json"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

// Entry is a single audit record. Before/After are marshalled to JSON; nil
// becomes an empty object. IP/UserAgent default to placeholders for
// backend-internal actions that have no HTTP request context.
type Entry struct {
	ActorID    uuid.UUID
	Action     string
	EntityType string
	EntityID   uuid.UUID
	Before     any
	After      any
	IP         string
	UserAgent  string
}

// Log appends an entry to admin_audit_log using the given (possibly
// transactional) repository. The append-only trigger on the table guarantees
// the row can never be updated or deleted afterwards.
func Log(ctx context.Context, repo *repository.Queries, e Entry) error {
	before, err := marshalState(e.Before)
	if err != nil {
		return err
	}
	after, err := marshalState(e.After)
	if err != nil {
		return err
	}
	ip := e.IP
	if ip == "" {
		ip = "0.0.0.0"
	}
	userAgent := e.UserAgent
	if userAgent == "" {
		userAgent = "backend-service"
	}
	return repo.LogAdminAudit(ctx, repository.LogAdminAuditParams{
		ID:          uuid.New(),
		ActorUserID: uuid.NullUUID{UUID: e.ActorID, Valid: e.ActorID != uuid.Nil},
		Action:      e.Action,
		EntityType:  e.EntityType,
		EntityID:    uuid.NullUUID{UUID: e.EntityID, Valid: e.EntityID != uuid.Nil},
		BeforeState: before,
		AfterState:  after,
		IpAddress:   ip,
		UserAgent:   userAgent,
	})
}

func marshalState(value any) (pqtype.NullRawMessage, error) {
	if value == nil {
		return pqtype.NullRawMessage{RawMessage: json.RawMessage(`{}`), Valid: true}, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return pqtype.NullRawMessage{}, err
	}
	return pqtype.NullRawMessage{RawMessage: data, Valid: true}, nil
}
