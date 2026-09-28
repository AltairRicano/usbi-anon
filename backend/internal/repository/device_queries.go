package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type CreateDeviceParams struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	DeviceKind string
	Platform   string
}

func (q *Queries) CreateDevice(ctx context.Context, arg CreateDeviceParams) (Device, error) {
	row := q.db.QueryRowContext(ctx, `
INSERT INTO devices (id, user_id, device_kind, platform)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, device_kind, platform, registered_at, last_seen_at, wipe_local_data, revoked_at
`, arg.ID, arg.UserID, arg.DeviceKind, arg.Platform)
	return scanDevice(row)
}

type GetActiveDeviceParams struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

func (q *Queries) GetActiveDevice(ctx context.Context, arg GetActiveDeviceParams) (Device, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT id, user_id, device_kind, platform, registered_at, last_seen_at, wipe_local_data, revoked_at
FROM devices
WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL
`, arg.ID, arg.UserID)
	return scanDevice(row)
}

func (q *Queries) TouchDevice(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE devices
SET last_seen_at = NOW()
WHERE id = $1 AND revoked_at IS NULL
`, id)
	return err
}

// TouchDeviceReturning es la mitad "actualizar" del upsert de C1
// (devices.Service.RegisterDevice): mismo WHERE que GetActiveDevice —
// (id, user_id) coincide y no está revocado — pero como UPDATE ...
// RETURNING en una sola consulta, para no depender de un GetActiveDevice +
// TouchDevice separados con una fila potencialmente cambiando entre medio.
// sql.ErrNoRows significa "no es tuyo, no existe, o está revocado" — el
// llamador decide crear un dispositivo nuevo en ese caso.
func (q *Queries) TouchDeviceReturning(ctx context.Context, arg GetActiveDeviceParams) (Device, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE devices
SET last_seen_at = NOW()
WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL
RETURNING id, user_id, device_kind, platform, registered_at, last_seen_at, wipe_local_data, revoked_at
`, arg.ID, arg.UserID)
	return scanDevice(row)
}

func (q *Queries) MarkUserDevicesForWipe(ctx context.Context, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE devices
SET wipe_local_data = true
WHERE user_id = $1 AND revoked_at IS NULL
`, userID)
	return err
}

// RevokeDevice revoca un dispositivo (no hace DELETE físico para no borrar el historial en cascada).
// Marca revoked_at y wipe_local_data para borrar el SQLite local en el siguiente sync.
func (q *Queries) RevokeDevice(ctx context.Context, arg GetActiveDeviceParams) (Device, error) {
	row := q.db.QueryRowContext(ctx, `
UPDATE devices
SET revoked_at = NOW(), wipe_local_data = true
WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL
RETURNING id, user_id, device_kind, platform, registered_at, last_seen_at, wipe_local_data, revoked_at
`, arg.ID, arg.UserID)
	return scanDevice(row)
}

type ListDevicesParams struct {
	UserID uuid.UUID
}

func (q *Queries) ListDevices(ctx context.Context, arg ListDevicesParams) ([]Device, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT id, user_id, device_kind, platform, registered_at, last_seen_at, wipe_local_data, revoked_at
FROM devices
WHERE user_id = $1 AND revoked_at IS NULL
ORDER BY last_seen_at DESC
`, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		device, err := scanDevice(rows)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return devices, nil
}

func scanDevice(row interface {
	Scan(dest ...interface{}) error
}) (Device, error) {
	var device Device
	var revokedAt sql.NullTime
	err := row.Scan(
		&device.ID,
		&device.UserID,
		&device.DeviceKind,
		&device.Platform,
		&device.RegisteredAt,
		&device.LastSeenAt,
		&device.WipeLocalData,
		&revokedAt,
	)
	device.RevokedAt = revokedAt
	return device, err
}
