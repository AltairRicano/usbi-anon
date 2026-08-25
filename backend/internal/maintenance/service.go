// Reescrito desde ../usbi/backend/internal/maintenance/service.go para el
// reparto en dos bases. cancelUser (retención legal automática) ya no tiene
// una fila arco_requests que actualizar como checkpoint — a diferencia de la
// saga de ResolveArcoRequest (internal/auth), aquí basta con que PurgeMain y
// PseudonymizeIdentity sean idempotentes (internal/privacy): si el proceso
// muere entre las dos, el siguiente RunOnce vuelve a listar al mismo usuario
// (SuspendInactivePlayers/ListSuspendedUsersForCancellation solo lo sacan de
// la lista cuando ya quedó con status='deleted') y reintenta sin corromper
// nada. reconcileStuckArcoRequests es la pieza nueva: el "job de
// reconciliación" que plan/02_Backend.md §5 regla 5 exige dentro de este
// mismo planificador, para trámites ARCO que se quedaron a medio camino.
package maintenance

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/altair/usbi-anon-backend/internal/identityrepo"
	"github.com/altair/usbi-anon-backend/internal/privacy"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/google/uuid"
)

type Config struct {
	EncryptionKey        string
	BlindIndexSecret     []byte
	PendingTutorTTL      time.Duration
	InactiveSuspendAfter time.Duration
	SuspendedCancelAfter time.Duration
	// ArcoStuckAfter es cuánto tiempo debe llevar un trámite ARCO en
	// purging_main/identity_pseudonymized antes de que el job de
	// reconciliación lo retome. No debería importar demasiado en la
	// práctica (RunOnce corre cada 24h por defecto), pero evita que este
	// job compita con una llamada a ResolveArcoRequest que apenas está en
	// curso en otro proceso.
	ArcoStuckAfter time.Duration
	BatchSize      int32
}

type Summary struct {
	PendingTutorPurged  int
	InactiveSuspended   int64
	SuspendedCancelled  int
	ArcoRequestsResumed int
	TutorTokensPurged   int64
	RefreshTokensPurged int64
}

// Service recibe DOS Queries, igual que internal/auth: la retención legal
// automática y la reconciliación de ARCO atascados tocan ambas bases.
type Service struct {
	ident *identityrepo.Queries
	main  *repository.Queries
	cfg   Config
}

