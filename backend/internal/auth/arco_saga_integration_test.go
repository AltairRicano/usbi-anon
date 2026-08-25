// Prueba de integración contra Postgres real (via internal/testdb) para la
// saga ARCO reanudable descrita en plan/02_Backend.md §5. No es una prueba
// unitaria con mocks: el objetivo específico de esta prueba es demostrar que
// un checkpoint persistido en arco_requests.status sobrevive a que el
// proceso original "muera" a mitad de camino, y que una llamada posterior a
// ResolveArcoRequest retoma exactamente donde se quedó sin repetir ni perder
// trabajo — algo que un mock de la base de datos no podría validar.
package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/identityrepo"
	"github.com/altair/usbi-anon-backend/internal/privacy"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/altair/usbi-anon-backend/internal/testdb"
	"github.com/google/uuid"
)

func testConfig() Config {
	return Config{
		EncryptionKey:    "test-encryption-key-not-for-prod",
		BlindIndexSecret: []byte("test-blind-index-secret-32-bytes!"),
		HMACSecret:       []byte("test-hmac-secret-32-bytes-long!!"),
		TokenConfig: crypto.TokenConfig{
			Secret:       []byte("test-jwt-secret-32-bytes-long!!!"),
			AccessExpiry: 15 * time.Minute,
		},
	}
}

func mustRegisterActiveAdult(t *testing.T, svc *Service, email string) uuid.UUID {
	t.Helper()
	resp, err := svc.Register(context.Background(), RegisterRequest{
		Email:                email,
		Password:             "correct horse battery staple",
		IsAdult:              true,
		PrivacyNoticeVersion: "v1.0",
	})
	if err != nil {
		t.Fatalf("registering %s: %v", email, err)
	}
	if resp.Status != string(domain.StatusActive) {
		t.Fatalf("expected active status for %s, got %s", email, resp.Status)
	}
	return resp.UserID
}

// seedAccountAndDevice crea la réplica accounts [P] (como haría el primer
// Login) y un dispositivo registrado, para que PurgeMain tenga algo
// observable que purgar/marcar para wipe.
func seedAccountAndDevice(t *testing.T, db *testdb.DB, userID uuid.UUID) uuid.UUID {
	t.Helper()
	if err := db.Main.UpsertAccount(context.Background(), repository.UpsertAccountParams{
		ID:               userID,
		Role:             string(domain.RolePlayer),
		Status:           string(domain.StatusActive),
		IsAdult:          true,
		AliasAdjectiveID: 1,
		AliasNounID:      1,
		AliasNumber:      1,
	}); err != nil {
		t.Fatalf("seeding account: %v", err)
	}
	device, err := db.Main.CreateDevice(context.Background(), repository.CreateDeviceParams{
		ID:         uuid.New(),
		UserID:     userID,
		DeviceKind: "movil",
		Platform:   "web",
	})
	if err != nil {
		t.Fatalf("seeding device: %v", err)
	}
	return device.ID
}

