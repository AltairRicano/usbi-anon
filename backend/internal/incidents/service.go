// Package incidents implementa el registro y consulta de incidentes de seguridad para administradores,
// sellándolos criptográficamente para evidencia de No-Repudio.
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

// validSeverities es el conjunto aceptado; también se hace cumplir con una restricción CHECK.
var validSeverities = map[string]struct{}{
	"low": {}, "medium": {}, "high": {}, "critical": {},
}

// CreateIncidentRequest es el cuerpo enviado por el admin para POST /admin/security-incidents.
type CreateIncidentRequest struct {
	Severity           string `json:"severity"` // low|medium|high|critical
	AffectedScope      string `json:"affected_scope"`
	Description        string `json:"description"`
	ContainmentActions string `json:"containment_actions"`
	DetectedAt         string `json:"detected_at,omitempty"` // RFC3339; por defecto es la fecha/hora actual
	ReportedToCutai    bool   `json:"reported_to_cutai,omitempty"`
}

// CreateIncidentResponse se devuelve en caso de éxito.
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

// CreateIncident registra un incidente de seguridad (solo admins), sellándolo
// con un hash de evidencia HMAC para No-Repudio y escribiendo una entrada en audit_log.
func (s *Service) CreateIncident(ctx context.Context, actor domain.JWTClaims, req CreateIncidentRequest, ip, userAgent string) (CreateIncidentResponse, error) {
	if actor.Role != domain.RoleAdmin {
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
