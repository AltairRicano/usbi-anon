package devices

import (
	"time"

	"github.com/google/uuid"
)

type RegisterDeviceRequest struct {
	DeviceLabel string `json:"device_label"`
	Platform    string `json:"platform"`
}

type DeviceResponse struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	DeviceLabel   string     `json:"device_label"`
	Platform      string     `json:"platform"`
	RegisteredAt  time.Time  `json:"registered_at"`
	LastSeenAt    time.Time  `json:"last_seen_at"`
	WipeLocalData bool       `json:"wipe_local_data"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
}

type DevicesResponse struct {
	Items []DeviceResponse `json:"items"`
}
