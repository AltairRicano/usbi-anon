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
