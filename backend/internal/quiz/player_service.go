// PlayerService agrupa las operaciones de internal/quiz que corren con el
// pool de usbi_app (F3, 2026-09-09): muestreo de preguntas para el registro.
// Sin CRUD — usbi_app solo tiene SELECT sobre registration_questions/
// registration_settings (00_roles_unificado.sql), coherente con que esta
// mitad nunca escribe. internal/auth es el único consumidor: llama estos dos
// métodos directo durante el registro en 3 pasos, sin pasar por Handler. (Útil)
package quiz

import (
	"context"
	"database/sql"
	"errors"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

type PlayerService struct {
	repo *repository.Queries
}

func NewPlayerService(repo *repository.Queries) *PlayerService {
	return &PlayerService{repo: repo}
}

// GetActiveQuestionByID valida un question_id recibido en
// POST /auth/register/answers (internal/auth, F9) y devuelve el texto a
// congelar en account_quiz_answers.question_text_snapshot. internal/auth
// pasa por aquí en vez de tocar internal/repository directo — el banco de
// preguntas es dominio de este paquete, no del repositorio genérico. (Útil)
func (s *PlayerService) GetActiveQuestionByID(ctx context.Context, id uuid.UUID) (QuestionResponse, error) {
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

// SelectQuestionsForRegistration alimenta POST /auth/register/questions
// (internal/auth, F9): muestreo aleatorio PURO —no ponderado, no por
// display_order— de las preguntas activas, hasta max_questions_shown. Con
// más activas que el máximo configurado, el resto queda en reserva y rota
// entre registros porque cada llamada vuelve a sortear desde cero (decisión
// 8 del rediseño). (Útil)
func (s *PlayerService) SelectQuestionsForRegistration(ctx context.Context) (RegistrationQuestionsResponse, error) {
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
