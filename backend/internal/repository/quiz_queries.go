// quiz_queries.go es NUEVO en F8 (plan/04_Rediseno_identidad_gustos.md §2):
// acceso a datos del banco de preguntas de registro (registration_questions,
// registration_settings). internal/quiz es el único consumidor — orquesta
// aquí el guard de "mínimo 4 preguntas activas" dentro de una transacción,
// tal como pide el comentario de registration_questions en
// 0001_esquema_unificado.up.sql: la regla vive en Go, no en un CHECK/trigger.
//
// account_quiz_answers no tiene consultas aquí todavía: las respuestas se
// insertan en el registro (POST /auth/register/confirm) y se leen desde el
// panel de admin (GET /admin/accounts/{id}/quiz-answers) — ambos F9, que es
// quien las necesita primero.
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RegistrationQuestion struct {
	ID           uuid.UUID
	QuestionText string
	IsActive     bool
	DisplayOrder int16
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func scanRegistrationQuestion(row scanner) (RegistrationQuestion, error) {
	var q RegistrationQuestion
	err := row.Scan(&q.ID, &q.QuestionText, &q.IsActive, &q.DisplayOrder, &q.CreatedAt, &q.UpdatedAt)
	return q, err
}

const registrationQuestionColumns = `id, question_text, is_active, display_order, created_at, updated_at`

// ListRegistrationQuestions trae TODAS las preguntas (activas e inactivas) —
// la vista de administración del banco completo.
func (q *Queries) ListRegistrationQuestions(ctx context.Context) ([]RegistrationQuestion, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT `+registrationQuestionColumns+`
FROM registration_questions
ORDER BY display_order ASC, created_at ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []RegistrationQuestion
	for rows.Next() {
		item, err := scanRegistrationQuestion(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListActiveRegistrationQuestions alimenta el muestreo aleatorio del
// registro (internal/quiz.SelectQuestionsForRegistration): solo mira las
// activas, igual que el índice parcial registration_questions_active_idx.
func (q *Queries) ListActiveRegistrationQuestions(ctx context.Context) ([]RegistrationQuestion, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT `+registrationQuestionColumns+`
FROM registration_questions
WHERE is_active
ORDER BY display_order ASC, created_at ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []RegistrationQuestion
	for rows.Next() {
		item, err := scanRegistrationQuestion(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// CountActiveRegistrationQuestions es el guard de "mínimo 4 activas". Se
// llama SIEMPRE dentro de la misma transacción que el UPDATE/DELETE que
// podría bajar el conteo, con FOR UPDATE para que dos administradores
// desactivando preguntas a la vez no pasen ambos el guard por una
// condición de carrera.
func (q *Queries) CountActiveRegistrationQuestions(ctx context.Context) (int, error) {
	var count int
	err := q.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM registration_questions WHERE is_active FOR UPDATE
`).Scan(&count)
	return count, err
}

type CreateRegistrationQuestionParams struct {
	ID           uuid.UUID
	QuestionText string
	IsActive     bool
	DisplayOrder int16
}

func (q *Queries) CreateRegistrationQuestion(ctx context.Context, arg CreateRegistrationQuestionParams) (RegistrationQuestion, error) {
	row := q.db.QueryRowContext(ctx, `
INSERT INTO registration_questions (id, question_text, is_active, display_order)
VALUES ($1, $2, $3, $4)
RETURNING `+registrationQuestionColumns,
		arg.ID, arg.QuestionText, arg.IsActive, arg.DisplayOrder)
	return scanRegistrationQuestion(row)
}

type UpdateRegistrationQuestionParams struct {
	ID           uuid.UUID
	QuestionText string
	IsActive     bool
	DisplayOrder int16
}

func (q *Queries) UpdateRegistrationQuestion(ctx context.Context, arg UpdateRegistrationQuestionParams) (RegistrationQuestion, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE registration_questions
SET question_text = $2, is_active = $3, display_order = $4, updated_at = NOW()
WHERE id = $1
RETURNING `+registrationQuestionColumns,
		arg.ID, arg.QuestionText, arg.IsActive, arg.DisplayOrder)
	return scanRegistrationQuestion(row)
}

// GetRegistrationQuestionForUpdate lee una pregunta bajo lock — la usa el
// guard de DeleteRegistrationQuestion para saber si la fila a borrar está
// activa (y por tanto cuenta contra el mínimo) sin una segunda consulta sin
// lock que otro admin pudiera adelantar.
func (q *Queries) GetRegistrationQuestionForUpdate(ctx context.Context, id uuid.UUID) (RegistrationQuestion, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT `+registrationQuestionColumns+`
FROM registration_questions
WHERE id = $1
FOR UPDATE
`, id)
	return scanRegistrationQuestion(row)
}

func (q *Queries) DeleteRegistrationQuestion(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `DELETE FROM registration_questions WHERE id = $1`, id)
	return err
}

type RegistrationSettings struct {
	MaxQuestionsShown int16
	UpdatedAt         time.Time
}

func (q *Queries) GetRegistrationSettings(ctx context.Context) (RegistrationSettings, error) {
	var s RegistrationSettings
	err := q.db.QueryRowContext(ctx, `
SELECT max_questions_shown, updated_at FROM registration_settings WHERE id = 1
`).Scan(&s.MaxQuestionsShown, &s.UpdatedAt)
	return s, err
}

func (q *Queries) UpdateRegistrationSettings(ctx context.Context, maxQuestionsShown int16) (RegistrationSettings, error) {
	var s RegistrationSettings
	err := q.db.QueryRowContext(ctx, `
UPDATE registration_settings
SET max_questions_shown = $1, updated_at = NOW()
WHERE id = 1
RETURNING max_questions_shown, updated_at
`, maxQuestionsShown).Scan(&s.MaxQuestionsShown, &s.UpdatedAt)
	return s, err
}
