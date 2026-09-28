// AdminService agrupa las operaciones CRUD del banco de preguntas (rutas 100% admin).
package quiz

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/httpjson"
	"github.com/altair/usbi-anon-backend/internal/httpproblem"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AdminService struct {
	repo *repository.Queries
}

func NewAdminService(repo *repository.Queries) *AdminService {
	return &AdminService{repo: repo}
}

func (s *AdminService) ListQuestions(ctx context.Context) (QuestionsResponse, error) {
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

func (s *AdminService) CreateQuestion(ctx context.Context, adminID uuid.UUID, req CreateQuestionRequest) (QuestionResponse, error) {
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

// UpdateQuestion es full-replace.
//
// Si desactiva una pregunta, usa el guard de "mínimo 4" en transacción.
func (s *AdminService) UpdateQuestion(ctx context.Context, adminID, id uuid.UUID, req UpdateQuestionRequest) (QuestionResponse, error) {
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

func (s *AdminService) DeleteQuestion(ctx context.Context, adminID, id uuid.UUID) error {
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

func (s *AdminService) GetSettings(ctx context.Context) (SettingsResponse, error) {
	settings, err := s.repo.GetRegistrationSettings(ctx)
	if err != nil {
		return SettingsResponse{}, err
	}
	return settingsToResponse(settings), nil
}

func (s *AdminService) UpdateSettings(ctx context.Context, adminID uuid.UUID, req UpdateSettingsRequest) (SettingsResponse, error) {
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

// ── Handler HTTP: solo las rutas admin del banco.

type Handler struct {
	svc *AdminService
}

func NewHandler(svc *AdminService) *Handler {
	return &Handler{svc: svc}
}

// canManageQuizBank restringe el banco de preguntas a admin.
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

// GetSettings maneja GET /admin/registration-settings.
func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromContext(r)
	if !ok || !canManageQuizBank(claims.Role) {
		httpproblem.WriteProblem(w, r, http.StatusForbidden, "forbidden", "Forbidden", "Only admins can view the registration settings")
		return
	}
	resp, err := h.svc.GetSettings(r.Context())
	if err != nil {
		httpproblem.WriteProblem(w, r, http.StatusInternalServerError, "internal-error", "Internal Server Error", "Could not get registration settings")
		return
	}
	httpproblem.WriteJSON(w, http.StatusOK, resp)
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
