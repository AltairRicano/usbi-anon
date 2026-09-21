// Package maintenance implementa las tareas periódicas de retención legal y purga:
// suspender y cancelar cuentas de jugadores inactivos de acuerdo a las políticas
// de retención, y purgar refresh tokens expirados o revocados.
package maintenance

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/altair/usbi-anon-backend/internal/privacy"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

type Config struct {
	InactiveSuspendAfter time.Duration
	SuspendedCancelAfter time.Duration
	BatchSize            int32
}

type Summary struct {
	InactiveSuspended   int64
	SuspendedCancelled  int
	RefreshTokensPurged int64
}

type Service struct {
	repo *repository.Queries
	cfg  Config
}

func NewService(repo *repository.Queries, cfg Config) *Service {
	if cfg.InactiveSuspendAfter <= 0 {
		cfg.InactiveSuspendAfter = 365 * 24 * time.Hour
	}
	if cfg.SuspendedCancelAfter <= 0 {
		cfg.SuspendedCancelAfter = 30 * 24 * time.Hour
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	return &Service{repo: repo, cfg: cfg}
}

func (s *Service) RunOnce(ctx context.Context, now time.Time) (Summary, error) {
	now = now.UTC()
	var summary Summary

	suspended, err := s.repo.SuspendInactivePlayers(ctx, now.Add(-s.cfg.InactiveSuspendAfter))
	if err != nil {
		return summary, fmt.Errorf("suspending inactive players: %w", err)
	}
	summary.InactiveSuspended = suspended

	cancelIDs, err := s.repo.ListSuspendedUsersForCancellation(ctx, now.Add(-s.cfg.SuspendedCancelAfter), s.cfg.BatchSize)
	if err != nil {
		return summary, fmt.Errorf("listing suspended users for cancellation: %w", err)
	}
	for _, accountID := range cancelIDs {
		if err := s.cancelAccount(ctx, accountID); err != nil {
			return summary, err
		}
		summary.SuspendedCancelled++
	}

	// Purga de mantenimiento: elimina refresh tokens expirados/revocados hace tiempo para que la
	// tabla no crezca sin límite.
	refreshTokens, err := s.repo.PurgeExpiredRefreshTokens(ctx)
	if err != nil {
		return summary, fmt.Errorf("purging expired refresh tokens: %w", err)
	}
	summary.RefreshTokensPurged = refreshTokens

	return summary, nil
}

// cancelAccount ejecuta la retención legal automática con la misma
// CancelAccount que usa la cancelación autoservicio (DELETE /auth/me).
// Idempotente por construcción: si un reintento la vuelve a listar,
// CancelAccount no encuentra una fila activa que actualizar.
func (s *Service) cancelAccount(ctx context.Context, accountID uuid.UUID) error {
	return privacy.CancelAccount(ctx, s.repo, privacy.CancelAccountParams{
		AccountID: accountID,
		Reason:    "inactivity_retention_expired",
	})
}

func StartScheduler(ctx context.Context, svc *Service, interval time.Duration, logger *log.Logger) {
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	if logger == nil {
		logger = log.Default()
	}

	go func() {
		run := func() {
			summary, err := svc.RunOnce(ctx, time.Now())
			if err != nil {
				logger.Printf("[WARN] legal maintenance failed: %v", err)
				return
			}
			if summary.InactiveSuspended > 0 || summary.SuspendedCancelled > 0 || summary.RefreshTokensPurged > 0 {
				logger.Printf("[INFO] legal maintenance completed: inactive_suspended=%d suspended_cancelled=%d refresh_tokens_purged=%d",
					summary.InactiveSuspended, summary.SuspendedCancelled, summary.RefreshTokensPurged)
			}
		}

		run()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}