func NewService(ident *identityrepo.Queries, main *repository.Queries, cfg Config) *Service {
	if cfg.EncryptionKey == "" {
		panic("maintenance.Config: EncryptionKey must not be empty")
	}
	if len(cfg.BlindIndexSecret) == 0 {
		panic("maintenance.Config: BlindIndexSecret must not be empty")
	}
	if cfg.PendingTutorTTL <= 0 {
		cfg.PendingTutorTTL = 48 * time.Hour
	}
	if cfg.InactiveSuspendAfter <= 0 {
		cfg.InactiveSuspendAfter = 365 * 24 * time.Hour
	}
	if cfg.SuspendedCancelAfter <= 0 {
		cfg.SuspendedCancelAfter = 30 * 24 * time.Hour
	}
	if cfg.ArcoStuckAfter <= 0 {
		cfg.ArcoStuckAfter = 15 * time.Minute
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	return &Service{ident: ident, main: main, cfg: cfg}
}

func (s *Service) RunOnce(ctx context.Context, now time.Time) (Summary, error) {
	now = now.UTC()
	var summary Summary

	pendingIDs, err := s.ident.ListPendingTutorConsentUsers(ctx, now.Add(-s.cfg.PendingTutorTTL), s.cfg.BatchSize)
	if err != nil {
		return summary, fmt.Errorf("listing pending tutor users: %w", err)
	}
	for _, userID := range pendingIDs {
		if err := s.cancelUser(ctx, userID, "pending_tutor_consent_expired"); err != nil {
			return summary, err
		}
		summary.PendingTutorPurged++
	}

	suspended, err := s.ident.SuspendInactivePlayers(ctx, now.Add(-s.cfg.InactiveSuspendAfter))
	if err != nil {
		return summary, fmt.Errorf("suspending inactive players: %w", err)
	}
	summary.InactiveSuspended = suspended

	cancelIDs, err := s.ident.ListSuspendedUsersForCancellation(ctx, now.Add(-s.cfg.SuspendedCancelAfter), s.cfg.BatchSize)
	if err != nil {
		return summary, fmt.Errorf("listing suspended users for cancellation: %w", err)
	}
	for _, userID := range cancelIDs {
		if err := s.cancelUser(ctx, userID, "inactivity_retention_expired"); err != nil {
			return summary, err
		}
		summary.SuspendedCancelled++
	}

	resumed, err := s.reconcileStuckArcoRequests(ctx, now)
	if err != nil {
		return summary, fmt.Errorf("reconciling stuck arco requests: %w", err)
	}
	summary.ArcoRequestsResumed = resumed

	// Housekeeping purges (A8 / A1): drop expired unverified tutor-consent tokens
	// and expired/long-revoked refresh tokens so neither table grows unbounded.
	tutorTokens, err := s.ident.PurgeExpiredTutorConsentTokens(ctx)
	if err != nil {
		return summary, fmt.Errorf("purging expired tutor consent tokens: %w", err)
	}
	summary.TutorTokensPurged = tutorTokens

	refreshTokens, err := s.ident.PurgeExpiredRefreshTokens(ctx)
	if err != nil {
		return summary, fmt.Errorf("purging expired refresh tokens: %w", err)
	}
	summary.RefreshTokensPurged = refreshTokens

	return summary, nil
}

// cancelUser ejecuta la retención legal automática con las mismas dos fases
// que ResolveArcoRequest (A4): purga en [P], luego seudonimiza en [I]. Sin
// arco_requests de por medio, así que sin checkpoint explícito — un
// reintento en el siguiente RunOnce vuelve a listar al usuario mientras siga
// sin quedar en status='deleted', y ambas fases son idempotentes por
// construcción (ver internal/privacy), así que reintentar la que ya corrió
// no hace daño.
func (s *Service) cancelUser(ctx context.Context, userID uuid.UUID, reason string) error {
	if err := privacy.PurgeMain(ctx, s.main, userID); err != nil {
		return fmt.Errorf("purging main data: %w", err)
	}
	if err := privacy.PseudonymizeIdentity(ctx, s.ident, privacy.CancelParams{
		UserID:           userID,
		Reason:           reason,
		EncryptionKey:    s.cfg.EncryptionKey,
		BlindIndexSecret: s.cfg.BlindIndexSecret,
	}); err != nil {
		return fmt.Errorf("pseudonymizing identity: %w", err)
	}
	return nil
}

// reconcileStuckArcoRequests es el job exigido por plan/02_Backend.md §5
// regla 5: rebarre solicitudes ARCO de cancelación que quedaron atascadas en
// purging_main o identity_pseudonymized porque el proceso que las manejaba
// murió a mitad de la saga, y las retoma con la misma
// privacy.ResumeArcoCancellation que usa internal/auth.Service.ResolveArcoRequest
// tras reclamar bajo lock — nunca hay dos implementaciones de "cómo terminar
// una cancelación a medias" (misma regla A4 de siempre).
func (s *Service) reconcileStuckArcoRequests(ctx context.Context, now time.Time) (int, error) {
	stuck, err := s.ident.ListStuckArcoRequests(ctx, now.Add(-s.cfg.ArcoStuckAfter), s.cfg.BatchSize)
	if err != nil {
		return 0, fmt.Errorf("listing stuck arco requests: %w", err)
	}

	resumed := 0
	for _, req := range stuck {
		if !req.UserID.Valid {
			// No debería ocurrir para una cancelación (solo llega a estos
			// estados si UserID era válido al reclamarse), pero si pasara,
			// no hay a quién purgar/pseudonimizar — se deja para revisión
			// manual en vez de fallar todo el batch.
			continue
		}
		if err := privacy.ResumeArcoCancellation(ctx, s.ident, s.main, req.ID, req.Status, privacy.CancelParams{
			UserID:           req.UserID.UUID,
			Reason:           "arco_cancelacion",
			EncryptionKey:    s.cfg.EncryptionKey,
			BlindIndexSecret: s.cfg.BlindIndexSecret,
		}); err != nil {
			return resumed, fmt.Errorf("resuming arco request %s: %w", req.ID, err)
		}
		if err := s.ident.LogIdentityAudit(ctx, identityrepo.IdentityAuditEntry{
			Action:     "arco.resolve.reconciled",
			EntityType: "arco_request",
			EntityID:   req.ID,
			Before:     map[string]any{"status": req.Status, "request_type": req.RequestType},
			After:      map[string]any{"status": "resolved", "subject_user_id": req.UserID.UUID},
			UserAgent:  "maintenance-scheduler",
		}); err != nil {
			return resumed, fmt.Errorf("logging arco reconciliation for %s: %w", req.ID, err)
		}
		resumed++
	}
	return resumed, nil
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
			if summary.PendingTutorPurged > 0 || summary.InactiveSuspended > 0 || summary.SuspendedCancelled > 0 ||
				summary.ArcoRequestsResumed > 0 || summary.TutorTokensPurged > 0 || summary.RefreshTokensPurged > 0 {
				logger.Printf("[INFO] legal maintenance completed: pending_tutor_purged=%d inactive_suspended=%d suspended_cancelled=%d arco_requests_resumed=%d tutor_tokens_purged=%d refresh_tokens_purged=%d",
					summary.PendingTutorPurged, summary.InactiveSuspended, summary.SuspendedCancelled,
					summary.ArcoRequestsResumed, summary.TutorTokensPurged, summary.RefreshTokensPurged)
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
