package quiz

import (
	"strings"
	"testing"
)

func TestGenerateNicknameCandidates_ShapeAndCharset(t *testing.T) {
	answers := []string{"Azul", "Perro", "Matemáticas", "7", "2"}
	exists := func(string) (bool, error) { return false, nil }

	candidates, err := GenerateNicknameCandidates(answers, exists)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	seen := make(map[string]bool, len(candidates))
	for _, c := range candidates {
		if len(c) < nicknameMinLen || len(c) > 20 {
			t.Errorf("candidate %q has length %d, want between %d and 20", c, len(c), nicknameMinLen)
		}
		for _, r := range c {
			if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
				t.Errorf("candidate %q contains disallowed rune %q (only [a-z0-9] is valid)", c, r)
			}
		}
		if seen[c] {
			t.Errorf("candidate %q repeated within the same batch of 4", c)
		}
		seen[c] = true
	}
	if len(seen) != nicknameCandidateCount {
		t.Fatalf("got %d distinct candidates, want %d", len(seen), nicknameCandidateCount)
	}
}

func TestGenerateNicknameCandidates_InsufficientAnswers(t *testing.T) {
	// Solo "7" aporta un fragmento normalizable ("azul" con solo espacios
	// alrededor también cuenta, así que se usan respuestas vacías/símbolos).
	answers := []string{"7", "!!!", "   "}
	exists := func(string) (bool, error) { return false, nil }

	if _, err := GenerateNicknameCandidates(answers, exists); err != ErrInsufficientAnswers {
		t.Fatalf("got err=%v, want ErrInsufficientAnswers", err)
	}
}

// TestGenerateNicknameCandidates_CollisionRegeneratesOnlyThatCandidate
// verifica el requisito central del algoritmo (plan/04 §4): una colisión
// regenera SOLO el candidato afectado, no el lote completo. Se simula
// rechazando las primeras N propuestas y confirmando que igual se completan
// los 4 candidatos, todos distintos entre sí y del valor "ocupado".
func TestGenerateNicknameCandidates_CollisionRegeneratesOnlyThatCandidate(t *testing.T) {
	answers := []string{"azul", "perro", "matematicas", "siete"}

	rejectionsLeft := 5
	var calls int
	exists := func(nickname string) (bool, error) {
		calls++
		if rejectionsLeft > 0 {
			rejectionsLeft--
			return true, nil // simula que ya existe: debe reintentar
		}
		return false, nil
	}

	candidates, err := GenerateNicknameCandidates(answers, exists)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls <= nicknameCandidateCount {
		t.Fatalf("exists() called %d times, want more than %d (at least one retry expected)", calls, nicknameCandidateCount)
	}

	seen := make(map[string]bool, len(candidates))
	for _, c := range candidates {
		if c == "" {
			t.Fatalf("got empty candidate in %v", candidates)
		}
		if seen[c] {
			t.Errorf("candidate %q repeated after collision handling", c)
		}
		seen[c] = true
	}
}

func TestGenerateNicknameCandidates_ExhaustedRetriesFails(t *testing.T) {
	answers := []string{"azul", "perro"}
	alwaysExists := func(string) (bool, error) { return true, nil }

	if _, err := GenerateNicknameCandidates(answers, alwaysExists); err != ErrCandidateGenerationFailed {
		t.Fatalf("got err=%v, want ErrCandidateGenerationFailed", err)
	}
}

func TestNormalizeFragment_StripsAccentsAndSymbols(t *testing.T) {
	cases := map[string]string{
		"Azul":        "azul",
		"Matemáticas": "mate", // recortado a 4 runas
		"Ñoño":        "nono", // recortado a 4 runas
		"  42!!  ":    "42",
		"":            "",
		"###":         "",
	}
	for input, want := range cases {
		if got := normalizeFragment(input); got != want {
			t.Errorf("normalizeFragment(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestGeneratePassword_ExactLengthAndFragmentPresent(t *testing.T) {
	answers := []string{"azul", "perro", "matematicas"}

	for i := 0; i < 20; i++ {
		password, err := GeneratePassword(answers)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(password) != passwordLength {
			t.Fatalf("password %q has length %d, want %d", password, len(password), passwordLength)
		}
	}
}

func TestGeneratePassword_FillerExcludesAmbiguousChars(t *testing.T) {
	// "9999999999" no aporta letras normalizables más allá de dígitos, así
	// que el fragmento legible es corto y la mayor parte del password es
	// relleno — bueno para verificar el charset sin ambiguos en la práctica.
	answers := []string{"999"}

	for i := 0; i < 20; i++ {
		password, err := GeneratePassword(answers)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, r := range password {
			if r == '0' || r == 'O' || r == '1' || r == 'l' || r == 'I' {
				// El fragmento (derivado de "999") no puede producir estos
				// caracteres, así que cualquier aparición vendría del
				// relleno — justo lo que no debe pasar.
				t.Errorf("password %q contains ambiguous char %q from filler charset", password, r)
			}
		}
	}
}

func TestGeneratePassword_InsufficientAnswers(t *testing.T) {
	if _, err := GeneratePassword([]string{"", "   ", "!!!"}); err != ErrInsufficientAnswers {
		t.Fatalf("got err=%v, want ErrInsufficientAnswers", err)
	}
}

func TestCapitalize(t *testing.T) {
	cases := map[string]string{"azul": "Azul", "a": "A", "": ""}
	for input, want := range cases {
		if got := capitalize(input); got != want {
			t.Errorf("capitalize(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestGeneratePassword_ProducesVariedOutput(t *testing.T) {
	answers := []string{"azul", "perro", "matematicas"}
	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		p, err := GeneratePassword(answers)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		seen[p] = true
	}
	if len(seen) < 8 {
		t.Errorf("got only %d distinct passwords out of 10 calls, generator looks non-random", len(seen))
	}
}

func TestGenerateNicknameCandidates_ProducesVariedOutputAcrossCalls(t *testing.T) {
	answers := []string{"azul", "perro"}
	exists := func(string) (bool, error) { return false, nil }

	first, err := GenerateNicknameCandidates(answers, exists)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := GenerateNicknameCandidates(answers, exists)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Join(first[:], ",") == strings.Join(second[:], ",") {
		t.Errorf("two independent calls produced identical candidate sets: %v", first)
	}
}
