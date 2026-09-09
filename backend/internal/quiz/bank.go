// Package quiz reemplaza el flujo de tutor por correo con un cuestionario de
// gustos no sensibles del que se derivan la credencial de login (nickname) y
// el password de una cuenta nueva — ver plan/04_Rediseno_identidad_gustos.md.
//
// Dos archivos, responsabilidades separadas (pedido explícito del usuario):
//   - bank.go (este archivo): CRUD del banco de preguntas de registro,
//     con su propio Handler HTTP para las rutas admin. Toca base de datos.
//   - credentials.go: generación de nickname/password a partir de las
//     respuestas. Sin HTTP, sin acceso al banco de preguntas — funciones
//     puras que internal/auth orquesta durante el registro. (Útil)
package quiz

import (
	"context"
	"database/sql"
	"errors"
	mathrand "math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/altair/usbi-anon-backend/internal/audit"
	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpjson"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/go-chi/chi/v5"
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

type Service struct {
	repo *repository.Queries
}

func NewService(repo *repository.Queries) *Service {
	return &Service{repo: repo}
}

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

func (s *Service) ListQuestions(ctx context.Context) (QuestionsResponse, error) {
	rows, err := s.repo.ListRegistrationQuestions(ctx)
	if err != nil {
		return QuestionsResponse{}, err
	}
	items := make([]QuestionResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, questionToResponse(row))
	}
	return QuestionsResponse{Items: items}, nil
}

func (s *Service) CreateQuestion(ctx context.Context, adminID uuid.UUID, req CreateQuestionRequest) (QuestionResponse, error) {
	text := strings.TrimSpace(req.QuestionText)
	if text == "" || len(text) > questionTextMaxLen {
		return QuestionResponse{}, ErrValidation
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return QuestionResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	question, err := qtx.CreateRegistrationQuestion(ctx, repository.CreateRegistrationQuestionParams{
		ID:           uuid.New(),
		QuestionText: text,
		IsActive:     req.IsActive,
		DisplayOrder: req.DisplayOrder,
	})
	if err != nil {
		return QuestionResponse{}, err
	}
	resp := questionToResponse(question)

	if err := logAudit(ctx, qtx, adminID, "quiz_question.create", "registration_question", question.ID, nil, questionAuditPayload(resp)); err != nil {
		return QuestionResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return QuestionResponse{}, err
	}
	return resp, nil
}

// UpdateQuestion es full-replace (mismo estilo que internal/levels), no
// JSON-Patch parcial: el caller siempre manda los tres campos.
//
// Cuando la actualización DESACTIVA una pregunta que estaba activa, corre en
// una transacción con el mismo guard de mínimo 4 que DeleteQuestion — el
// comentario de registration_questions en el esquema exige el guard tanto en
// UPDATE como en DELETE, no solo en DELETE. (Útil)
func (s *Service) UpdateQuestion(ctx context.Context, adminID, id uuid.UUID, req UpdateQuestionRequest) (QuestionResponse, error) {
	text := strings.TrimSpace(req.QuestionText)
	if id == uuid.Nil || text == "" || len(text) > questionTextMaxLen {
		return QuestionResponse{}, ErrValidation
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return QuestionResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	before, err := qtx.GetRegistrationQuestionForUpdate(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return QuestionResponse{}, ErrNotFound
		}
		return QuestionResponse{}, err
	}

	if before.IsActive && !req.IsActive {
		activeCount, err := qtx.CountActiveRegistrationQuestions(ctx)
		if err != nil {
			return QuestionResponse{}, err
		}
		if activeCount-1 < minActiveQuestions {
			return QuestionResponse{}, ErrMinActiveQuestions
		}
	}

	after, err := qtx.UpdateRegistrationQuestion(ctx, repository.UpdateRegistrationQuestionParams{
		ID:           id,
		QuestionText: text,
		IsActive:     req.IsActive,
		DisplayOrder: req.DisplayOrder,
	})
	if err != nil {
		return QuestionResponse{}, err
	}
	resp := questionToResponse(after)

	if err := logAudit(ctx, qtx, adminID, "quiz_question.update", "registration_question", id,
		questionAuditPayload(questionToResponse(before)), questionAuditPayload(resp)); err != nil {
		return QuestionResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return QuestionResponse{}, err
	}
	return resp, nil
}

func (s *Service) DeleteQuestion(ctx context.Context, adminID, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrValidation
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	before, err := qtx.GetRegistrationQuestionForUpdate(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if before.IsActive {
		activeCount, err := qtx.CountActiveRegistrationQuestions(ctx)
		if err != nil {
			return err
		}
		if activeCount-1 < minActiveQuestions {
			return ErrMinActiveQuestions
		}
	}

	if err := qtx.DeleteRegistrationQuestion(ctx, id); err != nil {
		return err
	}
	if err := logAudit(ctx, qtx, adminID, "quiz_question.delete", "registration_question", id,
		questionAuditPayload(questionToResponse(before)), nil); err != nil {
		return err
	}
	return tx.Commit()
}

// GetActiveQuestionByID valida un question_id recibido en
// POST /auth/register/answers (internal/auth, F9) y devuelve el texto a
// congelar en account_quiz_answers.question_text_snapshot. internal/auth
// pasa por aquí en vez de tocar internal/repository directo — el banco de
// preguntas es dominio de este paquete, no del repositorio genérico. (Útil)
func (s *Service) GetActiveQuestionByID(ctx context.Context, id uuid.UUID) (QuestionResponse, error) {
	question, err := s.repo.GetRegistrationQuestionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return QuestionResponse{}, ErrNotFound
		}
		return QuestionResponse{}, err
	}
	if !question.IsActive {
		return QuestionResponse{}, ErrNotFound
	}
	return questionToResponse(question), nil
}

