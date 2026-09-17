package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/altair/usbi-anon-backend/internal/auth"
	"github.com/altair/usbi-anon-backend/internal/config"
	"github.com/altair/usbi-anon-backend/internal/crypto"
	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/altair/usbi-anon-backend/internal/repository"
	_ "github.com/lib/pq"
	"golang.org/x/term"
)

// runAdmin implementa `usbictl admin create` — sustituye a
// backend/sql/01_seed_primer_admin.sql (SQL a mano + cmd/hash_password) por
// una sola invocación que llama exactamente al mismo código que ya usa
// POST /admin/accounts (auth.Service.CreateAdminAccount), así que las
// validaciones de nickname/password nunca pueden divergir entre el bootstrap
// y el uso normal (M3.4).
func runAdmin(args []string) error {
	if len(args) == 0 || args[0] != "create" {
		return errors.New("uso: usbictl admin create [--nickname <nick>] [--role admin|player]")
	}
	fs := flag.NewFlagSet("admin create", flag.ContinueOnError)
	nickname := fs.String("nickname", "admin01", "nickname de la cuenta (6-20 minúsculas/dígitos)")
	role := fs.String("role", "admin", "rol de la cuenta nueva (admin|player)")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	password, err := adminBootstrapPassword()
	if err != nil {
		return err
	}

	dbURL := config.DatabaseURL()
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		// Nunca incluir dbURL en el error: lleva la contraseña del rol.
		return fmt.Errorf("connecting with DB_USER/DB_HOST from the environment: %w", err)
	}

	hmacSecret := config.RequireSecret("HMAC_SECRET")
	jwtSecret := config.RequireSecret("JWT_SECRET")
	svc := auth.NewService(repository.New(db), nil, auth.Config{
		HMACSecret:  []byte(hmacSecret),
		TokenConfig: crypto.TokenConfig{Secret: []byte(jwtSecret), AccessExpiry: time.Minute},
	})

	// Actor sintético: usbictl corre con acceso directo al servidor —
	// es el mismo nivel de confianza que ya exige tocar la base a mano hoy
	// (backend/sql/01_seed_primer_admin.sql), solo que ahora pasa por las
	// mismas validaciones que la ruta HTTP en vez de un INSERT suelto.
	bootstrapActor := domain.JWTClaims{Role: domain.RoleAdmin}

	resp, err := svc.CreateAdminAccount(ctx, bootstrapActor, auth.AdminCreateAccountRequest{
		Nickname: *nickname,
		Password: password,
		Role:     domain.UserRole(*role),
	})
	if err != nil {
		return fmt.Errorf("creating account: %w", err)
	}

	fmt.Printf("cuenta creada: id=%s nickname=%s rol=%s alias=%q\n",
		resp.ID, resp.Nickname, resp.Role, resp.DisplayAlias)
	return nil
}

// adminBootstrapPassword lee ADMIN_BOOTSTRAP_PASSWORD del entorno (la
// escribe `usbictl secrets init`) o, si no está, la pide interactivamente
// sin eco de terminal — la única credencial de este flujo que una persona
// debe teclear (M3.6): nadie más la recuerda ni la vuelve a escribir después.
func adminBootstrapPassword() (string, error) {
	if pw := os.Getenv("ADMIN_BOOTSTRAP_PASSWORD"); pw != "" {
		return pw, nil
	}
	fmt.Fprint(os.Stderr, "Contraseña del administrador (no se mostrará en pantalla): ")
	raw, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("reading password: %w", err)
	}
	password := strings.TrimSpace(string(raw))
	if password == "" {
		return "", errors.New("la contraseña no puede estar vacía")
	}
	return password, nil
}
