package devices

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/altair/usbi-anon-backend/internal/audit"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
)

// validDeviceKinds refleja el CHECK de devices.device_kind. Rechazar aquí,
// antes del INSERT, da un 422 legible en vez de un error crudo de Postgres. (Útil)
var validDeviceKinds = map[string]bool{
	"movil": true, "tablet": true, "laptop": true, "escritorio": true, "otro": true,
}

type Service struct {
	repo *repository.Queries
}

func NewService(repo *repository.Queries) *Service {
	return &Service{repo: repo}
}

// RegisterDevice es un upsert (C1, estado_proyecto.md 2026-09-10): el
// frontend lo llama después de cada login exitoso, no detrás de un botón
// manual. Si req.DeviceID viene y sigue siendo tuyo y activo, solo se
// actualiza last_seen_at; si no —dispositivo nuevo, ajeno, o revocado—, se
// crea uno con un id generado por el servidor. Así un mismo dispositivo no
// acumula una fila nueva por cada inicio de sesión, y el cliente nunca
// puede reclamar el dispositivo de otra cuenta.
//
// El segundo valor de retorno indica si se creó una fila nueva (para que el
// handler devuelva 201 solo en ese caso, y 200 cuando fue un touch).
func (s *Service) RegisterDevice(ctx context.Context, userID uuid.UUID, req RegisterDeviceRequest) (DeviceResponse, bool, error) {
	deviceKind := strings.TrimSpace(req.DeviceKind)
	if userID == uuid.Nil || !validDeviceKinds[deviceKind] {
		return DeviceResponse{}, false, ErrValidation
	}
	platform := strings.TrimSpace(req.Platform)
	if platform != "web" && platform != "tauri" {
		return DeviceResponse{}, false, ErrValidation
	}

	if req.DeviceID != nil {
		device, err := s.repo.TouchDeviceReturning(ctx, repository.GetActiveDeviceParams{
			ID:     *req.DeviceID,
			UserID: userID,
		})
		switch {
		case err == nil:
			return deviceToResponse(device), false, nil
		case errors.Is(err, sql.ErrNoRows):
			// No es tuyo, no existe, o está revocado: cae a crear uno nuevo.
		default:
			return DeviceResponse{}, false, err
		}
	}

	device, err := s.repo.CreateDevice(ctx, repository.CreateDeviceParams{
		ID:         uuid.New(),
		UserID:     userID,
		DeviceKind: deviceKind,
		Platform:   platform,
	})
	if err != nil {
		return DeviceResponse{}, false, err
	}
	return deviceToResponse(device), true, nil
}

// RevokeDevice maneja DELETE /devices/{device_id} (C2, estado_proyecto.md
// 2026-09-10): revocación lógica, no un DELETE físico — ver el comentario
// de repository.RevokeDevice para el porqué (la FK compuesta de
// sync_events). Se audita (`device.revoke`) porque, a diferencia de
// internal/suggestions, devices ya guarda user_id en claro — no hay
// anonimato que proteger aquí — y el pool de jugador (usbi_app) conserva
// INSERT sobre audit_log (00_roles_unificado.sql).
func (s *Service) RevokeDevice(ctx context.Context, userID, deviceID uuid.UUID) error {
	if userID == uuid.Nil || deviceID == uuid.Nil {
		return ErrValidation
	}

	tx, err := s.repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.repo.WithTx(tx)

	device, err := qtx.RevokeDevice(ctx, repository.GetActiveDeviceParams{ID: deviceID, UserID: userID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if err := audit.Log(ctx, qtx, audit.Entry{
		ActorID:    userID,
		Action:     "device.revoke",
		EntityType: "device",
		EntityID:   device.ID,
		After:      map[string]any{"id": device.ID, "device_kind": device.DeviceKind, "platform": device.Platform},
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) ListDevices(ctx context.Context, userID uuid.UUID) (DevicesResponse, error) {
	if userID == uuid.Nil {
		return DevicesResponse{}, ErrValidation
	}
	rows, err := s.repo.ListDevices(ctx, repository.ListDevicesParams{UserID: userID})
	if err != nil {
		return DevicesResponse{}, err
	}
	items := make([]DeviceResponse, 0, len(rows))
	for _, device := range rows {
		items = append(items, deviceToResponse(device))
	}
	return DevicesResponse{Items: items}, nil
}

func deviceToResponse(device repository.Device) DeviceResponse {
	resp := DeviceResponse{
		ID:            device.ID,
		UserID:        device.UserID,
		DeviceKind:    device.DeviceKind,
		Platform:      device.Platform,
		RegisteredAt:  device.RegisteredAt,
		LastSeenAt:    device.LastSeenAt,
		WipeLocalData: device.WipeLocalData,
	}
	if device.RevokedAt.Valid {
		t := device.RevokedAt.Time
		resp.RevokedAt = &t
	}
	return resp
}
