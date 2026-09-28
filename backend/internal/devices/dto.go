package devices

import (
	"time"

	"github.com/google/uuid"
)

// DeviceKind debe ser uno de: movil, tablet, laptop, escritorio, otro — el
// mismo vocabulario cerrado que el CHECK de devices.device_kind (ver
// plan/01_Base_de_datos.md §3.4). No es texto libre.
// DeviceID es opcional (C1, estado_proyecto.md 2026-09-10): el cliente lo
// manda cuando ya conoce el id de este dispositivo (guardado localmente en
// un login anterior), para que el registro automático en cada login sea un
// upsert en vez de crear una fila nueva cada vez. El id que propone el
// cliente nunca se usa para el INSERT — solo se consulta contra
// (id, user_id) antes de decidir si actualiza o crea, así que nadie puede
// reclamar el dispositivo de otra cuenta ni forzar una colisión de PK.
type RegisterDeviceRequest struct {
	DeviceID   *uuid.UUID `json:"device_id,omitempty"`
	DeviceKind string     `json:"device_kind"`
	Platform   string     `json:"platform"`
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
