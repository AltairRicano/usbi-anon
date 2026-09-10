// admin_read.go implementa la lectura y edición de security_incidents (B2,
// estado_proyecto.md 2026-09-09) — antes de esto la tabla solo tenía
// escritores (CreateIncident, en service.go). No hay Delete: ni una ruta lo
// expone, ni ningún rol tiene el privilegio, ni el esquema lo permite desde
// que la migración 0006 le agregó un trigger BEFORE DELETE — tres capas
// independientes, ninguna de las cuales depende de que las otras dos no
// fallen.
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

var ErrNotFound = errors.New("not found")

const (
	defaultPageSize = 20
	maxPageSize     = 50
)

// IncidentResponse representa un incidente para el panel de administración.
// EvidenceValid recalcula el HMAC al leer y lo compara contra evidence_hash
// — convierte el sello de un dato inerte en un control verificable en cada
// lectura, no solo en el momento de crear el incidente. (Relleno)
type IncidentResponse struct {
	ID                   uuid.UUID  `json:"id"`
	DetectedAt           time.Time  `json:"detected_at"`
	ReportedAt           *time.Time `json:"reported_at,omitempty"`
	Severity             string     `json:"severity"`
	AffectedScope        string     `json:"affected_scope"`
	Description          string     `json:"description"`
	ContainmentActions   string     `json:"containment_actions"`
	ResolvedAt           *time.Time `json:"resolved_at,omitempty"`
	ReportedToCutai      bool       `json:"reported_to_cutai"`
	NotifiedToCutaiAt    *time.Time `json:"notified_to_cutai_at,omitempty"`
	NotifiedToSubjectsAt *time.Time `json:"notified_to_subjects_at,omitempty"`
	EvidenceValid        bool       `json:"evidence_valid"`
}

// IncidentsPage es la respuesta paginada de GET /admin/security-incidents.
// Cursor opaco "<RFC3339Nano>_<uuid>", mismo patrón que auditlog.Page: el ID
// es uuid.New() (v4 aleatorio), ordenar por id no aproxima ningún orden
// temporal. (Relleno)
type IncidentsPage struct {
	Items      []IncidentResponse `json:"items"`
	NextCursor string              `json:"next_cursor,omitempty"`
}

// UpdateIncidentRequest es el cuerpo de PATCH /admin/security-incidents/{id}.
// Reemplaza el estado completo de los campos editables (mismo contrato que
// UpdateBadgeRequest/UpdateCategoryRequest) — no es un merge parcial.
// detected_at NO es editable: es el hecho histórico de cuándo se detectó el
// incidente, no parte de la narrativa que un PATCH puede corregir.
type UpdateIncidentRequest struct {
	Severity             string  `json:"severity"`
	AffectedScope        string  `json:"affected_scope"`
	Description          string  `json:"description"`
	ContainmentActions   string  `json:"containment_actions"`
	ReportedAt           *string `json:"reported_at,omitempty"`
	ResolvedAt           *string `json:"resolved_at,omitempty"`
	ReportedToCutai      bool    `json:"reported_to_cutai"`
	NotifiedToCutaiAt    *string `json:"notified_to_cutai_at,omitempty"`
	NotifiedToSubjectsAt *string `json:"notified_to_subjects_at,omitempty"`
}

func sealPayload(id uuid.UUID, severity, scope, description, containment string, detectedAt time.Time) []byte {
	return []byte(strings.Join([]string{
		id.String(), severity, scope, description, containment,
		detectedAt.Format(time.RFC3339),
	}, "|"))
}

func (s *Service) toResponse(row repository.SecurityIncident) IncidentResponse {
	payload := sealPayload(row.ID, row.Severity, row.AffectedScope, row.Description, row.ContainmentActions, row.DetectedAt)
	resp := IncidentResponse{
		ID:                 row.ID,
		DetectedAt:         row.DetectedAt,
		Severity:           row.Severity,
		AffectedScope:      row.AffectedScope,
		Description:        row.Description,
		ContainmentActions: row.ContainmentActions,
		ReportedToCutai:    row.ReportedToCutai,
		EvidenceValid:      crypto.VerifyHMAC(payload, row.EvidenceHash, s.hmacSecret),
	}
	if row.ReportedAt.Valid {
		t := row.ReportedAt.Time
		resp.ReportedAt = &t
	}
	if row.ResolvedAt.Valid {
		t := row.ResolvedAt.Time
		resp.ResolvedAt = &t
	}
	if row.NotifiedToCutaiAt.Valid {
		t := row.NotifiedToCutaiAt.Time
		resp.NotifiedToCutaiAt = &t
	}
	if row.NotifiedToSubjectsAt.Valid {
		t := row.NotifiedToSubjectsAt.Time
		resp.NotifiedToSubjectsAt = &t
	}
	return resp
}

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

// List pagina los incidentes de seguridad, más recientes primero.
func (s *Service) List(ctx context.Context, actor domain.JWTClaims, cursor string, pageSize int32) (IncidentsPage, error) {
	if actor.Role != domain.RoleAdmin {
		return IncidentsPage{}, ErrForbidden
	}
	cursorTime, cursorID, err := parseCursor(cursor)
	if err != nil {
		return IncidentsPage{}, err
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	rows, err := s.queries.ListSecurityIncidents(ctx, repository.ListSecurityIncidentsParams{
		CursorTime: sql.NullTime{Time: cursorTime, Valid: !cursorTime.IsZero()},
		CursorID:   uuid.NullUUID{UUID: cursorID, Valid: cursorID != uuid.Nil},
		PageSize:   pageSize + 1,
	})
	if err != nil {
		return IncidentsPage{}, err
	}

	page := IncidentsPage{}
	hasMore := int32(len(rows)) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}
	page.Items = make([]IncidentResponse, 0, len(rows))
	for _, row := range rows {
		page.Items = append(page.Items, s.toResponse(row))
	}
	if hasMore {
		last := rows[len(rows)-1]
		page.NextCursor = encodeCursor(last.DetectedAt, last.ID)
	}
	return page, nil
}

