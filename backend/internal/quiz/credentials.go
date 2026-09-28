// credentials.go deriva nickname y password de las respuestas del
// cuestionario de registro mediante funciones puras (plan/04_Rediseno_identidad_gustos.md §4).
//
// Dos generadores aleatorios DISTINTOS:
//   - Nickname: math/rand (identificador público, no secreto).
//   - Password: crypto/rand. Es el secreto real de la cuenta.
package quiz

import (
	cryptorand "crypto/rand"
	"errors"
	"math/big"
	mathrand "math/rand"
	"strings"
	"time"
	"unicode"
)

var (
	// ErrInsufficientAnswers: hacen falta al menos dos respuestas con algún
	// carácter [a-z0-9] tras normalizar (p. ej. "42" solo aporta dígitos,
	// pero cuenta) para construir un nickname con dos fragmentos distintos.
	ErrInsufficientAnswers = errors.New("quiz: insufficient answers to derive credentials")
	// ErrCandidateGenerationFailed: no fue posible generar un candidato sin
	// colisión tras el número máximo de reintentos. Con nicknameFragmentMaxLen
	// (4) × 4 dígitos aleatorios el espacio es grande, así que esto solo debería
	// ocurrir si existsFn está mal implementada (siempre true) o el banco de
	// nicknames está patológicamente lleno.
	ErrCandidateGenerationFailed = errors.New("quiz: could not generate a unique nickname candidate")
)

const (
	nicknameFragmentMaxLen = 4
	nicknameMinLen         = 6 // debe coincidir con el CHECK de accounts.nickname
	nicknameCandidateCount = 4
	maxCandidateAttempts   = 50

	passwordLength      = 12
	passwordFragMinLen  = 3
	passwordFragMaxLen  = 5
	passwordFillCharset = "ABCDEFGHJKMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789" // sin 0/O/1/l/I
)

// accentMap traduce vocales y ñ acentuadas a su forma sin acento. Manual en
// vez de golang.org/x/text/unicode/norm: esa dependencia no está en go.mod y
// no vale la pena sumarla solo para esto (ver plan/04 §4).
var accentMap = map[rune]rune{
	'á': 'a', 'é': 'e', 'í': 'i', 'ó': 'o', 'ú': 'u', 'ü': 'u', 'ñ': 'n',
	'Á': 'a', 'É': 'e', 'Í': 'i', 'Ó': 'o', 'Ú': 'u', 'Ü': 'u', 'Ñ': 'n',
}

// normalizeFragment reduce una respuesta a minúsculas [a-z0-9] puro, recorta
// a nicknameFragmentMaxLen runas. Es la misma normalización que exige el
// CHECK de accounts.nickname, aplicada por fragmento antes de combinarlos.
func normalizeFragment(answer string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(answer) {
		if mapped, ok := accentMap[r]; ok {
			r = mapped
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			if b.Len() >= nicknameFragmentMaxLen {
				break
			}
		}
	}
	return b.String()
}

// GenerateNicknameCandidates produce 4 candidatos de nickname a partir de las
// respuestas del cuestionario, verificando cada uno contra existsFn (una
// consulta a accounts.nickname inyectada por el llamador, para no acoplar
// este paquete a internal/repository) y regenerando SOLO el candidato que
// choque — no el lote completo.
func GenerateNicknameCandidates(answers []string, exists func(nickname string) (bool, error)) ([nicknameCandidateCount]string, error) {
	var candidates [nicknameCandidateCount]string

	fragments := make([]string, 0, len(answers))
	for _, a := range answers {
		if f := normalizeFragment(a); f != "" {
			fragments = append(fragments, f)
		}
	}
	if len(fragments) < 2 {
		return candidates, ErrInsufficientAnswers
	}

	rng := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	used := make(map[string]bool, nicknameCandidateCount)

	for i := 0; i < nicknameCandidateCount; i++ {
		candidate, err := generateNicknameCandidate(rng, fragments, func(n string) (bool, error) {
			if used[n] {
				return true, nil
			}
			return exists(n)
		})
		if err != nil {
			return candidates, err
		}
		used[candidate] = true
		candidates[i] = candidate
	}
	return candidates, nil
}