func (s *Service) GetSettings(ctx context.Context) (SettingsResponse, error) {
	settings, err := s.repo.GetRegistrationSettings(ctx)
	if err != nil {
		return SettingsResponse{}, err
	}
	return settingsToResponse(settings), nil
}

func (s *Service) UpdateSettings(ctx context.Context, adminID uuid.UUID, req UpdateSettingsRequest) (SettingsResponse, error) {
	if req.MaxQuestionsShown < minMaxQuestionsShown || req.MaxQuestionsShown > maxMaxQuestionsShown {
		return SettingsResponse{}, ErrValidation
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return SettingsResponse{}, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	before, err := qtx.GetRegistrationSettings(ctx)
	if err != nil {
		return SettingsResponse{}, err
	}
	after, err := qtx.UpdateRegistrationSettings(ctx, req.MaxQuestionsShown)
	if err != nil {
		return SettingsResponse{}, err
	}
	resp := settingsToResponse(after)

	if err := logAudit(ctx, qtx, adminID, "quiz_settings.update", "registration_settings", uuid.Nil,
		map[string]any{"max_questions_shown": before.MaxQuestionsShown},
		map[string]any{"max_questions_shown": after.MaxQuestionsShown}); err != nil {
		return SettingsResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return SettingsResponse{}, err
	}
	return resp, nil
}

// SelectQuestionsForRegistration alimenta POST /auth/register/questions
// (internal/auth, F9): muestreo aleatorio PURO —no ponderado, no por
// display_order— de las preguntas activas, hasta max_questions_shown. Con
// más activas que el máximo configurado, el resto queda en reserva y rota
// entre registros porque cada llamada vuelve a sortear desde cero (decisión
// 8 del rediseño). (Útil)
func (s *Service) SelectQuestionsForRegistration(ctx context.Context) (RegistrationQuestionsResponse, error) {
	settings, err := s.repo.GetRegistrationSettings(ctx)
	if err != nil {
		return RegistrationQuestionsResponse{}, err
	}
	active, err := s.repo.ListActiveRegistrationQuestions(ctx)
	if err != nil {
		return RegistrationQuestionsResponse{}, err
	}

	shown := selectRandom(active, int(settings.MaxQuestionsShown))
	questions := make([]PublicQuestion, 0, len(shown))
	for _, q := range shown {
		questions = append(questions, PublicQuestion{ID: q.ID, Text: q.QuestionText})
	}
	return RegistrationQuestionsResponse{
		Questions:         questions,
		MaxQuestionsShown: settings.MaxQuestionsShown,
	}, nil
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

// ── Handler HTTP: solo las rutas admin del banco. El endpoint público
// POST /auth/register/questions vive en internal/auth (F9), que llama a
// Service.SelectQuestionsForRegistration directo — no pasa por este Handler. (Útil)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// canManageQuizBank restringe el banco de preguntas a admin, no a
// operator/director como el contenido de niveles (internal/levels): estas
// preguntas determinan cómo se genera la credencial de login de cada cuenta
// (ver credentials.go). (Útil)
func canManageQuizBank(role domain.UserRole) bool {
	return role == domain.RoleAdmin
}

func claimsFromContext(r *http.Request) (*domain.JWTClaims, bool) {
	claims, ok := r.Context().Value(domain.ClaimsKey).(*domain.JWTClaims)
	return claims, ok
}

func (h *Handler) ListQuestions(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageQuizBank(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can view the registration question bank")
		return
	}
	resp, err := h.svc.ListQuestions(r.Context())
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not list questions")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageQuizBank(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can edit the registration question bank")
		return
	}
	var req CreateQuestionRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}
	resp, err := h.svc.CreateQuestion(r.Context(), claims.UserID, req)
	if err != nil {
		writeServiceError(w, r, err, "Could not create question")
		return
	}
	httpproblem.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageQuizBank(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can edit the registration question bank")
		return
	}
	id, ok := parseURLUUID(w, r, "id")
	if !ok {
		return
	}
	var req UpdateQuestionRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}
	resp, err := h.svc.UpdateQuestion(r.Context(), claims.UserID, id, req)
	if err != nil {
		writeServiceError(w, r, err, "Could not update question")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageQuizBank(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can edit the registration question bank")
		return
	}
	id, ok := parseURLUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteQuestion(r.Context(), claims.UserID, id); err != nil {
		writeServiceError(w, r, err, "Could not delete question")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageQuizBank(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can edit the registration settings")
		return
	}
	var req UpdateSettingsRequest
	if err := httpjson.DecodeStrict(r.Body, &req); err != nil {
		httpproblem.WriteDecodeProblem(w, r, err)
		return
	}
	resp, err := h.svc.UpdateSettings(r.Context(), claims.UserID, req)
	if err != nil {
		writeServiceError(w, r, err, "Could not update registration settings")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
}

func parseURLUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusBadRequest, "bad-request", "Bad Request", key+" must be a valid UUID")
		return uuid.Nil, false
	}
	return parsed, true
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	switch {
	case errors.Is(err, ErrValidation):
		httpproblem.WriteProblem(w, r, http.StatusUnprocessableEntity, "validation-error", "Validation Error", err.Error())
	case errors.Is(err, ErrNotFound):
		httpproblem.WriteProblem(w, r, http.StatusNotFound, "not-found", "Not Found", "Resource not found")
	case errors.Is(err, ErrMinActiveQuestions):
		httpproblem.WriteProblem(w, r, http.StatusConflict, "min-active-questions", "Conflict", err.Error())
	case errors.Is(err, ErrForbidden):
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", err.Error())
	default:
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", fallback)
	}
}
