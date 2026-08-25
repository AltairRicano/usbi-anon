// Prueba de integración contra Postgres real: el job de reconciliación
// exigido por plan/02_Backend.md §5 regla 5. A diferencia de
// internal/auth.TestResolveArcoRequest_ResumesFromCrashedCheckpoint (que
// prueba la reanudación cuando SÍ vuelve a llamarse ResolveArcoRequest),
// esta prueba cubre el caso en que nadie vuelve a llamarlo — el admin cerró
// la pestaña, el proceso murió y nunca hubo un segundo intento HTTP. Solo el
// planificador de mantenimiento, corriendo cada 24h, puede terminar ese
// trámite.
package maintenance

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/identityrepo"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/altair/usbi-anon-backend/internal/testdb"
	"github.com/google/uuid"
)

func testConfig() Config {
	return Config{
		EncryptionKey:    "test-encryption-key-not-for-prod",
		BlindIndexSecret: []byte("test-blind-index-secret-32-bytes!"),
	}
}

// TestRunOnce_ReconcilesStuckArcoRequest simula un trámite ARCO de
// cancelación cuyo checkpoint quedó en purging_main (reclamado, pero sin que
// PurgeMain/PseudonymizeIdentity llegaran a correr) y nadie volvió a llamar
// ResolveArcoRequest. RunOnce debe encontrarlo vía ListStuckArcoRequests y
// terminarlo con la misma privacy.ResumeArcoCancellation que usa
// internal/auth — nunca una segunda implementación de "cómo cerrar una
// cancelación a medias" (regla A4).
func TestRunOnce_ReconcilesStuckArcoRequest(t *testing.T) {
	db := testdb.Setup(t)
	ctx := context.Background()

	adminID := uuid.New()
	if _, err := db.Ident.CreateIdentity(ctx, identityrepo.CreateIdentityParams{
		ID:                      adminID,
		EmailPlaintext:          "admin-reconcile@example.test",
		EmailLookupHash:         []byte("admin-reconcile-lookup-hash"),
		PasswordHash:            "unused-hash",
		TokenVersion:            1,
		IsAdult:                 true,
		Role:                    string(domain.RoleAdmin),
		PrivacyNoticeVersion:    "v1.0",
		PrivacyNoticeAcceptedAt: time.Now().UTC(),
		PrivacyAcceptanceHash:   []byte("seal"),
		CryptoKeyVersion:        1,
		Status:                  string(domain.StatusActive),
		EncryptionKey:           testConfig().EncryptionKey,
	}); err != nil {
		t.Fatalf("seeding admin identity: %v", err)
	}

	userID := uuid.New()
	if _, err := db.Ident.CreateIdentity(ctx, identityrepo.CreateIdentityParams{
		ID:                      userID,
		EmailPlaintext:          "player-reconcile@example.test",
		EmailLookupHash:         []byte("player-reconcile-lookup-hash"),
		PasswordHash:            "unused-hash",
		TokenVersion:            1,
		IsAdult:                 true,
		Role:                    string(domain.RolePlayer),
		PrivacyNoticeVersion:    "v1.0",
		PrivacyNoticeAcceptedAt: time.Now().UTC(),
		PrivacyAcceptanceHash:   []byte("seal"),
		CryptoKeyVersion:        1,
		Status:                  string(domain.StatusActive),
		EncryptionKey:           testConfig().EncryptionKey,
	}); err != nil {
		t.Fatalf("seeding player identity: %v", err)
	}

	if err := db.Main.UpsertAccount(ctx, repository.UpsertAccountParams{
		ID: userID, Role: string(domain.RolePlayer), Status: string(domain.StatusActive),
		IsAdult: true, AliasAdjectiveID: 1, AliasNounID: 1, AliasNumber: 1,
	}); err != nil {
		t.Fatalf("seeding account: %v", err)
	}
	device, err := db.Main.CreateDevice(ctx, repository.CreateDeviceParams{
		ID: uuid.New(), UserID: userID, DeviceKind: "movil", Platform: "web",
	})
	if err != nil {
		t.Fatalf("seeding device: %v", err)
	}

	requestID := uuid.New()
	if err := db.Ident.InsertArcoRequest(ctx, identityrepo.InsertArcoRequestParams{
		ID:            requestID,
		UserID:        uuid.NullUUID{UUID: userID, Valid: true},
		RequesterType: "user",
		RequestType:   string(domain.ArcoCancelacion),
		Status:        "pending",
		EvidenceHash:  []byte("integration-test-evidence"),
	}); err != nil {
		t.Fatalf("inserting arco request: %v", err)
	}
	if err := db.Ident.ClaimArcoRequestForCancellation(ctx, requestID,
		uuid.NullUUID{UUID: adminID, Valid: true}, "aprobado, proceso murió antes de purgar",
	); err != nil {
		t.Fatalf("claiming purging_main checkpoint: %v", err)
	}
	// Backdate received_at: ListStuckArcoRequests solo mira trámites más
	// viejos que cfg.ArcoStuckAfter, para no pisarse con una resolución que
	// apenas está en curso en otro proceso.
	if _, err := db.IdentDB.Exec(
		`UPDATE arco_requests SET received_at = NOW() - INTERVAL '1 hour' WHERE id = $1`, requestID,
	); err != nil {
		t.Fatalf("backdating received_at: %v", err)
	}

	svc := NewService(db.Ident, db.Main, testConfig())
	summary, err := svc.RunOnce(ctx, time.Now())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if summary.ArcoRequestsResumed != 1 {
		t.Errorf("ArcoRequestsResumed = %d, want 1", summary.ArcoRequestsResumed)
	}

	var wipe bool
	if err := db.MainDB.QueryRow(`SELECT wipe_local_data FROM devices WHERE id = $1`, device.ID).Scan(&wipe); err != nil {
		t.Fatalf("reading device wipe flag: %v", err)
	}
	if !wipe {
		t.Error("device.wipe_local_data = false, want true after reconciliation")
	}

	var status string
	var deletedAt sql.NullTime
	if err := db.IdentDB.QueryRow(
		`SELECT status, deleted_at FROM identities WHERE id = $1`, userID,
	).Scan(&status, &deletedAt); err != nil {
		t.Fatalf("reading identity after reconciliation: %v", err)
	}
	if status != string(domain.StatusDeleted) || !deletedAt.Valid {
		t.Errorf("identity status = %q (deleted_at valid=%v), want deleted", status, deletedAt.Valid)
	}

	var arcoStatus, handledBy string
	var resolvedAt sql.NullTime
	if err := db.IdentDB.QueryRow(
		`SELECT status, handled_by::text, resolved_at FROM arco_requests WHERE id = $1`, requestID,
	).Scan(&arcoStatus, &handledBy, &resolvedAt); err != nil {
		t.Fatalf("reading arco_request after reconciliation: %v", err)
	}
	if arcoStatus != "resolved" || !resolvedAt.Valid {
		t.Errorf("arco_requests.status = %q (resolved_at valid=%v), want resolved", arcoStatus, resolvedAt.Valid)
	}
	if handledBy != adminID.String() {
		t.Errorf("handled_by = %q, want %q (preserved from the original claim, not the reconciliation job)", handledBy, adminID)
	}

	var auditCount int
	if err := db.IdentDB.QueryRow(
		`SELECT COUNT(*) FROM identity_audit_log WHERE entity_id = $1 AND action = 'arco.resolve.reconciled'`, requestID,
	).Scan(&auditCount); err != nil {
		t.Fatalf("counting audit entries: %v", err)
	}
	if auditCount != 1 {
		t.Errorf("identity_audit_log entries for arco.resolve.reconciled = %d, want 1", auditCount)
	}

	// Un segundo RunOnce no debe volver a "resolver" nada: el trámite ya
	// salió del conjunto que ListStuckArcoRequests puede ver.
	summary2, err := svc.RunOnce(ctx, time.Now())
	if err != nil {
		t.Fatalf("second RunOnce: %v", err)
	}
	if summary2.ArcoRequestsResumed != 0 {
		t.Errorf("second RunOnce ArcoRequestsResumed = %d, want 0", summary2.ArcoRequestsResumed)
	}
}
