// Package quiz reemplaza el flujo de tutor por correo con un cuestionario de
// gustos no sensibles del que se derivan la credencial de login (nickname) y
// el password de una cuenta nueva — ver plan/04_Rediseno_identidad_gustos.md.
//
// Responsabilidades separadas (pedido explícito del usuario):
//   - bank.go (este archivo): tipos de respuesta y lógica compartida entre
//     PlayerService y AdminService — no toca base de datos por sí mismo.
//   - player_service.go: PlayerService, pool usbi_app (F3, 2026-09-09) —
//     muestreo de preguntas para el registro, sin CRUD.
//   - admin_service.go: AdminService, pool usbi_moderador (F3) — CRUD del
//     banco de preguntas y su Handler HTTP.
//   - credentials.go: generación de nickname/password a partir de las
//     respuestas. Sin HTTP, sin acceso al banco de preguntas — funciones
//     puras que internal/auth orquesta durante el registro. (Útil)
package quiz

import (
	"context"
	"errors"
	mathrand "math/rand"
	"time"

	"github.com/altair/usbi-anon-backend/internal/audit"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrValidation         = errors.New("validation error")
	ErrNotFound           = errors.New("not found")
	ErrForbidden          = errors.New("forbidden")
	ErrMinActiveQuestions = errors.New("would leave fewer than 4 active questions")
)

// minActiveQuestions/minMaxQuestionsShown/maxMaxQuestionsShown reflejan los
// CHECK del esquema (0001_esquema_unificado.up.sql): la regla de mínimo 4
// activas se valida aquí en Go, dentro de transacción, tal como documenta el
// comentario de registration_questions — un CHECK/trigger no puede contar
// filas de la misma tabla de forma segura bajo concurrencia sin el mismo
// FOR UPDATE que ya hacen CountActiveRegistrationQuestions/
// GetRegistrationQuestionForUpdate. (Útil)
const (
	minActiveQuestions   = 4
	minMaxQuestionsShown = 4
	maxMaxQuestionsShown = 10
	questionTextMaxLen   = 280
)

type QuestionResponse struct {
	ID           uuid.UUID `json:"id"`
	QuestionText string    `json:"question_text"`
	IsActive     bool      `json:"is_active"`
	DisplayOrder int16     `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type QuestionsResponse struct {
	Items []QuestionResponse `json:"items"`
}

type CreateQuestionRequest struct {
	QuestionText string `json:"question_text"`
	IsActive     bool   `json:"is_active"`
	DisplayOrder int16  `json:"display_order"`
}

type UpdateQuestionRequest struct {
	QuestionText string `json:"question_text"`
	IsActive     bool   `json:"is_active"`
	DisplayOrder int16  `json:"display_order"`
}

type SettingsResponse struct {
	MaxQuestionsShown int16     `json:"max_questions_shown"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type UpdateSettingsRequest struct {
	MaxQuestionsShown int16 `json:"max_questions_shown"`
}

// PublicQuestion es la proyección que ve quien se está registrando: sin
// is_active ni display_order, que son detalles de administración. (Útil)
type PublicQuestion struct {
	ID   uuid.UUID `json:"id"`
	Text string    `json:"text"`
}

type RegistrationQuestionsResponse struct {
	Questions         []PublicQuestion `json:"questions"`
	MaxQuestionsShown int16            `json:"max_questions_shown"`
}

// selectRandom hace el muestreo aleatorio puro que pide plan/04 §2: sin
// ponderar por display_order ni nada más, un shuffle completo del pool y se
// toman los primeros `max`. math/rand (no crypto/rand): esto no es un
// secreto que proteger, es solo variar qué preguntas ve cada registro. (Útil)
func selectRandom(pool []repository.RegistrationQuestion, max int) []repository.RegistrationQuestion {
	if max <= 0 || len(pool) == 0 {
		return nil
	}
	shuffled := make([]repository.RegistrationQuestion, len(pool))
	copy(shuffled, pool)
	rng := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	if max > len(shuffled) {
		max = len(shuffled)
	}
	return shuffled[:max]
}

func questionToResponse(q repository.RegistrationQuestion) QuestionResponse {
	return QuestionResponse{
		ID:           q.ID,
		QuestionText: q.QuestionText,
		IsActive:     q.IsActive,
		DisplayOrder: q.DisplayOrder,
		CreatedAt:    q.CreatedAt,
		UpdatedAt:    q.UpdatedAt,
	}
}

func settingsToResponse(s repository.RegistrationSettings) SettingsResponse {
	return SettingsResponse{MaxQuestionsShown: s.MaxQuestionsShown, UpdatedAt: s.UpdatedAt}
}

func questionAuditPayload(q QuestionResponse) map[string]any {
	return map[string]any{
		"id":            q.ID,
		"is_active":     q.IsActive,
		"display_order": q.DisplayOrder,
		// question_text se omite a propósito: no es dato personal, pero el
		// texto de la pregunta no aporta nada al rastro de auditoría que
		// is_active/display_order no den ya, y mantiene el payload chico. (Útil)
	}
}

func logAudit(ctx context.Context, repo *repository.Queries, actorID uuid.UUID, action, entityType string, entityID uuid.UUID, before, after any) error {
	return audit.Log(ctx, repo, audit.Entry{
		ActorID:    actorID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Before:     before,
		After:      after,
	})
}