// TestResolveArcoRequest_ResumesFromCrashedCheckpoint simula el escenario
// central de plan/02_Backend.md §5: un proceso reclama el checkpoint
// purging_main (identityrepo.ClaimArcoRequestForCancellation) y muere ANTES
// de ejecutar la purga en la base principal. Una llamada posterior a
// ResolveArcoRequest debe detectar el checkpoint ya existente y terminar el
// trámite desde ahí — sin volver a pedir handled_by/response_summary, que ya
// quedaron guardados en el reclamo original.
func TestResolveArcoRequest_ResumesFromCrashedCheckpoint(t *testing.T) {
	db := testdb.Setup(t)
	svc := NewService(db.Ident, db.Main, testConfig())
	ctx := context.Background()

	adminID := mustRegisterActiveAdult(t, svc, "admin-arco-resume@example.test")
	if _, err := db.IdentDB.Exec(`UPDATE identities SET role = 'admin' WHERE id = $1`, adminID); err != nil {
		t.Fatalf("promoting admin: %v", err)
	}

	userID := mustRegisterActiveAdult(t, svc, "player-arco-resume@example.test")
	deviceID := seedAccountAndDevice(t, db, userID)

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

	// Simula el "crash": el checkpoint de reclamo ya se persistió (handled_by
	// y response_summary incluidos), pero PurgeMain nunca llegó a correr.
	if err := db.Ident.ClaimArcoRequestForCancellation(ctx, requestID,
		uuid.NullUUID{UUID: adminID, Valid: true}, "aprobado por prueba de integración",
	); err != nil {
		t.Fatalf("claiming purging_main checkpoint: %v", err)
	}

	assertDeviceWipeFlag(t, db, deviceID, false) // todavía no se tocó la base principal

	actor := domain.JWTClaims{UserID: adminID, Role: domain.RoleAdmin, TokenVersion: 1}
	resolveReq := ResolveArcoRequestDTO{Approved: true, ResponseSummary: "ignorado: ya se guardó en el reclamo"}

	if err := svc.ResolveArcoRequest(ctx, actor, requestID, resolveReq, "127.0.0.1", "integration-test"); err != nil {
		t.Fatalf("ResolveArcoRequest (resuming from purging_main): %v", err)
	}

	assertDeviceWipeFlag(t, db, deviceID, true)

	var status, deletionReason string
	var deletedAt sql.NullTime
	if err := db.IdentDB.QueryRow(
		`SELECT status, deletion_reason, deleted_at FROM identities WHERE id = $1`, userID,
	).Scan(&status, &deletionReason, &deletedAt); err != nil {
		t.Fatalf("reading identity after resolve: %v", err)
	}
	if status != string(domain.StatusDeleted) {
		t.Errorf("identity status = %q, want %q", status, domain.StatusDeleted)
	}
	if !deletedAt.Valid {
		t.Error("identity deleted_at not set")
	}
	if deletionReason != "arco_cancelacion" {
		t.Errorf("deletion_reason = %q, want arco_cancelacion", deletionReason)
	}

	var arcoStatus, handledBy, responseSummary string
	var resolvedAt sql.NullTime
	if err := db.IdentDB.QueryRow(
		`SELECT status, handled_by::text, response_summary, resolved_at FROM arco_requests WHERE id = $1`, requestID,
	).Scan(&arcoStatus, &handledBy, &responseSummary, &resolvedAt); err != nil {
		t.Fatalf("reading arco_request after resolve: %v", err)
	}
	if arcoStatus != "resolved" {
		t.Errorf("arco_requests.status = %q, want resolved", arcoStatus)
	}
	if !resolvedAt.Valid {
		t.Error("arco_requests.resolved_at not set")
	}
	if handledBy != adminID.String() {
		t.Errorf("handled_by = %q, want %q (preserved from the original claim)", handledBy, adminID)
	}
	if responseSummary != "aprobado por prueba de integración" {
		t.Errorf("response_summary = %q, want the value captured at claim time, not the (ignored) second call", responseSummary)
	}

	var auditCount int
	if err := db.IdentDB.QueryRow(
		`SELECT COUNT(*) FROM identity_audit_log WHERE entity_id = $1 AND action = 'arco.resolve'`, requestID,
	).Scan(&auditCount); err != nil {
		t.Fatalf("counting audit entries: %v", err)
	}
	if auditCount != 1 {
		t.Errorf("identity_audit_log entries for arco.resolve = %d, want 1", auditCount)
	}

	// Reintentar sobre un trámite ya resuelto debe rechazarse: demuestra que
	// la saga no vuelve a ejecutar la purga/seudonimización indefinidamente.
	if err := svc.ResolveArcoRequest(ctx, actor, requestID, resolveReq, "127.0.0.1", "integration-test"); err == nil {
		t.Error("expected ResolveArcoRequest on an already-resolved request to fail, got nil")
	}
}

func assertDeviceWipeFlag(t *testing.T, db *testdb.DB, deviceID uuid.UUID, want bool) {
	t.Helper()
	var got bool
	if err := db.MainDB.QueryRow(`SELECT wipe_local_data FROM devices WHERE id = $1`, deviceID).Scan(&got); err != nil {
		t.Fatalf("reading device wipe flag: %v", err)
	}
	if got != want {
		t.Errorf("devices.wipe_local_data = %v, want %v", got, want)
	}
}

