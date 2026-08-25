// Package testdb provides isolated, real-Postgres test databases for
// integration tests (audit finding B9, heredado de ../usbi). Existe para que
// internal/auth, internal/maintenance y internal/privacy puedan probar sus
// caminos transaccionales y de seguridad (saga ARCO, seudonimización,
// auditoría) contra SQL real — nunca mockeado.
//
// A diferencia de ../usbi/backend/internal/testdb, USBI-Anon necesita DOS
// bases aisladas, no una: Setup crea dos esquemas Postgres nuevos con nombre
// aleatorio dentro de la MISMA instancia que apunta TEST_DATABASE_URL — uno
// para migrations/identity, otro para migrations/main — y los borra en
// t.Cleanup. Es el mismo patrón de aislamiento por esquema desechable que ya
// se usó para verificar F1 (sqlcheck_ident/sqlcheck_main, ver
// estado_proyecto.md), solo que ahora vive en código reutilizable en vez de
// ser un paso manual de verificación.
//
// Es seguro apuntar TEST_DATABASE_URL a una base de desarrollo real que ya
// tenga datos en "public": estas pruebas nunca leen ni escriben fuera de sus
// dos esquemas desechables.
//
// TEST_DATABASE_URL es opt-in a propósito: si no está definida, Setup llama
// a t.Skip para que `go test ./...` siga en verde en cualquier entorno sin
// Postgres alcanzable (runners de CI, otras máquinas, etc).
package testdb

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/altair/usbi-anon-backend/internal/identityrepo"
	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/lib/pq"
)

// DB agrupa las Queries y la conexión cruda de las DOS bases de prueba.
// MainDB/IdentDB solo hacen falta para pruebas que necesiten aserciones SQL
// directas, ya que ni repository.Queries ni identityrepo.Queries exponen un
// método de consulta ad-hoc.
type DB struct {
	Ident   *identityrepo.Queries
	IdentDB *sql.DB
	Main    *repository.Queries
	MainDB  *sql.DB
}

// migrationsDir es la raíz backend/migrations, resuelta relativa a este
// archivo fuente (no al directorio de trabajo del llamador), así que Setup
// funciona igual sin importar qué paquete importe testdb.
func migrationsDir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations")
}

// Setup crea dos esquemas aislados (identidad y principal), aplica las
// migraciones correspondientes en cada uno, y devuelve Queries + *sql.DB
// para ambas. Cada pool queda anclado a su esquema vía el parámetro de
// conexión `options=-c search_path=...`, así que es seguro bajo consultas
// concurrentes dentro de una misma prueba, no solo con una única conexión.
func Setup(t *testing.T) *DB {
	t.Helper()

	rawDSN := os.Getenv("TEST_DATABASE_URL")
	if rawDSN == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test (see internal/testdb doc comment)")
	}

	kvDSN, err := toKeywordDSN(rawDSN)
	if err != nil {
		t.Fatalf("testdb: parsing TEST_DATABASE_URL: %v", err)
	}

	admin, err := sql.Open("postgres", kvDSN)
	if err != nil {
		t.Fatalf("testdb: opening admin connection: %v", err)
	}
	t.Cleanup(func() { _ = admin.Close() })
	if err := admin.Ping(); err != nil {
		t.Fatalf("testdb: TEST_DATABASE_URL unreachable: %v", err)
	}

	identDB := setupSchema(t, admin, kvDSN, "usbi_test_ident_", "identity")
	mainDB := setupSchema(t, admin, kvDSN, "usbi_test_main_", "main")

	return &DB{
		Ident:   identityrepo.New(identDB),
		IdentDB: identDB,
		Main:    repository.New(mainDB),
		MainDB:  mainDB,
	}
}

// setupSchema crea un esquema con prefijo+hex aleatorio, lo registra para
// borrarse en t.Cleanup, abre un pool anclado a él vía search_path, y le
// aplica las migraciones *.up.sql de migrations/<migrationsSubdir>/.
func setupSchema(t *testing.T, admin *sql.DB, kvDSN, schemaPrefix, migrationsSubdir string) *sql.DB {
	t.Helper()

	schema := schemaPrefix + randomHex(8)
	if _, err := admin.Exec(fmt.Sprintf(`CREATE SCHEMA %s`, quoteIdent(schema))); err != nil {
		t.Fatalf("testdb: creating schema %s: %v", schema, err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, quoteIdent(schema))); err != nil {
			t.Logf("testdb: failed to drop schema %s (manual cleanup may be needed): %v", schema, err)
		}
	})

	scopedDSN := fmt.Sprintf("%s options='-c search_path=%s,public'", kvDSN, schema)
	db, err := sql.Open("postgres", scopedDSN)
	if err != nil {
		t.Fatalf("testdb: opening scoped connection for %s: %v", schema, err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("testdb: scoped connection for %s unreachable: %v", schema, err)
	}

	applyMigrations(t, db, filepath.Join(migrationsDir(), migrationsSubdir))
	return db
}

func applyMigrations(t *testing.T, db *sql.DB, dir string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("testdb: reading migrations dir %s: %v", dir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files) // filenames are zero-padded (0001_, 0002_, ...): lexical order == intended order.

	for _, name := range files {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("testdb: reading migration %s: %v", name, err)
		}
		if _, err := db.Exec(upOnly(string(raw))); err != nil {
			t.Fatalf("testdb: applying migration %s: %v", name, err)
		}
	}
}

// upOnly strips a goose "-- +goose Down" section and everything after it.
// Ninguna migración de USBI-Anon lleva anotaciones goose (nota histórica en
// el propio baseline), así que hoy es un no-op — se conserva por si alguna
// migración futura sí las trajera.
func upOnly(raw string) string {
	if idx := strings.Index(raw, "-- +goose Down"); idx != -1 {
		return raw[:idx]
	}
	return raw
}

// toKeywordDSN normalises either a postgres:// URL or an already-keyword=value
// DSN into keyword=value form, so callers can safely append ` options=...`.
func toKeywordDSN(dsn string) (string, error) {
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		return pq.ParseURL(dsn)
	}
	return dsn, nil
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func randomHex(n int) string {
	b := make([]byte, n/2)
	if _, err := rand.Read(b); err != nil {
		panic(err) // crypto/rand failing is unrecoverable
	}
	return hex.EncodeToString(b)
}
