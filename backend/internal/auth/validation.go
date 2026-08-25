// validation.go valida las respuestas del cuestionario en el backend —
// Zod en el frontend nunca es la única barrera (plan/04 §5): un cliente que
// se salte el frontend igual debe chocar con estas mismas reglas aquí.
package auth

import (
	"errors"
	"regexp"
	"strings"
)

const (
	answerTextMaxLen    = 200
	adminPasswordMinLen = 8
)

// htmlTagRe replica la exclusión de HTML del esquema Zod del frontend
// (plan/04 §5): cualquier "<letra" seguido de cualquier cosa y un ">".
var htmlTagRe = regexp.MustCompile(`(?i)<[a-z][\s\S]*>`)

var nicknameFormatRe = regexp.MustCompile(`^[a-z0-9]{6,20}$`)

// validateAnswerText replica en Go las tres reglas del esquema Zod
// compartido en plan/04 §5: sin caracteres de control, sin apariencia de
// JSON, sin HTML. answer_text es el único campo de texto libre escrito por
// una persona usuaria en toda la base (comentario de account_quiz_answers en
// el esquema), así que es el único que necesita esta validación.
func validateAnswerText(text string) error {
	if text == "" {
		return errors.New("answer_text is required")
	}
	if len(text) > answerTextMaxLen {
		return errors.New("answer_text must be at most 200 characters")
	}
	for _, r := range text {
		if r < 0x20 {
			return errors.New("answer_text must not contain control characters")
		}
	}
	if strings.HasPrefix(text, "[") || strings.HasPrefix(text, "{") {
		return errors.New("answer_text must not look like JSON")
	}
	if htmlTagRe.MatchString(text) {
		return errors.New("answer_text must not contain HTML")
	}
	return nil
}

func normalizeNickname(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func nicknameFormatValid(v string) bool {
	return nicknameFormatRe.MatchString(v)
}

func validateLogin(req LoginRequest) error {
	var errs []string
	if strings.TrimSpace(req.Nickname) == "" {
		errs = append(errs, "nickname is required")
	}
	if req.Password == "" {
		errs = append(errs, "password is required")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}
