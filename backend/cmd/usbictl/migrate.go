package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/altair/usbi-anon-backend/internal/config"
	"github.com/altair/usbi-anon-backend/migrations"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

// newMigrator conecta como usbi_migrate (dueño del esquema, ver
// backend/sql/00_roles_unificado.sql) contra las migraciones empotradas en
// el binario (backend/migrations, package migrations). Nunca como postgres
// superusuario: las tablas nacen con el dueño correcto y el bloque de GRANT
// correctivo que el entrypoint de la instancia de stress-test necesitaba
// deja de hacer falta (M3.3).
func newMigrator() (*migrate.Migrate, error) {
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("loading embedded migrations: %w", err)
	}
	dsn := config.MigrateDatabaseURL()
	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting as usbi_migrate: %w", err)
	}
	return m, nil
}

// runMigrate implementa `usbictl migrate up|down [N]|force <version>|version`.
// Deliberadamente un subcomando explícito, nunca ejecución automática al
// arrancar el servidor (M3.3): permite ver qué va a pasar antes de que pase,
// permite revertir, y no compite consigo mismo si algún día hay más de una
// instancia del backend arrancando a la vez.
func runMigrate(args []string) error {
	if len(args) == 0 {
		return errors.New("uso: usbictl migrate up|down [N]|force <version>|version")
	}
	m, err := newMigrator()
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()

	switch args[0] {
	case "up":
		err = m.Up()
	case "down":
		// Sin N explícito: un solo paso hacia abajo. "Todo abajo" por
		// accidente no debe ser posible con una sola invocación — repetir
		// el subcomando esa cantidad de veces es a propósito, no un
		// descuido de quien opera.
		steps := 1
		if len(args) >= 2 {
			n, convErr := strconv.Atoi(args[1])
			if convErr != nil || n <= 0 {
				return fmt.Errorf("N inválido para down: %q", args[1])
			}
			steps = n
		}
		err = m.Steps(-steps)
	case "force":
		if len(args) < 2 {
			return errors.New("uso: usbictl migrate force <version>")
		}
		version, convErr := strconv.Atoi(args[1])
		if convErr != nil {
			return fmt.Errorf("version inválida: %q", args[1])
		}
		err = m.Force(version)
	case "version":
		version, dirty, verErr := m.Version()
		if verErr != nil {
			return verErr
		}
		fmt.Printf("version=%d dirty=%v\n", version, dirty)
		return nil
	default:
		return fmt.Errorf("subcomando de migrate desconocido: %q", args[0])
	}

	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("sin cambios: el esquema ya está en la versión más reciente")
		return nil
	}
	if err != nil {
		return err
	}
	version, dirty, verErr := m.Version()
	if verErr == nil {
		fmt.Printf("ok: version=%d dirty=%v\n", version, dirty)
	}
	return nil
}
