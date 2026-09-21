// validation.go valida las respuestas del cuestionario en el backend.
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

// htmlTagRe replica la exclusión de HTML del esquema Zod del frontend.
var htmlTagRe = regexp.MustCompile(`(?i)<[a-z][\s\S]*>`)

var nicknameFormatRe = regexp.MustCompile(`^[a-z0-9]{6,20}$`)

// validateAnswerText valida el formato de la respuesta del usuario (sin control, JSON o HTML).
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