func generateNicknameCandidate(rng *mathrand.Rand, fragments []string, exists func(string) (bool, error)) (string, error) {
	for attempt := 0; attempt < maxCandidateAttempts; attempt++ {
		i, j := rng.Intn(len(fragments)), rng.Intn(len(fragments))
		if len(fragments) > 1 {
			for j == i {
				j = rng.Intn(len(fragments))
			}
		}
		candidate := fragments[i] + fragments[j]
		for k := 0; k < 3; k++ {
			candidate += string(rune('0' + rng.Intn(10)))
		}
		// Relleno de seguridad: dos fragmentos de 1 carácter (p. ej. dos
		// respuestas numéricas) + 3 dígitos suman 5, por debajo del mínimo
		// de 6 del CHECK. Se completa con más dígitos en vez de fallar.
		for len(candidate) < nicknameMinLen {
			candidate += string(rune('0' + rng.Intn(10)))
		}

		collides, err := exists(candidate)
		if err != nil {
			return "", err
		}
		if !collides {
			return candidate, nil
		}
	}
	return "", ErrCandidateGenerationFailed
}

// GeneratePassword produce un password de EXACTAMENTE 12 caracteres:
// un fragmento legible de una respuesta (3-5 caracteres, capitalizado) más
// relleno crypto/rand sobre un charset sin caracteres ambiguos, intercalados
// —no "fragmento + relleno" en bloque— para que el patrón no sea
// trivialmente adivinable conociendo el algoritmo (ver plan/04 §4). Se
// devuelve en texto plano UNA vez; el llamador (internal/auth) lo hashea con
// crypto.HashPassword antes de guardar y nunca lo persiste en claro.
func GeneratePassword(answers []string) (string, error) {
	fragment, err := passwordFragment(answers)
	if err != nil {
		return "", err
	}

	fillLen := passwordLength - len(fragment)
	filler := make([]byte, fillLen)
	for i := range filler {
		c, err := randomCharsetByte(passwordFillCharset)
		if err != nil {
			return "", err
		}
		filler[i] = c
	}

	positions, err := randomDistinctPositions(passwordLength, len(fragment))
	if err != nil {
		return "", err
	}

	out := make([]byte, passwordLength)
	isFragmentPos := make(map[int]bool, len(positions))
	for _, p := range positions {
		isFragmentPos[p] = true
	}
	fragIdx, fillIdx := 0, 0
	for pos := 0; pos < passwordLength; pos++ {
		if isFragmentPos[pos] {
			out[pos] = fragment[fragIdx]
			fragIdx++
		} else {
			out[pos] = filler[fillIdx]
			fillIdx++
		}
	}
	return string(out), nil
}

// passwordFragment elige la respuesta normalizada más larga (letras+dígitos,
// sin acentos) y la recorta a un largo aleatorio entre passwordFragMinLen y
// passwordFragMaxLen, capitalizando la primera letra. Si ninguna respuesta
// alcanza el mínimo, usa la más larga disponible tal cual — degradación
// razonable en vez de fallar el registro por una respuesta corta como "7".
func passwordFragment(answers []string) (string, error) {
	best := ""
	for _, a := range answers {
		norm := normalizeAlnumFull(a)
		if len(norm) > len(best) {
			best = norm
		}
	}
	if best == "" {
		return "", ErrInsufficientAnswers
	}

	fragLen := len(best)
	if fragLen > passwordFragMinLen {
		n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(passwordFragMaxLen-passwordFragMinLen+1)))
		if err != nil {
			return "", err
		}
		fragLen = passwordFragMinLen + int(n.Int64())
		if fragLen > len(best) {
			fragLen = len(best)
		}
	}

	frag := best[:fragLen]
	return capitalize(frag), nil
}

// normalizeAlnumFull es como normalizeFragment pero SIN el recorte a
// nicknameFragmentMaxLen: el password necesita hasta passwordFragMaxLen (5)
// caracteres, más que los 4 del nickname.
func normalizeAlnumFull(answer string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(answer) {
		if mapped, ok := accentMap[r]; ok {
			r = mapped
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func randomCharsetByte(charset string) (byte, error) {
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return 0, err
	}
	return charset[n.Int64()], nil
}

// randomDistinctPositions sortea `count` posiciones distintas en [0, total)
// vía crypto/rand — es lo que dispersa el fragmento legible entre el relleno
// en vez de dejarlo en un bloque contiguo.
func randomDistinctPositions(total, count int) ([]int, error) {
	pool := make([]int, total)
	for i := range pool {
		pool[i] = i
	}
	for i := 0; i < count; i++ {
		n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(total-i)))
		if err != nil {
			return nil, err
		}
		j := i + int(n.Int64())
		pool[i], pool[j] = pool[j], pool[i]
	}
	return pool[:count], nil
}