// TestResolveArcoRequest_ResumesFromIdentityPseudonymizedCheckpoint cubre el
// TERCER estado intermedio de la saga (criterio 6 de plan/02_Backend.md §8
// exige probar la reanudación "desde cada estado intermedio", no solo uno):
// PurgeMain y PseudonymizeIdentity ya corrieron y el checkpoint
// identity_pseudonymized ya se escribió, pero el proceso murió antes de
// llegar a MarkArcoRequestResolved + la entrada de auditoría. Una llamada
// posterior a ResolveArcoRequest NO debe repetir la purga ni la
// seudonimización (ya se demuestra en el otro caso que ambas son
// idempotentes) — solo debe cerrar el trámite.
func TestResolveArcoRequest_ResumesFromIdentityPseudonymizedCheckpoint(t *testing.T) {
	db := testdb.Setup(t)
	cfg := testConfig()
	svc := NewService(db.Ident, db.Main, cfg)
	ctx := context.Background()

	adminID := mustRegisterActiveAdult(t, svc, "admin-arco-resume-2@example.test")
	if _, err := db.IdentDB.Exec(`UPDATE identities SET role = 'admin' WHERE id = $1`, adminID); err != nil {
		t.Fatalf("promoting admin: %v", err)
	}

	userID := mustRegisterActiveAdult(t, svc, "player-arco-resume-2@example.test")
	deviceID := seedAccountAndDevice(t, db, userID)

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
		uuid.NullUUID{UUID: adminID, Valid: true}, "aprobado por prueba de integración",
	); err != nil {
		t.Fatalf("claiming purging_main checkpoint: %v", err)
	}

	// Simula que [P] e [I] ya corrieron en un intento previo, y que ese
	// intento alcanzó a escribir el checkpoint identity_pseudonymized antes
	// de morir — sin llamar a svc.ResolveArcoRequest, que es justamente lo
	// que esta prueba necesita NO haber corrido todavía.
	if err := privacy.PurgeMain(ctx, db.Main, userID); err != nil {
		t.Fatalf("simulating [P] phase: %v", err)
	}
	if err := privacy.PseudonymizeIdentity(ctx, db.Ident, privacy.CancelParams{
		UserID: userID, Reason: "arco_cancelacion",
		EncryptionKey: cfg.EncryptionKey, BlindIndexSecret: cfg.BlindIndexSecret,
	}); err != nil {
		t.Fatalf("simulating [I] phase: %v", err)
	}
	if err := db.Ident.UpdateArcoRequestStatus(ctx, requestID, "identity_pseudonymized"); err != nil {
		t.Fatalf("writing identity_pseudonymized checkpoint: %v", err)
	}

	assertDeviceWipeFlag(t, db, deviceID, true) // ya lo dejó la fase [P] simulada arriba

	actor := domain.JWTClaims{UserID: adminID, Role: domain.RoleAdmin, TokenVersion: 1}
	resolveReq := ResolveArcoRequestDTO{Approved: true, ResponseSummary: "ignorado: ya se guardó en el reclamo"}
	if err := svc.ResolveArcoRequest(ctx, actor, requestID, resolveReq, "127.0.0.1", "integration-test"); err != nil {
		t.Fatalf("ResolveArcoRequest (resuming from identity_pseudonymized): %v", err)
	}

	var arcoStatus string
	var resolvedAt sql.NullTime
	if err := db.IdentDB.QueryRow(
		`SELECT status, resolved_at FROM arco_requests WHERE id = $1`, requestID,
	).Scan(&arcoStatus, &resolvedAt); err != nil {
		t.Fatalf("reading arco_request after resolve: %v", err)
	}
	if arcoStatus != "resolved" || !resolvedAt.Valid {
		t.Errorf("arco_requests.status = %q (resolved_at valid=%v), want resolved", arcoStatus, resolvedAt.Valid)
	}
}
