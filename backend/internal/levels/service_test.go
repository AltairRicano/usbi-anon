package levels

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// MockRepository is a basic mock for testing CreateLevel logic without a DB.
type MockRepository struct {
	createdID uuid.UUID
	shouldErr bool
}

func (m *MockRepository) CreateLevel(ctx context.Context, arg interface{}) error {
	if m.shouldErr {
		return ErrValidation
	}
	return nil
}

func TestCreateLevel_Validation(t *testing.T) {
	// Not testing full DB injection here, just the service validation logic.
	// Since the service depends on *repository.Queries which is a concrete type,
	// we will just test the pure validation cases.

	svc := NewService(nil) // Will panic if DB is called, but we expect validation to catch it first.

	ctx := context.Background()
	adminID := uuid.New()

	tests := []struct {
		name    string
		req     CreateLevelRequest
		wantErr bool
	}{
		{
			name: "Missing Title",
			req: CreateLevelRequest{
				SectionID:    uuid.New(),
				TemplateType: "trivia",
				Content:      json.RawMessage(`[]`),
			},
			wantErr: true,
		},
		{
			name: "Missing SectionID",
			req: CreateLevelRequest{
				Title:        "Test Level",
				TemplateType: "trivia",
				Content:      json.RawMessage(`[]`),
			},
			wantErr: true,
		},
		{
			name: "Missing TemplateType",
			req: CreateLevelRequest{
				Title:     "Test Level",
				SectionID: uuid.New(),
				Content:   json.RawMessage(`[]`),
			},
			wantErr: true,
		},
		{
			name: "Invalid JSON Content",
			req: CreateLevelRequest{
				Title:        "Test Level",
				SectionID:    uuid.New(),
				TemplateType: "trivia",
				Content:      json.RawMessage(`invalid`),
			},
			wantErr: true,
		},
		{
			name: "Invalid TemplateType enum",
			req: CreateLevelRequest{
				Title:        "Test Level",
				SectionID:    uuid.New(),
				TemplateType: "unknown_type",
				Content:      json.RawMessage(`[{"question":"q","answer":"a"}]`),
			},
			wantErr: true,
		},
		{
			name: "Empty content array",
			req: CreateLevelRequest{
				Title:        "Test Level",
				SectionID:    uuid.New(),
				TemplateType: "trivia",
				Content:      json.RawMessage(`[]`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateLevel(ctx, adminID, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateXP(t *testing.T) {
	tests := []struct {
		name          string
		difficulty    int32
		attemptNumber int32
		completed     bool
		want          int32
	}{
		{name: "first completed attempt gets full base XP", difficulty: 5, attemptNumber: 1, completed: true, want: 20},
		{name: "second completed attempt gets half XP", difficulty: 5, attemptNumber: 2, completed: true, want: 10},
		{name: "third completed attempt gets half XP", difficulty: 5, attemptNumber: 3, completed: true, want: 10},
		{name: "fourth completed attempt gets zero XP", difficulty: 5, attemptNumber: 4, completed: true, want: 0},
		{name: "failed attempt gets zero XP", difficulty: 5, attemptNumber: 1, completed: false, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateXP(tt.difficulty, tt.attemptNumber, tt.completed)
			if got != tt.want {
				t.Fatalf("CalculateXP() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestValidateLevelInput_TemplateContracts(t *testing.T) {
	tests := []struct {
		name         string
		templateType string
		content      json.RawMessage
		wantErr      bool
	}{
		{
			name:         "valid trivia",
			templateType: "trivia",
			content:      json.RawMessage(`[{"question":"Que es el catalogo?","options":["Busqueda","Sancion"],"correct_index":0}]`),
		},
		{
			name:         "valid memory",
			templateType: "memory",
			content: json.RawMessage(`{"pairs":[
				{"id":"a","content1":"Catalogo","content2":"Busca materiales"},
				{"id":"b","content1":"Prestamo","content2":"Servicio de salida"},
				{"id":"c","content1":"Hemeroteca","content2":"Revistas"},
				{"id":"d","content1":"Referencia","content2":"Apoyo"}
			]}`),
		},
		{
			name:         "invalid memory needs four pairs",
			templateType: "memory",
			content:      json.RawMessage(`{"pairs":[{"id":"a","content1":"Uno","content2":"Dos"}]}`),
			wantErr:      true,
		},
		{
			name:         "valid fake news",
			templateType: "fake_news",
			content:      json.RawMessage(`{"news":[{"title":"Fuente anonima","content":"Siempre basta","isFake":true,"reference":"Guia USBI"}]}`),
		},
		{
			name:         "invalid fake news requires isFake",
			templateType: "fake_news",
			content:      json.RawMessage(`{"news":[{"title":"Fuente","content":"Texto","reference":"Guia USBI"}]}`),
			wantErr:      true,
		},
		{
			name:         "valid word search",
			templateType: "word_search",
			content:      json.RawMessage(`{"words":["catalogo","prestamo"],"width":12,"height":12,"seed":2026}`),
		},
		{
			name:         "invalid word search width",
			templateType: "word_search",
			content:      json.RawMessage(`{"words":["catalogo","prestamo"],"width":4}`),
			wantErr:      true,
		},
		{
			name:         "valid puzzle",
			templateType: "puzzle",
			content:      json.RawMessage(`{"phrase":"Ordena esta frase de prueba","pieces":5,"seed":2026}`),
		},
		{
			name:         "invalid puzzle needs phrase",
			templateType: "puzzle",
			content:      json.RawMessage(`{"pieces":5}`),
			wantErr:      true,
		},
		{
			name:         "invalid puzzle pieces",
			templateType: "puzzle",
			content:      json.RawMessage(`{"phrase":"Ordena esta frase","pieces":2}`),
			wantErr:      true,
		},
		{
			name:         "invalid puzzle unknown field",
			templateType: "puzzle",
			content:      json.RawMessage(`{"phrase":"Ordena esta frase","pieces":5,"gridSize":3}`),
			wantErr:      true,
		},
		{
			name:         "valid crossword",
			templateType: "crossword",
			content:      json.RawMessage(`{"words":[{"word":"catalogo","clue":"Busca materiales"},{"word":"cita","clue":"Referencia breve"}]}`),
		},
		{
			name:         "invalid crossword needs clue",
			templateType: "crossword",
			content:      json.RawMessage(`{"words":[{"word":"catalogo","clue":""},{"word":"cita","clue":"Referencia breve"}]}`),
			wantErr:      true,
		},
		{
			name:         "invalid crossword needs crossing words",
			templateType: "crossword",
			content:      json.RawMessage(`{"words":[{"word":"SI","clue":"Palabra usada para afirmar"},{"word":"NO","clue":"Palabra usada para negar"}]}`),
			wantErr:      true,
		},
		{
			name:         "invalid crossword rejects duplicate answers",
			templateType: "crossword",
			content:      json.RawMessage(`{"words":[{"word":"catalogo","clue":"Busca materiales"},{"word":"catálogo","clue":"Respuesta duplicada"}]}`),
			wantErr:      true,
		},
		{
			name:         "invalid snakes ladders needs questions",
			templateType: "snakes_ladders",
			content:      json.RawMessage(`{"board_width":6,"board_height":6,"start_position":1,"end_position":36,"snakes":[{"start":29,"end":13}],"ladders":[{"start":4,"end":16}]}`),
			wantErr:      true,
		},
		{
			name:         "valid snakes ladders with eight questions",
			templateType: "snakes_ladders",
			content: json.RawMessage(`{"board_width":6,"board_height":6,"start_position":1,"end_position":36,"ai_config":{"difficulty":"MEDIUM"},"questions":[
				{"question":"Capital de Veracruz","options":["Xalapa","Cordoba"],"correct_index":0},
				{"question":"Color institucional principal","options":["Azul","Rojo"],"correct_index":0},
				{"question":"USBI es una","options":["Biblioteca","Cafeteria"],"correct_index":0},
				{"question":"2+2","options":["4","5"],"correct_index":0},
				{"question":"El sol sale por el","options":["Este","Oeste"],"correct_index":0},
				{"question":"Agua moja","options":["Verdadero","Falso"],"correct_index":0},
				{"question":"La UV es una universidad","options":["Publica","Privada"],"correct_index":0},
				{"question":"Un dado tiene","options":["6 caras","4 caras"],"correct_index":0}
			]}`),
		},
		{
			name:         "invalid snakes ladders needs at least eight questions",
			templateType: "snakes_ladders",
			content:      json.RawMessage(`{"board_width":6,"board_height":6,"start_position":1,"end_position":36,"questions":[{"question":"Una","options":["A","B"],"correct_index":0}]}`),
			wantErr:      true,
		},
		{
			name:         "invalid snakes ladders question needs exactly two options",
			templateType: "snakes_ladders",
			content: json.RawMessage(`{"board_width":6,"board_height":6,"start_position":1,"end_position":36,"questions":[
				{"question":"Una","options":["A","B","C"],"correct_index":0},
				{"question":"Dos","options":["A","B"],"correct_index":0},
				{"question":"Tres","options":["A","B"],"correct_index":0},
				{"question":"Cuatro","options":["A","B"],"correct_index":0},
				{"question":"Cinco","options":["A","B"],"correct_index":0},
				{"question":"Seis","options":["A","B"],"correct_index":0},
				{"question":"Siete","options":["A","B"],"correct_index":0},
				{"question":"Ocho","options":["A","B"],"correct_index":0}
			]}`),
			wantErr: true,
		},
		{
			name:         "invalid snake must go down",
			templateType: "snakes_ladders",
			content:      json.RawMessage(`{"board_width":6,"board_height":6,"start_position":1,"end_position":36,"snakes":[{"start":5,"end":13}]}`),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLevelInput("Nivel", "#18529D", tt.templateType, 1, tt.content)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateLevelInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
