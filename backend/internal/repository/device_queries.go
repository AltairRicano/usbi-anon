package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type CreateDeviceParams struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	DeviceLabel string
	Platform    string
}

func (q *Queries) CreateDevice(ctx context.Context, arg CreateDeviceParams) (Device, error) {
	row := q.db.QueryRowContext(ctx, `
INSERT INTO devices (id, user_id, device_label, platform)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, device_label, platform, registered_at, last_seen_at, wipe_local_data, revoked_at
`, arg.ID, arg.UserID, arg.DeviceLabel, arg.Platform)
	return scanDevice(row)
}

type GetActiveDeviceParams struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

func (q *Queries) GetActiveDevice(ctx context.Context, arg GetActiveDeviceParams) (Device, error) {
	row := q.db.QueryRowContext(ctx, `
SELECT id, user_id, device_label, platform, registered_at, last_seen_at, wipe_local_data, revoked_at
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

func (q *Queries) MarkUserDevicesForWipe(ctx context.Context, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE devices
SET wipe_local_data = true
WHERE user_id = $1 AND revoked_at IS NULL
`, userID)
	return err
}

func (q *Queries) ClearDeviceWipeFlag(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, `
UPDATE devices
SET wipe_local_data = false
WHERE id = $1
`, id)
	return err
}

type ListDevicesParams struct {
	UserID uuid.UUID
}

func (q *Queries) ListDevices(ctx context.Context, arg ListDevicesParams) ([]Device, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT id, user_id, device_label, platform, registered_at, last_seen_at, wipe_local_data, revoked_at
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
		&device.DeviceLabel,
		&device.Platform,
		&device.RegisteredAt,
		&device.LastSeenAt,
		&device.WipeLocalData,
		&revokedAt,
	)
	device.RevokedAt = revokedAt
	return device, err
}
