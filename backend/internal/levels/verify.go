// verify.go recalcula completed/score en el servidor a partir de las respuestas del cliente.icamente verificable sin reproducir la partida completa.
//
// memory y snakes_ladders se quedan permanentemente sin verificar (M1.4-B3):
// no hay forma de comprobarlas sin que el servidor genere la semilla y valide
// un registro completo de jugadas, un costo desproporcionado para dos
// plantillas. PlayerService.CompleteLevel acepta ahí el completed/score que
// reporta el cliente y lo marca como VerificationOnlineReported.
package levels

import (
	"encoding/json"
	"strings"
)

const passRatio = 0.6

type verifiedResult struct {
	completed bool
	score     int32
	maxScore  int32
}

// verifyAnswers intenta recalcular el resultado desde `answers` contra
// `content`. ok=false cuando la plantilla no es verificable todavía o
// answers/content no calzan con el shape esperado — el llamador debe caer
// entonces al comportamiento anterior (confiar en el cliente).
func verifyAnswers(templateType string, content, answers json.RawMessage) (result verifiedResult, ok bool) {
	if len(answers) == 0 {
		return verifiedResult{}, false
	}
	switch templateType {
	case "trivia":
		return verifyTriviaAnswers(content, answers)
	case "fake_news":
		return verifyFakeNewsAnswers(content, answers)
	case "crossword":
		return verifyCrosswordAnswers(content, answers)
	case "word_search":
		return verifyWordSearchAnswers(content, answers)
	case "puzzle":
		return verifyPuzzleAnswers(content, answers)
	default:
		return verifiedResult{}, false
	}
}

// verifyTriviaAnswers espera answers = {"selected_indices": [int, ...]},
// paralelo a las preguntas de content (mismo shape flexible — array plano o
// envuelto en {"questions": [...]} — que valida validateTriviaContent).
func verifyTriviaAnswers(content, answers json.RawMessage) (verifiedResult, bool) {
	type question struct {
		CorrectIndex int `json:"correct_index"`
	}
	type questionEnvelope struct {
		Questions []question `json:"questions"`
	}

	var questions []question
	if err := json.Unmarshal(content, &questions); err != nil || len(questions) == 0 {
		var envelope questionEnvelope
		if err := json.Unmarshal(content, &envelope); err != nil || len(envelope.Questions) == 0 {
			return verifiedResult{}, false
		}
		questions = envelope.Questions
	}

	var payload struct {
		SelectedIndices []int `json:"selected_indices"`
	}
	if err := json.Unmarshal(answers, &payload); err != nil {
		return verifiedResult{}, false
	}
	if len(payload.SelectedIndices) != len(questions) {
		return verifiedResult{}, false
	}

	var correct int32
	for i, q := range questions {
		if payload.SelectedIndices[i] == q.CorrectIndex {
			correct++
		}
	}
	return passRatioResult(correct, int32(len(questions))), true
}

// verifyFakeNewsAnswers espera answers = {"guesses": [bool, ...]}, paralelo a
// content.news — guesses[i] es lo que el jugador marcó como "es falsa".
func verifyFakeNewsAnswers(content, answers json.RawMessage) (verifiedResult, bool) {
	type item struct {
		IsFake *bool `json:"isFake"`
	}
	var payload struct {
		News []item `json:"news"`
	}
	if err := json.Unmarshal(content, &payload); err != nil || len(payload.News) == 0 {
		return verifiedResult{}, false
	}

	var answerPayload struct {
		Guesses []bool `json:"guesses"`
	}
	if err := json.Unmarshal(answers, &answerPayload); err != nil {
		return verifiedResult{}, false
	}
	if len(answerPayload.Guesses) != len(payload.News) {
		return verifiedResult{}, false
	}

	var correct int32
	for i, item := range payload.News {
		if item.IsFake == nil {
			return verifiedResult{}, false
		}
		if *item.IsFake == answerPayload.Guesses[i] {
			correct++
		}
	}
	return passRatioResult(correct, int32(len(payload.News))), true
}

func passRatioResult(correct, maxScore int32) verifiedResult {
	completed := maxScore > 0 && float64(correct)/float64(maxScore) >= passRatio
	return verifiedResult{completed: completed, score: correct, maxScore: maxScore}
}

// verifyCrosswordAnswers espera answers = {"solved_words": ["PALABRA", ...]}
// — las palabras que el motor cliente (CrosswordEngine.getSolvedWords)
// determinó como completamente correctas, no coordenadas de celda. Evita
// reconstruir en Go el mismo layout de rejilla que ya arma el motor en
// TypeScript: el criterio de M1.2 para esta plantilla es "todas o ninguna"
// (isFinished exige el 100%, no el 60% de trivia/fake_news), así que basta
// con el conjunto de palabras, no su posición exacta.
func verifyCrosswordAnswers(content, answers json.RawMessage) (verifiedResult, bool) {
	type word struct {
		Word string `json:"word"`
	}
	var payload struct {
		Words []word `json:"words"`
	}
	if err := json.Unmarshal(content, &payload); err != nil || len(payload.Words) == 0 {
		return verifiedResult{}, false
	}

	expected := make(map[string]struct{}, len(payload.Words))
	for _, w := range payload.Words {
		answer := normalizeCrosswordAnswer(w.Word)
		if answer == "" {
			return verifiedResult{}, false
		}
		expected[answer] = struct{}{}
	}

	var answerPayload struct {
		SolvedWords []string `json:"solved_words"`
	}
	if err := json.Unmarshal(answers, &answerPayload); err != nil {
		return verifiedResult{}, false
	}

	correct := countDistinctMatches(expected, answerPayload.SolvedWords, normalizeCrosswordAnswer)
	maxScore := int32(len(expected))
	return verifiedResult{completed: correct == maxScore, score: correct, maxScore: maxScore}, true
}

