package levels

import (
	"encoding/json"
	"testing"
)

func TestVerifyTriviaAnswers(t *testing.T) {
	content := json.RawMessage(`{"questions":[
		{"question":"Q1","options":["A","B"],"correct_index":0},
		{"question":"Q2","options":["A","B"],"correct_index":1},
		{"question":"Q3","options":["A","B"],"correct_index":1}
	]}`)

	tests := []struct {
		name          string
		answers       json.RawMessage
		wantOk        bool
		wantCompleted bool
		wantScore     int32
	}{
		{
			name:          "all correct is completed",
			answers:       json.RawMessage(`{"selected_indices":[0,1,1]}`),
			wantOk:        true,
			wantCompleted: true,
			wantScore:     3,
		},
		{
			name:          "2 of 3 (66%) meets 60% pass ratio",
			answers:       json.RawMessage(`{"selected_indices":[0,1,0]}`),
			wantOk:        true,
			wantCompleted: true,
			wantScore:     2,
		},
		{
			name:          "1 of 3 (33%) does not meet pass ratio",
			answers:       json.RawMessage(`{"selected_indices":[1,1,0]}`),
			wantOk:        true,
			wantCompleted: false,
			wantScore:     1,
		},
		{
			name:    "wrong number of answers is not verifiable",
			answers: json.RawMessage(`{"selected_indices":[0,1]}`),
			wantOk:  false,
		},
		{
			name:    "malformed answers is not verifiable",
			answers: json.RawMessage(`{"selected_indices":"not-an-array"}`),
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := verifyTriviaAnswers(content, tt.answers)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if !ok {
				return
			}
			if result.completed != tt.wantCompleted {
				t.Errorf("completed = %v, want %v", result.completed, tt.wantCompleted)
			}
			if result.score != tt.wantScore {
				t.Errorf("score = %v, want %v", result.score, tt.wantScore)
			}
			if result.maxScore != 3 {
				t.Errorf("maxScore = %v, want 3", result.maxScore)
			}
		})
	}

	t.Run("flat array content (no envelope) also works", func(t *testing.T) {
		flatContent := json.RawMessage(`[{"question":"Q1","options":["A","B"],"correct_index":0}]`)
		result, ok := verifyTriviaAnswers(flatContent, json.RawMessage(`{"selected_indices":[0]}`))
		if !ok {
			t.Fatal("expected ok=true for flat array content")
		}
		if !result.completed || result.score != 1 || result.maxScore != 1 {
			t.Errorf("unexpected result: %+v", result)
		}
	})
}

func TestVerifyFakeNewsAnswers(t *testing.T) {
	content := json.RawMessage(`{"news":[
		{"title":"N1","content":"C1","isFake":true,"reference":"R1"},
		{"title":"N2","content":"C2","isFake":false,"reference":"R2"},
		{"title":"N3","content":"C3","isFake":false,"reference":"R3"}
	]}`)

	tests := []struct {
		name          string
		answers       json.RawMessage
		wantOk        bool
		wantCompleted bool
		wantScore     int32
	}{
		{
			name:          "all correct is completed",
			answers:       json.RawMessage(`{"guesses":[true,false,false]}`),
			wantOk:        true,
			wantCompleted: true,
			wantScore:     3,
		},
		{
			name:          "1 of 3 does not meet pass ratio",
			answers:       json.RawMessage(`{"guesses":[false,false,true]}`),
			wantOk:        true,
			wantCompleted: false,
			wantScore:     1,
		},
		{
			name:    "wrong number of guesses is not verifiable",
			answers: json.RawMessage(`{"guesses":[true]}`),
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := verifyFakeNewsAnswers(content, tt.answers)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if !ok {
				return
			}
			if result.completed != tt.wantCompleted {
				t.Errorf("completed = %v, want %v", result.completed, tt.wantCompleted)
			}
			if result.score != tt.wantScore {
				t.Errorf("score = %v, want %v", result.score, tt.wantScore)
			}
		})
	}
}

func TestVerifyCrosswordAnswers(t *testing.T) {
	content := json.RawMessage(`{"words":[{"word":"HELLO","clue":"Greeting"},{"word":"WORLD","clue":"Earth"}]}`)

	t.Run("all words solved is completed", func(t *testing.T) {
		result, ok := verifyCrosswordAnswers(content, json.RawMessage(`{"solved_words":["HELLO","WORLD"]}`))
		if !ok {
			t.Fatal("expected ok=true")
		}
		if !result.completed || result.score != 2 || result.maxScore != 2 {
			t.Errorf("unexpected result: %+v", result)
		}
	})

	t.Run("partial grid is not completed (all-or-nothing per M1.2)", func(t *testing.T) {
		result, ok := verifyCrosswordAnswers(content, json.RawMessage(`{"solved_words":["HELLO"]}`))
		if !ok {
			t.Fatal("expected ok=true")
		}
		if result.completed {
			t.Error("one of two words should not be completed")
		}
		if result.score != 1 {
			t.Errorf("score = %d, want 1", result.score)
		}
	})

	t.Run("duplicate or unknown submitted words do not inflate the score", func(t *testing.T) {
		result, ok := verifyCrosswordAnswers(content, json.RawMessage(`{"solved_words":["HELLO","hello","NOPE"]}`))
		if !ok {
			t.Fatal("expected ok=true")
		}
		if result.score != 1 {
			t.Errorf("score = %d, want 1 (case-insensitive duplicate + unknown word must not count twice)", result.score)
		}
	})
}

