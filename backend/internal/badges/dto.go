package badges

import (
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

// BadgeResponse es la representación de una insignia para el panel admin. (Relleno)
type BadgeResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	XPThreshold int32     `json:"xp_threshold"`
	IconKey     string    `json:"icon_key"`
}

// CreateBadgeRequest es el cuerpo de POST /admin/badges. (Relleno)
type CreateBadgeRequest struct {
	Name        string `json:"name"`
	XPThreshold int32  `json:"xp_threshold"`
	IconKey     string `json:"icon_key"`
}

// UpdateBadgeRequest es el cuerpo de PATCH /admin/badges/{id}. (Relleno)
type UpdateBadgeRequest struct {
	Name        string `json:"name"`
	XPThreshold int32  `json:"xp_threshold"`
	IconKey     string `json:"icon_key"`
}

func toResponse(b repository.Badge) BadgeResponse {
	return BadgeResponse{
		ID:          b.ID,
		Name:        b.Name,
		XPThreshold: b.XpThreshold,
		IconKey:     b.IconKey,
	}
}

func badgeAuditPayload(b BadgeResponse) map[string]any {
	return map[string]any{
		"id": b.ID, "name": b.Name, "xp_threshold": b.XPThreshold, "icon_key": b.IconKey,
	}
}
