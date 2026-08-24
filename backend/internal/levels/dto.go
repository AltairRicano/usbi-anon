package levels

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AllowedTemplateTypes are the valid values for template_type.
var AllowedTemplateTypes = map[string]struct{}{
	"trivia":         {},
	"puzzle":         {},
	"word_search":    {},
	"fake_news":      {},
	"crossword":      {},
	"memory":         {},
	"snakes_ladders": {},
}

type CreateLevelRequest struct {
	SectionID    uuid.UUID       `json:"section_id"`
	Title        string          `json:"title"`
	Color        string          `json:"color"`
	TemplateType string          `json:"template_type"`
	Content      json.RawMessage `json:"content"`
	Difficulty   int32           `json:"difficulty"`
}

type UpdateLevelRequest struct {
	Title        string          `json:"title"`
	Color        string          `json:"color"`
	TemplateType string          `json:"template_type"`
	Content      json.RawMessage `json:"content"`
	Difficulty   int32           `json:"difficulty"`
}

type LevelResponse struct {
	ID               uuid.UUID       `json:"id"`
	SectionID        uuid.UUID       `json:"section_id"`
	Title            string          `json:"title"`
	Color            string          `json:"color"`
	TemplateType     string          `json:"template_type"`
	Content          json.RawMessage `json:"content"`
	Difficulty       int32           `json:"difficulty"`
	IsPublished      bool            `json:"is_published"`
	CreatedByAdminID uuid.UUID       `json:"created_by_admin_id,omitempty"`
	CreatedAt        time.Time       `json:"created_at,omitempty"`
	UpdatedAt        time.Time       `json:"updated_at,omitempty"`
}

type CreateSectionRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type UpdateSectionRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type SectionResponse struct {
	ID               uuid.UUID `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Color            string    `json:"color"`
	IsPublished      bool      `json:"is_published"`
	CreatedByAdminID uuid.UUID `json:"created_by_admin_id,omitempty"`
	CreatedAt        time.Time `json:"created_at,omitempty"`
}

// LevelSummary is the public-safe DTO for listing levels (no content blob).
type LevelSummary struct {
	ID           uuid.UUID `json:"id"`
	SectionID    uuid.UUID `json:"section_id"`
	Title        string    `json:"title"`
	Color        string    `json:"color"`
	TemplateType string    `json:"template_type"`
	Difficulty   int32     `json:"difficulty"`
	IsPublished  bool      `json:"is_published"`
	CreatedAt    time.Time `json:"created_at"`
}

// LevelsPage is the paginated response for GET /levels.
type LevelsPage struct {
	Items      []LevelSummary `json:"items"`
	NextCursor string         `json:"next_cursor,omitempty"` // UUID string of last item, empty if no more pages
}

type SectionsResponse struct {
	Items []SectionResponse `json:"items"`
}

type CompleteLevelRequest struct {
	Score            int32           `json:"score"`
	Completed        bool            `json:"completed"`
	Answers          json.RawMessage `json:"answers,omitempty"`
	ClientFinishedAt string          `json:"client_finished_at,omitempty"`
}

type CompleteLevelResponse struct {
	LevelID       uuid.UUID       `json:"level_id"`
	Completed     bool            `json:"completed"`
	AttemptNumber int32           `json:"attempt_number"`
	XPAwarded     int32           `json:"xp_awarded"`
	TotalXP       int32           `json:"total_xp"`
	CurrentStreak int32           `json:"current_streak"`
	BadgesAwarded []BadgeResponse `json:"badges_awarded"`
}

type BadgeResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	XPThreshold int32     `json:"xp_threshold"`
	IconKey     string    `json:"icon_key"`
	EarnedAt    time.Time `json:"earned_at"`
}

type ProgressLevelResponse struct {
	LevelID          uuid.UUID  `json:"level_id"`
	Title            string     `json:"title"`
	TemplateType     string     `json:"template_type"`
	Difficulty       int32      `json:"difficulty"`
	BestScore        int32      `json:"best_score"`
	XPTotalForLevel  int32      `json:"xp_total_for_level"`
	AttemptsCount    int32      `json:"attempts_count"`
	FirstCompletedAt *time.Time `json:"first_completed_at,omitempty"`
	LastCompletedAt  *time.Time `json:"last_completed_at,omitempty"`
}

type ProfileProgressResponse struct {
	TotalXP         int32                   `json:"total_xp"`
	CompletedLevels int32                   `json:"completed_levels"`
	TotalAttempts   int32                   `json:"total_attempts"`
	CurrentStreak   int32                   `json:"current_streak"`
	Badges          []BadgeResponse         `json:"badges"`
	Levels          []ProgressLevelResponse `json:"levels"`
}