// Get devuelve un incidente por ID, con su evidence_valid recalculado.
func (s *Service) Get(ctx context.Context, actor domain.JWTClaims, id uuid.UUID) (IncidentResponse, error) {
	if actor.Role != domain.RoleAdmin {
		return IncidentResponse{}, ErrForbidden
	}
	row, err := s.queries.GetSecurityIncident(ctx, id)
	if err != nil {
		if repository.IsNoRows(err) {
			return IncidentResponse{}, ErrNotFound
		}
		return IncidentResponse{}, err
	}
	return s.toResponse(row), nil
}

// Update corrige la narrativa y/o las columnas de seguimiento de un
// incidente, resellando evidence_hash con el contenido nuevo (decisión D1:
// evidence_hash pasa de probar "esto nunca se tocó" a probar "esto coincide
// con la última versión guardada por la aplicación" — sigue detectando
// manipulación fuera de banda, que es el ataque que importa). La versión
// anterior completa queda en audit_log.before_state, que es append-only por
// trigger: ninguna corrección borra la historia.
func (s *Service) Update(ctx context.Context, actor domain.JWTClaims, id uuid.UUID, req UpdateIncidentRequest, ip, userAgent string) (IncidentResponse, error) {
	if actor.Role != domain.RoleAdmin {
		return IncidentResponse{}, ErrForbidden
	}

	severity := strings.ToLower(strings.TrimSpace(req.Severity))
	scope := strings.TrimSpace(req.AffectedScope)
	description := strings.TrimSpace(req.Description)
	containment := strings.TrimSpace(req.ContainmentActions)
	if _, ok := validSeverities[severity]; !ok {
		return IncidentResponse{}, ErrValidation
	}
	if scope == "" || description == "" || containment == "" {
		return IncidentResponse{}, ErrValidation
	}
	if len(scope) > 2000 || len(description) > 10000 || len(containment) > 10000 {
		return IncidentResponse{}, ErrValidation
	}

	reportedAt, err := parseOptionalTime(req.ReportedAt)
	if err != nil {
		return IncidentResponse{}, ErrValidation
	}
	resolvedAt, err := parseOptionalTime(req.ResolvedAt)
	if err != nil {
		return IncidentResponse{}, ErrValidation
	}
	notifiedToCutaiAt, err := parseOptionalTime(req.NotifiedToCutaiAt)
	if err != nil {
		return IncidentResponse{}, ErrValidation
	}
	notifiedToSubjectsAt, err := parseOptionalTime(req.NotifiedToSubjectsAt)
	if err != nil {
		return IncidentResponse{}, ErrValidation
	}

	tx, err := s.queries.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return IncidentResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.queries.WithTx(tx)

	before, err := qtx.GetSecurityIncident(ctx, id)
	if err != nil {
		if repository.IsNoRows(err) {
			return IncidentResponse{}, ErrNotFound
		}
		return IncidentResponse{}, err
	}

	evidenceHash := crypto.GenerateHMAC(
		sealPayload(id, severity, scope, description, containment, before.DetectedAt),
		s.hmacSecret,
	)

	after, err := qtx.UpdateSecurityIncident(ctx, repository.UpdateSecurityIncidentParams{
		ID:                   id,
		Severity:             severity,
		AffectedScope:        scope,
		Description:          description,
		ContainmentActions:   containment,
		ReportedAt:           reportedAt,
		ResolvedAt:           resolvedAt,
		ReportedToCutai:      req.ReportedToCutai,
		NotifiedToCutaiAt:    notifiedToCutaiAt,
		NotifiedToSubjectsAt: notifiedToSubjectsAt,
		EvidenceHash:         evidenceHash,
	})
	if err != nil {
		return IncidentResponse{}, fmt.Errorf("updating security incident: %w", err)
	}

	// before_state lleva la narrativa completa anterior — audit_log es
	// append-only, así que ninguna versión se pierde aunque el incidente se
	// resele (ver nota de paquete). Referirse a personas por UUID: el mismo
	// COMMENT ON COLUMN de la migración 0001 aplica aquí igual que en creación.
	if err := audit.Log(ctx, qtx, audit.Entry{
		ActorID:    actor.UserID,
		Action:     "security_incident.update",
		EntityType: "security_incident",
		EntityID:   id,
		Before: map[string]any{
			"severity": before.Severity, "affected_scope": before.AffectedScope,
			"description": before.Description, "containment_actions": before.ContainmentActions,
			"reported_to_cutai": before.ReportedToCutai,
		},
		After: map[string]any{
			"severity": after.Severity, "affected_scope": after.AffectedScope,
			"description": after.Description, "containment_actions": after.ContainmentActions,
			"reported_to_cutai": after.ReportedToCutai,
		},
		IP:        ip,
		UserAgent: userAgent,
	}); err != nil {
		return IncidentResponse{}, fmt.Errorf("logging security incident update: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return IncidentResponse{}, err
	}
	return s.toResponse(after), nil
}

func parseOptionalTime(raw *string) (sql.NullTime, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return sql.NullTime{}, nil
	}
	t, err := time.Parse(time.RFC3339, *raw)
	if err != nil {
		return sql.NullTime{}, err
	}
	return sql.NullTime{Time: t.UTC(), Valid: true}, nil
}
