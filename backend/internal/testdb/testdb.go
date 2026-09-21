// Package testdb provee bases de datos Postgres reales y aisladas para
// pruebas de integración. Permite que paquetes como internal/auth, internal/maintenance
// e internal/privacy prueben sus caminos transaccionales y de seguridad contra SQL real.
//
// Setup crea un esquema Postgres con nombre aleatorio dentro de la instancia
// apuntada por TEST_DATABASE_URL, aplica las migraciones de backend/migrations/,
// y lo elimina en t.Cleanup.
//
// Es seguro apuntar TEST_DATABASE_URL a una base de desarrollo con datos en "public":
// estas pruebas nunca leen ni escriben fuera de su esquema desechable.
//
// TEST_DATABASE_URL es opt-in: si no está definida, Setup llama a t.Skip para
// que las pruebas continúen pasando en entornos sin Postgres alcanzable.
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

	"github.com/altair/usbi-anon-backend/internal/repository"
	"github.com/lib/pq"
)

// DB agrupa las Queries y la conexión cruda de la base de prueba. DB solo
// hace falta para pruebas que necesiten aserciones SQL directas, ya que
// repository.Queries no expone un método de consulta ad-hoc.
type DB struct {
	Repo *repository.Queries
	DB   *sql.DB
}

// migrationsDir es la raíz backend/migrations, resuelta relativa a este
// archivo fuente (no al directorio de trabajo del llamador), así que Setup
// funciona igual sin importar qué paquete importe testdb.
func migrationsDir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations")
}

// Setup crea un esquema aislado, aplica el baseline unificado, y devuelve
// Queries + *sql.DB. El pool queda anclado al esquema vía el parámetro de
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

	db := setupSchema(t, admin, kvDSN, "usbi_test_")

	return &DB{
		Repo: repository.New(db),
		DB:   db,
	}
}

// setupSchema crea un esquema con prefijo+hex aleatorio, lo registra para
// borrarse en t.Cleanup, abre un pool anclado a él vía search_path, y le
// aplica las migraciones *.up.sql de backend/migrations/.
func setupSchema(t *testing.T, admin *sql.DB, kvDSN, schemaPrefix string) *sql.DB {
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

	applyMigrations(t, db, migrationsDir())
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
	sort.Strings(files) // los nombres de archivo tienen ceros a la izquierda (0001_, 0002_, ...): orden léxico == orden deseado.

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

// upOnly elimina la sección "-- +goose Down" y todo lo que sigue.
// Ninguna migración de USBI-Anon lleva anotaciones goose, así que hoy es un no-op
// — se conserva por si alguna migración futura las incluyera.
func upOnly(raw string) string {
	if idx := strings.Index(raw, "-- +goose Down"); idx != -1 {
		return raw[:idx]
	}
	return raw
}

// toKeywordDSN normaliza ya sea una URL postgres:// o un DSN en formato
// keyword=value a la forma keyword=value, para que los invocadores puedan añadir ` options=...`.
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
		panic(err) // un fallo en crypto/rand es irrecuperable
	}
	return hex.EncodeToString(b)
}
