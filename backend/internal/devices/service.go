package devices

import (
	"context"
	"errors"
	"strings"

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

var ErrValidation = errors.New("validation error")

type Service struct {
	repo *repository.Queries
}

func NewService(repo *repository.Queries) *Service {
	return &Service{repo: repo}
}

func (s *Service) RegisterDevice(ctx context.Context, userID uuid.UUID, req RegisterDeviceRequest) (DeviceResponse, error) {
	if userID == uuid.Nil || strings.TrimSpace(req.DeviceLabel) == "" {
		return DeviceResponse{}, ErrValidation
	}
	platform := strings.TrimSpace(req.Platform)
	if platform != "web" && platform != "tauri" {
		return DeviceResponse{}, ErrValidation
	}
	device, err := s.repo.CreateDevice(ctx, repository.CreateDeviceParams{
		ID:          uuid.New(),
		UserID:      userID,
		DeviceLabel: strings.TrimSpace(req.DeviceLabel),
		Platform:    platform,
	})
	if err != nil {
		return DeviceResponse{}, err
	}
	return deviceToResponse(device), nil
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
		DeviceLabel:   device.DeviceLabel,
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
