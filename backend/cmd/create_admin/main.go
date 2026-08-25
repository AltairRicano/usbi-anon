// Reescrito desde ../usbi/backend/cmd/create_admin/main.go: escribe en las
// DOS bases (plan/02_Backend.md §2.2). Registrar y promover a admin son
// operaciones sobre `identities` [I], pero además crea la fila `accounts`
// [P] directamente en vez de esperar al primer Login: admin_audit_log.
// actor_user_id tiene FK a accounts(id) (ver
// migrations/main/0001_esquema_principal.up.sql), así que un admin recién
// creado por este CLI no podría publicar ni una sección hasta iniciar
// sesión al menos una vez si no se le crea la réplica aquí mismo.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/altair/usbi-anon-backend/internal/auth"
	"github.com/altair/usbi-anon-backend/internal/config"
	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/identityrepo"
	"github.com/altair/usbi-anon-backend/internal/repository"
	_ "github.com/lib/pq"
)

func main() {
	config.LoadEnvironment()
	ctx := context.Background()
	identDBURL := config.IdentityDatabaseURL()
	mainDBURL := config.DatabaseURL()

	identDB, err := sql.Open("postgres", identDBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer identDB.Close()

	mainDB, err := sql.Open("postgres", mainDBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer mainDB.Close()

	identQueries := identityrepo.New(identDB)
	mainQueries := repository.New(mainDB)
	cfg := auth.Config{
		EncryptionKey:    config.RequireSecret("PGP_ENCRYPTION_KEY"),
		BlindIndexSecret: []byte(config.RequireSecret("BLIND_INDEX_SECRET")),
		HMACSecret:       []byte(config.RequireSecret("HMAC_SECRET")),
		TokenConfig: crypto.TokenConfig{
			Secret: []byte(config.RequireSecret("JWT_SECRET")),
		},
	}

	svc := auth.NewService(identQueries, mainQueries, cfg)

	// Credentials come from the environment — never hardcoded. Without this the
	// initial admin login would be public knowledge from the repository.
	adminPassword := config.RequireEnv("ADMIN_PASSWORD")
	if len(adminPassword) < 12 {
		log.Fatalf("ADMIN_PASSWORD must be at least 12 characters (got %d)", len(adminPassword))
	}
	req := auth.RegisterRequest{
		Email:                config.RequireEnv("ADMIN_EMAIL"),
		Password:             adminPassword,
		IsAdult:              true,
		PrivacyNoticeVersion: config.GetEnv("ADMIN_PRIVACY_NOTICE_VERSION", "v1.0"),
	}

	resp, err := svc.Register(ctx, req)
	if err != nil {
		log.Fatalf("Failed to register: %v", err)
	}

	fmt.Printf("Registered user: %v\n", resp.UserID)

	// Now promote to admin
	_, err = identDB.ExecContext(ctx, "UPDATE identities SET role = 'admin' WHERE id = $1", resp.UserID)
	if err != nil {
		log.Fatalf("Failed to promote: %v", err)
	}

	// Crea la réplica accounts [P] con el rol ya en 'admin' — sin esto,
	// cualquier acción de contenido que este admin haga antes de su primer
	// login (que es lo normal para un bootstrap por CLI) fallaría por la FK
	// de admin_audit_log.actor_user_id → accounts(id).
	adjectiveID, nounID, number, err := repository.RandomAlias()
	if err != nil {
		log.Fatalf("Failed to generate alias: %v", err)
	}
	if err := mainQueries.UpsertAccount(ctx, repository.UpsertAccountParams{
		ID:               resp.UserID,
		Role:             string(domain.RoleAdmin),
		Status:           string(domain.StatusActive),
		IsAdult:          true,
		AliasAdjectiveID: adjectiveID,
		AliasNounID:      nounID,
		AliasNumber:      number,
	}); err != nil {
		log.Fatalf("Failed to create account replica: %v", err)
	}

	// No-Repudio: record the privilege escalation in the identity audit
	// ledger (A3). identity_audit_log, no admin_audit_log: un cambio de rol
	// es una acción sobre la identidad, no sobre el progreso.
	if err := identQueries.LogIdentityAudit(ctx, identityrepo.IdentityAuditEntry{
		ActorID:    resp.UserID,
		Action:     "admin.bootstrap",
		EntityType: "user",
		EntityID:   resp.UserID,
		Before:     map[string]any{"role": "player"},
		After:      map[string]any{"role": "admin"},
		UserAgent:  "create_admin-cli",
	}); err != nil {
		log.Fatalf("Failed to write audit log: %v", err)
	}

	fmt.Println("Success! Admin user created.")
}