// verifyWordSearchAnswers espera answers = {"found_words": ["PALABRA", ...]}
// contra el conjunto completo de content.words. Limitación conocida: si una
// palabra del contenido no cupo en la rejilla en tiempo de juego (el motor ya
// descarta esas silenciosamente — WordSearchEngine.generateGrid), el jugador
// jamás podrá encontrarla y este cálculo nunca llegará a completed=true; es
// un defecto de autoría de contenido preexistente (validateWordSearchContent
// no comprueba que las palabras quepan), no algo que esta verificación deba
// enmascarar aceptando de vuelta lo que reporte el cliente.
func verifyWordSearchAnswers(content, answers json.RawMessage) (verifiedResult, bool) {
	var payload struct {
		Words []string `json:"words"`
	}
	if err := json.Unmarshal(content, &payload); err != nil || len(payload.Words) == 0 {
		return verifiedResult{}, false
	}

	expected := make(map[string]struct{}, len(payload.Words))
	for _, w := range payload.Words {
		normalized := normalizeWordSearchWord(w)
		if normalized != "" {
			expected[normalized] = struct{}{}
		}
	}
	if len(expected) == 0 {
		return verifiedResult{}, false
	}

	var answerPayload struct {
		FoundWords []string `json:"found_words"`
	}
	if err := json.Unmarshal(answers, &answerPayload); err != nil {
		return verifiedResult{}, false
	}

	correct := countDistinctMatches(expected, answerPayload.FoundWords, normalizeWordSearchWord)
	maxScore := int32(len(expected))
	return verifiedResult{completed: correct == maxScore, score: correct, maxScore: maxScore}, true
}

// verifyPuzzleAnswers espera answers = {"piece_order": [int, ...]} — el
// índice original de cada pieza en su posición final. No necesita la frase
// ni la semilla: el criterio de M1.2 (isFinished tal cual) es una permutación
// idéntica a la identidad, igual que PuzzleEngine.checkCompletion().
func verifyPuzzleAnswers(content, answers json.RawMessage) (verifiedResult, bool) {
	var payload struct {
		Pieces *int32 `json:"pieces,omitempty"`
	}
	if err := json.Unmarshal(content, &payload); err != nil {
		return verifiedResult{}, false
	}
	pieceCount := int32(3)
	if payload.Pieces != nil {
		pieceCount = *payload.Pieces
	}
	if pieceCount < 3 {
		return verifiedResult{}, false
	}

	var answerPayload struct {
		PieceOrder []int32 `json:"piece_order"`
	}
	if err := json.Unmarshal(answers, &answerPayload); err != nil {
		return verifiedResult{}, false
	}
	if int32(len(answerPayload.PieceOrder)) != pieceCount {
		return verifiedResult{}, false
	}

	var correct int32
	for i, idx := range answerPayload.PieceOrder {
		if idx == int32(i) {
			correct++
		}
	}
	return verifiedResult{completed: correct == pieceCount, score: correct, maxScore: pieceCount}, true
}

// countDistinctMatches cuenta cuántos elementos de `submitted` (normalizados
// con `normalize`) pertenecen a `expected`, contando cada uno como máximo una
// vez — evita que repetir la misma palabra infle el puntaje.
func countDistinctMatches(expected map[string]struct{}, submitted []string, normalize func(string) string) int32 {
	seen := make(map[string]struct{}, len(expected))
	var correct int32
	for _, raw := range submitted {
		normalized := normalize(raw)
		if _, ok := expected[normalized]; !ok {
			continue
		}
		if _, already := seen[normalized]; already {
			continue
		}
		seen[normalized] = struct{}{}
		correct++
	}
	return correct
}

// normalizeWordSearchWord replica WordSearchEngine.normalizeWord (frontend):
// mayúsculas, sin diacríticos, y a diferencia de normalizeCrosswordAnswer, Ñ
// se pliega a N en vez de conservarse — el motor de sopa de letras no
// distingue esa letra del resto del alfabeto latino.
func normalizeWordSearchWord(raw string) string {
	var builder []rune
	for _, letter := range strings.ToUpper(strings.TrimSpace(raw)) {
		switch letter {
		case 'Á', 'À', 'Ä', 'Â':
			letter = 'A'
		case 'É', 'È', 'Ë', 'Ê':
			letter = 'E'
		case 'Í', 'Ì', 'Ï', 'Î':
			letter = 'I'
		case 'Ó', 'Ò', 'Ö', 'Ô':
			letter = 'O'
		case 'Ú', 'Ù', 'Ü', 'Û':
			letter = 'U'
		case 'Ñ':
			letter = 'N'
		}
		if letter >= 'A' && letter <= 'Z' {
			builder = append(builder, letter)
		}
	}
	return string(builder)
}
