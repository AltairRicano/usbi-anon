// Package incidents implements the admin-facing security incident log required
// by the Documento de Seguridad (audit finding A5). Before this package the
// security_incidents table existed in the schema but had zero writers — the
// mandated incident ledger was inert.
package incidents

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/altair/usbi-anon-backend/internal/audit"
	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrForbidden  = errors.New("forbidden")
	ErrValidation = errors.New("validation error")
)

// validSeverities is the accepted set; also enforced by a CHECK constraint.
var validSeverities = map[string]struct{}{
	"low": {}, "medium": {}, "high": {}, "critical": {},
}

// CreateIncidentRequest is the admin-submitted body for POST /admin/security-incidents.
type CreateIncidentRequest struct {
	Severity           string `json:"severity"` // low|medium|high|critical
	AffectedScope      string `json:"affected_scope"`
	Description        string `json:"description"`
	ContainmentActions string `json:"containment_actions"`
	DetectedAt         string `json:"detected_at,omitempty"` // RFC3339; defaults to now
	ReportedToCutai    bool   `json:"reported_to_cutai,omitempty"`
}

// CreateIncidentResponse is returned on success.
type CreateIncidentResponse struct {
	ID       uuid.UUID `json:"id"`
	Severity string    `json:"severity"`
	Message  string    `json:"message"`
}

type Service struct {
	queries    *repository.Queries
	hmacSecret []byte
}

func NewService(q *repository.Queries, hmacSecret []byte) *Service {
	if len(hmacSecret) == 0 {
		panic("incidents.Service: hmacSecret must not be empty")
	}
	return &Service{queries: q, hmacSecret: hmacSecret}
}

// CreateIncident records a security incident (admins/directors only), sealing it
// with an HMAC evidence hash for No-Repudio and writing an admin_audit_log entry.
func (s *Service) CreateIncident(ctx context.Context, actor domain.JWTClaims, req CreateIncidentRequest, ip, userAgent string) (CreateIncidentResponse, error) {
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleDirector {
		return CreateIncidentResponse{}, ErrForbidden
	}

	severity := strings.ToLower(strings.TrimSpace(req.Severity))
	scope := strings.TrimSpace(req.AffectedScope)
	description := strings.TrimSpace(req.Description)
	containment := strings.TrimSpace(req.ContainmentActions)
	if _, ok := validSeverities[severity]; !ok {
		return CreateIncidentResponse{}, ErrValidation
	}
	if scope == "" || description == "" || containment == "" {
		return CreateIncidentResponse{}, ErrValidation
	}
	if len(scope) > 2000 || len(description) > 10000 || len(containment) > 10000 {
		return CreateIncidentResponse{}, ErrValidation
	}

	detectedAt := time.Now().UTC()
	if strings.TrimSpace(req.DetectedAt) != "" {
		parsed, err := time.Parse(time.RFC3339, req.DetectedAt)
		if err != nil {
			return CreateIncidentResponse{}, ErrValidation
		}
		detectedAt = parsed.UTC()
		if detectedAt.After(time.Now().UTC()) {
			return CreateIncidentResponse{}, ErrValidation
		}
	}

	incidentID := uuid.New()
	sealPayload := []byte(strings.Join([]string{
		incidentID.String(), severity, scope, description, containment,
		detectedAt.Format(time.RFC3339),
	}, "|"))
	evidenceHash := crypto.GenerateHMAC(sealPayload, s.hmacSecret)

	tx, err := s.queries.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return CreateIncidentResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.queries.WithTx(tx)

	if err := qtx.InsertSecurityIncident(ctx, repository.InsertSecurityIncidentParams{
		ID:                 incidentID,
		DetectedAt:         detectedAt,
		Severity:           severity,
		AffectedScope:      scope,
		Description:        description,
		ContainmentActions: containment,
		ReportedToCutai:    req.ReportedToCutai,
		EvidenceHash:       evidenceHash,
	}); err != nil {
		return CreateIncidentResponse{}, fmt.Errorf("inserting security incident: %w", err)
	}
	if err := audit.Log(ctx, qtx, audit.Entry{
		ActorID:    actor.UserID,
		Action:     "security_incident.create",
		EntityType: "security_incident",
		EntityID:   incidentID,
		After:      map[string]any{"severity": severity, "affected_scope": scope, "reported_to_cutai": req.ReportedToCutai},
		IP:         ip,
		UserAgent:  userAgent,
	}); err != nil {
		return CreateIncidentResponse{}, fmt.Errorf("logging security incident: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return CreateIncidentResponse{}, err
	}

	return CreateIncidentResponse{ID: incidentID, Severity: severity, Message: "Security incident recorded"}, nil
}
