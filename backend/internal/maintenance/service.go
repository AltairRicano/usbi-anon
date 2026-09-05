// Reescrito en F7 (rediseño de identidad, ver
// plan/04_Rediseno_identidad_gustos.md §1 y §2) para la base única. Antes
// (F4, dos bases) este servicio recibía DOS *Queries y el RunOnce tenía
// cuatro responsabilidades: purgar registros atascados en
// 'pending_tutor_consent', suspender/cancelar jugadores inactivos, y
// reconciliar solicitudes ARCO de cancelación que quedaron a medio camino
// entre las dos fases de la saga.
//
// De las cuatro, dos desaparecen por completo:
//   - El flujo de tutor por correo se eliminó (§1) — no hay
//     'pending_tutor_consent' que purgar.
//   - internal/privacy.CancelAccount es ahora una sola *sql.Tx (§2): no hay
//     "a medio camino" posible, así que tampoco hay nada que reconciliar.
//
// Las otras dos (retención legal automática por inactividad, purga de
// refresh_tokens vencidos) siguen aplicando igual que antes, solo que contra
// una única base. (Útil)
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

	// Purga de mantenimiento (A1): elimina refresh tokens expirados/revocados hace tiempo para que la
	// tabla no crezca sin límite. (Útil)
	refreshTokens, err := s.repo.PurgeExpiredRefreshTokens(ctx)
	if err != nil {
		return summary, fmt.Errorf("purging expired refresh tokens: %w", err)
	}
	summary.RefreshTokensPurged = refreshTokens

	return summary, nil
}

// cancelAccount ejecuta la retención legal automática con la misma
// CancelAccount que usa la cancelación autoservicio (DELETE /auth/me) — un
// solo camino para "cómo se cancela una cuenta", disparado por dos motivos
// distintos (A4). Idempotente por construcción: si un reintento la vuelve a
// listar (no debería, DeactivateAccount la saca de 'suspended' en la misma
// tx), CancelAccount simplemente no encuentra fila con deleted_at IS NULL
// que actualizar. (Útil)
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