func TestVerifyWordSearchAnswers(t *testing.T) {
	content := json.RawMessage(`{"words":["CAT","DOG","BIRD"]}`)

	t.Run("all words found is completed", func(t *testing.T) {
		result, ok := verifyWordSearchAnswers(content, json.RawMessage(`{"found_words":["CAT","DOG","BIRD"]}`))
		if !ok {
			t.Fatal("expected ok=true")
		}
		if !result.completed || result.score != 3 || result.maxScore != 3 {
			t.Errorf("unexpected result: %+v", result)
		}
	})

	t.Run("missing a word is not completed", func(t *testing.T) {
		result, ok := verifyWordSearchAnswers(content, json.RawMessage(`{"found_words":["CAT","DOG"]}`))
		if !ok {
			t.Fatal("expected ok=true")
		}
		if result.completed {
			t.Error("2 of 3 words should not be completed")
		}
	})

	t.Run("accented content word matches unaccented submission", func(t *testing.T) {
		accented := json.RawMessage(`{"words":["ÁGUILA"]}`)
		result, ok := verifyWordSearchAnswers(accented, json.RawMessage(`{"found_words":["AGUILA"]}`))
		if !ok {
			t.Fatal("expected ok=true")
		}
		if !result.completed {
			t.Error("expected accent-insensitive match to complete the level")
		}
	})
}

func TestVerifyPuzzleAnswers(t *testing.T) {
	content := json.RawMessage(`{"phrase":"secret message","pieces":4}`)

	t.Run("identity order is completed", func(t *testing.T) {
		result, ok := verifyPuzzleAnswers(content, json.RawMessage(`{"piece_order":[0,1,2,3]}`))
		if !ok {
			t.Fatal("expected ok=true")
		}
		if !result.completed || result.score != 4 || result.maxScore != 4 {
			t.Errorf("unexpected result: %+v", result)
		}
	})

	t.Run("shuffled order is not completed", func(t *testing.T) {
		result, ok := verifyPuzzleAnswers(content, json.RawMessage(`{"piece_order":[1,0,2,3]}`))
		if !ok {
			t.Fatal("expected ok=true")
		}
		if result.completed {
			t.Error("shuffled order should not be completed")
		}
		if result.score != 2 {
			t.Errorf("score = %d, want 2 (positions 2 and 3 already correct)", result.score)
		}
	})

	t.Run("missing pieces count defaults to 3", func(t *testing.T) {
		noCount := json.RawMessage(`{"phrase":"abc"}`)
		result, ok := verifyPuzzleAnswers(noCount, json.RawMessage(`{"piece_order":[0,1,2]}`))
		if !ok {
			t.Fatal("expected ok=true")
		}
		if result.maxScore != 3 {
			t.Errorf("maxScore = %d, want 3", result.maxScore)
		}
	})

	t.Run("wrong-length piece_order is not verifiable", func(t *testing.T) {
		if _, ok := verifyPuzzleAnswers(content, json.RawMessage(`{"piece_order":[0,1]}`)); ok {
			t.Error("expected ok=false for a piece_order of the wrong length")
		}
	})
}

func TestVerifyAnswers_UnsupportedOrMissing(t *testing.T) {
	content := json.RawMessage(`{"questions":[{"question":"Q1","options":["A","B"],"correct_index":0}]}`)

	if _, ok := verifyAnswers("trivia", content, nil); ok {
		t.Error("expected ok=false when answers is missing")
	}
	if _, ok := verifyAnswers("memory", content, json.RawMessage(`{}`)); ok {
		t.Error("expected ok=false for a template with no verifier (memory)")
	}
	if _, ok := verifyAnswers("snakes_ladders", content, json.RawMessage(`{}`)); ok {
		t.Error("expected ok=false for a template with no verifier (snakes_ladders)")
	}
}

func TestVerifyAnswers_ManipulatedClientPayloadIsIgnored(t *testing.T) {
	// El caso descrito en M1.1: un cliente que manda completed:true,
	// score:99999 pero cuyas respuestas reales solo aciertan 1 de 10.
	content := json.RawMessage(`{"questions":[
		{"question":"Q1","options":["A","B"],"correct_index":0},
		{"question":"Q2","options":["A","B"],"correct_index":0},
		{"question":"Q3","options":["A","B"],"correct_index":0},
		{"question":"Q4","options":["A","B"],"correct_index":0},
		{"question":"Q5","options":["A","B"],"correct_index":0},
		{"question":"Q6","options":["A","B"],"correct_index":0},
		{"question":"Q7","options":["A","B"],"correct_index":0},
		{"question":"Q8","options":["A","B"],"correct_index":0},
		{"question":"Q9","options":["A","B"],"correct_index":0},
		{"question":"Q10","options":["A","B"],"correct_index":0}
	]}`)
	answers := json.RawMessage(`{"selected_indices":[0,1,1,1,1,1,1,1,1,1]}`)

	result, ok := verifyAnswers("trivia", content, answers)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if result.completed {
		t.Error("1 of 10 correct should not be completed regardless of what the client claimed")
	}
	if result.score != 1 || result.maxScore != 10 {
		t.Errorf("score/maxScore = %d/%d, want 1/10", result.score, result.maxScore)
	}
}
