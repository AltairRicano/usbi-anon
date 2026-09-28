// quiz_queries.go es NUEVO en F8 (plan/04_Rediseno_identidad_gustos.md §2):
// acceso a datos del banco de preguntas de registro (registration_questions,
// registration_settings). internal/quiz es el único consumidor — orquesta
// aquí el guard de "mínimo 4 preguntas activas" dentro de una transacción,
// tal como pide el comentario de registration_questions en
// 0001_esquema_unificado.up.sql: la regla vive en Go, no en un CHECK/trigger.
//
// account_quiz_answers gana sus consultas en F9, que es quien primero las
// necesita: se insertan en el registro (POST /auth/register/confirm) y se
// leen desde el panel de admin (GET /admin/accounts/{id}/quiz-answers).
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
	// Postgres rechaza FOR UPDATE combinado directo con una función de
	// agregación ("FOR UPDATE is not allowed with aggregate functions") — el
	// lock de fila va en la subconsulta, el COUNT(*) en la externa, que ya no
	// lleva FOR UPDATE.
	err := q.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM (
    SELECT id FROM registration_questions WHERE is_active FOR UPDATE
) locked_active
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

// GetRegistrationQuestionByID es una lectura simple, sin lock — a diferencia
// de GetRegistrationQuestionForUpdate (guard de mínimo 4 en UPDATE/DELETE),
// esta la usa el registro (POST /auth/register/answers) solo para validar
// que question_id existe y está activa, y para congelar
// question_text_snapshot.
func (q *Queries) GetRegistrationQuestionByID(ctx context.Context, id uuid.UUID) (RegistrationQuestion, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT `+registrationQuestionColumns+`
FROM registration_questions
WHERE id = $1
`, id)
	return scanRegistrationQuestion(row)
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

// AccountQuizAnswer es una respuesta ya congelada — question_text_snapshot
// sobrevive aunque la pregunta original se edite o se borre (question_id
// pasa a NULL vía SET NULL, ver 0001_esquema_unificado.up.sql).
type AccountQuizAnswer struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	QuestionID           uuid.NullUUID
	QuestionTextSnapshot string
	AnswerText           string
	CreatedAt            time.Time
}

type InsertAccountQuizAnswerParams struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	QuestionID           uuid.UUID
	QuestionTextSnapshot string
	AnswerText           string
}

// InsertAccountQuizAnswer se llama una vez por respuesta dentro de la misma
// transacción que CreateAccount (POST /auth/register/confirm) — nunca
// suelta: una cuenta sin al menos sus respuestas persistidas no podría
// recuperarse nunca (§1 decisión 4).
func (q *Queries) InsertAccountQuizAnswer(ctx context.Context, arg InsertAccountQuizAnswerParams) error {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO account_quiz_answers (id, user_id, question_id, question_text_snapshot, answer_text)
VALUES ($1, $2, $3, $4, $5)
`, arg.ID, arg.UserID, arg.QuestionID, arg.QuestionTextSnapshot, arg.AnswerText)
	return err
}

// ListAccountQuizAnswers alimenta GET /admin/accounts/{id}/quiz-answers — el
// único endpoint que expone estas respuestas, protegido por rol admin y
// auditado por el llamador (comentario de account_quiz_answers en el
// esquema).
func (q *Queries) ListAccountQuizAnswers(ctx context.Context, accountID uuid.UUID) ([]AccountQuizAnswer, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT id, user_id, question_id, question_text_snapshot, answer_text, created_at
FROM account_quiz_answers
WHERE user_id = $1
ORDER BY created_at ASC
`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AccountQuizAnswer
	for rows.Next() {
		var a AccountQuizAnswer
		if err := rows.Scan(&a.ID, &a.UserID, &a.QuestionID, &a.QuestionTextSnapshot, &a.AnswerText, &a.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}
