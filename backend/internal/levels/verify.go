// verify.go recalcula completed/score en el servidor a partir de las
// respuestas que manda el cliente, contra el contenido real del nivel — M1
// Fase B (D-03). Solo cubre las plantillas donde el resultado es
// determinísticamente verificable sin reproducir la partida completa.
//
// memory y snakes_ladders se quedan permanentemente sin verificar (M1.4-B3):
// no hay forma de comprobarlas sin que el servidor genere la semilla y valide
// un registro completo de jugadas, un costo desproporcionado para dos
// plantillas. PlayerService.CompleteLevel acepta ahí el completed/score que
// reporta el cliente y lo marca como VerificationOnlineReported.
package levels

import "encoding/json"

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
