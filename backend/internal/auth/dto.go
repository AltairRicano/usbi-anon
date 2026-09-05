package auth

import (
	"time"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/google/uuid"
)

// ── Registro en 3 pasos (plan/04_Rediseno_identidad_gustos.md §3) ─────────── (Útil)

// RegisterQuestionsResponse es la respuesta de POST /auth/register/questions. (Relleno)
type RegisterQuestionsResponse struct {
	Questions         []QuestionOption `json:"questions"`
	MaxQuestionsShown int16            `json:"max_questions_shown"`
}

type QuestionOption struct {
	ID   uuid.UUID `json:"id"`
	Text string    `json:"text"`
}

// AnswerInput es una respuesta individual dentro de RegisterAnswersRequest. (Relleno)
type AnswerInput struct {
	QuestionID uuid.UUID `json:"question_id"`
	AnswerText string    `json:"answer_text"`
}

// RegisterAnswersRequest es el cuerpo para POST /auth/register/answers. (Relleno)
type RegisterAnswersRequest struct {
	Answers              []AnswerInput `json:"answers"`
	IsAdult              bool          `json:"is_adult"`
	PrivacyNoticeVersion string        `json:"privacy_notice_version"`
}

// RegisterAnswersResponse carga el token firmado que confirm() debe
// devolver, más los 4 candidatos de nickname generados a partir de las
// respuestas. (Útil)
type RegisterAnswersResponse struct {
	RegistrationToken  string   `json:"registration_token"`
	NicknameCandidates []string `json:"nickname_candidates"`
}

// RegisterConfirmRequest es el cuerpo para POST /auth/register/confirm. (Relleno)
type RegisterConfirmRequest struct {
	RegistrationToken string `json:"registration_token"`
	ChosenNickname    string `json:"chosen_nickname"`
}

// RegisterConfirmResponse se muestra UNA sola vez: password en claro, jamás
// vuelto a exponer por ningún otro endpoint. (Útil)
type RegisterConfirmResponse struct {
	AccountID    uuid.UUID `json:"account_id"`
	Nickname     string    `json:"nickname"`
	Password     string    `json:"password"`
	DisplayAlias string    `json:"display_alias"`
}

// ── Login / sesión ──────────────────────────────────────────────────────── (Útil)

// LoginRequest es el cuerpo para POST /auth/login. (Relleno)
type LoginRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

// LoginResponse se devuelve tras un inicio de sesión o actualización exitosa. (Relleno)
type LoginResponse struct {
	AccessToken           string      `json:"access_token"`
	RefreshToken          string      `json:"refresh_token"`
	TokenType             string      `json:"token_type"`
	AccessTokenExpiresIn  int64       `json:"access_token_expires_in"`
	RefreshTokenExpiresAt time.Time   `json:"refresh_token_expires_at"`
	User                  domain.User `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ── Administración de cuentas ──────────────────────────────────────────── (Útil)

// AdminCreateAccountRequest es el cuerpo para POST /admin/accounts. Sin
// cuestionario de gustos: un admin crea otra cuenta de staff con
// nickname+password explícitos (§2 del rediseño). (Útil)
type AdminCreateAccountRequest struct {
	Nickname string          `json:"nickname"`
	Password string          `json:"password"`
	Role     domain.UserRole `json:"role"`
}

type AdminAccountResponse struct {
	ID           uuid.UUID       `json:"id"`
	Nickname     string          `json:"nickname"`
	Role         domain.UserRole `json:"role"`
	DisplayAlias string          `json:"display_alias"`
	CreatedAt    time.Time       `json:"created_at"`
}

type AdminQuizAnswerDTO struct {
	QuestionTextSnapshot string    `json:"question_text_snapshot"`
	AnswerText           string    `json:"answer_text"`
	CreatedAt            time.Time `json:"created_at"`
}

type AdminQuizAnswersResponse struct {
	Items []AdminQuizAnswerDTO `json:"items"`
}

type AdminResetPasswordResponse struct {
	NewPassword string `json:"new_password"`
}
