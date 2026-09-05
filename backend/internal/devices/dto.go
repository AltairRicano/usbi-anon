package devices

import (
	"time"

	"github.com/google/uuid"
)

// DeviceKind debe ser uno de: movil, tablet, laptop, escritorio, otro — el
// mismo vocabulario cerrado que el CHECK de devices.device_kind (ver
// plan/01_Base_de_datos.md §3.4). No es texto libre. (Útil)
type RegisterDeviceRequest struct {
	DeviceKind string `json:"device_kind"`
	Platform   string `json:"platform"`
}

type DeviceResponse struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	DeviceKind    string     `json:"device_kind"`
	Platform      string     `json:"platform"`
	RegisteredAt  time.Time  `json:"registered_at"`
	LastSeenAt    time.Time  `json:"last_seen_at"`
	WipeLocalData bool       `json:"wipe_local_data"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
}

type DevicesResponse struct {
	Items []DeviceResponse `json:"items"`
}
