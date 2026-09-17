package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/altair/usbi-anon-backend/internal/config"
	_ "github.com/lib/pq"
	"github.com/google/uuid"
)

// poolCheck describe un pool a verificar: cómo conectarse y qué probar.
type poolCheck struct {
	name string
	dsn  func() string
	// allowed corre dentro de una transacción que SIEMPRE se revierte
	// (nunca deja datos de prueba) y debe tener éxito.
	allowed []namedStmt
	// forbidden debe FALLAR — si tiene éxito, doctor lo reporta como el
	// hallazgo más grave posible: un GRANT de más que nadie notó.
	forbidden []namedStmt
}

type namedStmt struct {
	label string
	sql   string
	args  []any
}

// newSuggestionArgs genera los argumentos de un INSERT de prueba sobre
// suggestions, con un id fresco por llamada (nunca reutiliza uno real ni dos
// veces el mismo, aunque de todas formas la transacción siempre se revierte).
func newSuggestionArgs() []any {
	return []any{uuid.New(), "usbictl doctor: fila de prueba, revertida siempre", 0, 0}
}

// runDoctor implementa `usbictl doctor` (M3.5 punto 4): por cada uno de los
// tres pools de aplicación, confirma identidad, ejecuta una operación
// representativa de lo que SÍ debe poder hacer, y confirma que lo prohibido
// sigue prohibido. Un doctor que solo comprobara lo permitido no detectaría
// un GRANT de más — la segunda mitad es la que evita que una migración
// futura afloje la matriz sin que nadie lo note.
func runDoctor(_ []string) error {
	checks := []poolCheck{
		{
			name: "usbi_app (jugador)",
			dsn:  config.DatabaseURL,
			allowed: []namedStmt{
				{label: "SELECT sobre accounts", sql: "SELECT 1 FROM accounts LIMIT 1"},
				{
					label: "INSERT sobre suggestions (sin RETURNING: usbi_app no tiene SELECT ahí)",
					sql:   "INSERT INTO suggestions (id, description, levels_completed_snapshot, xp_snapshot) VALUES ($1, $2, $3, $4)",
					args:  newSuggestionArgs(),
				},
			},
			forbidden: []namedStmt{
				{label: "SELECT sobre security_incidents", sql: "SELECT 1 FROM security_incidents LIMIT 1"},
				{label: "UPDATE sobre audit_log", sql: "UPDATE audit_log SET action = action WHERE false"},
			},
		},
		{
			name: "usbi_moderador (contenido y administración)",
			dsn:  config.ModeratorDatabaseURL,
			allowed: []namedStmt{
				{label: "SELECT sobre audit_log", sql: "SELECT 1 FROM audit_log LIMIT 1"},
				{label: "SELECT sobre sections", sql: "SELECT 1 FROM sections LIMIT 1"},
			},
			forbidden: []namedStmt{
				{label: "DELETE sobre audit_log (append-only)", sql: "DELETE FROM audit_log WHERE false"},
				{
					label: "INSERT sobre suggestions (buzón anónimo, solo usbi_app inserta)",
					sql:   "INSERT INTO suggestions (id, description, levels_completed_snapshot, xp_snapshot) VALUES ($1, $2, $3, $4)",
					args:  newSuggestionArgs(),
				},
			},
		},
		{
			name: "usbi_dbmaint (mantenimiento de particiones)",
			dsn:  config.DBMaintDatabaseURL,
			allowed: []namedStmt{
				{label: "EXECUTE ensure_yearly_partition", sql: "SELECT ensure_yearly_partition('level_attempts', EXTRACT(YEAR FROM NOW())::integer)"},
			},
			forbidden: []namedStmt{
				{label: "SELECT sobre levels (sin ningún GRANT de tabla)", sql: "SELECT 1 FROM levels LIMIT 1"},
			},
		},
	}

	overallOK := true
	for _, c := range checks {
		ok := runPoolCheck(c)
		overallOK = overallOK && ok
	}

	if !overallOK {
		fmt.Fprintln(os.Stderr, "\ndoctor: uno o más pools fallaron — ver detalle arriba")
		os.Exit(1)
	}
	fmt.Println("\ndoctor: todos los pools verificados correctamente")
	return nil
}

func runPoolCheck(c poolCheck) bool {
	fmt.Printf("== %s ==\n", c.name)
	db, err := sql.Open("postgres", c.dsn())
	if err != nil {
		fmt.Printf("  ✗ no se pudo abrir la conexión: %v\n", err)
		return false
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var currentUser, sessionUser string
	if err := db.QueryRowContext(ctx, "SELECT current_user, session_user").Scan(&currentUser, &sessionUser); err != nil {
		fmt.Printf("  ✗ SELECT current_user, session_user: %v\n", err)
		return false
	}
	fmt.Printf("  ✓ conectado como current_user=%s session_user=%s\n", currentUser, sessionUser)

	ok := true
	for _, stmt := range c.allowed {
		if err := execInRolledBackTx(ctx, db, stmt.sql, stmt.args...); err != nil {
			fmt.Printf("  ✗ %s: se esperaba que funcionara y falló: %v\n", stmt.label, err)
			ok = false
			continue
		}
		fmt.Printf("  ✓ permitido: %s\n", stmt.label)
	}

	for _, stmt := range c.forbidden {
		if err := execInRolledBackTx(ctx, db, stmt.sql, stmt.args...); err == nil {
			fmt.Printf("  ✗ %s: se esperaba que FALLARA y tuvo éxito — GRANT de más\n", stmt.label)
			ok = false
			continue
		}
		fmt.Printf("  ✓ prohibido correctamente: %s\n", stmt.label)
	}

	return ok
}

// execInRolledBackTx corre `query` dentro de una transacción que siempre se
// revierte — doctor nunca debe dejar datos de prueba (ni siquiera los de un
// INSERT "permitido") en una base real.
func execInRolledBackTx(ctx context.Context, db *sql.DB, query string, args ...any) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return err
	}
	return nil
}
