---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Provee un entorno aislado de base de datos PostgreSQL para pruebas de integración mediante esquemas temporales con nombres aleatorios. Garantiza la seguridad de los datos preexistentes al confinar las operaciones dentro del esquema efímero mediante `search_path` y registrar su eliminación en `t.Cleanup`. Si `TEST_DATABASE_URL` no está configurada, omite la ejecución de las pruebas de forma segura sin marcar error. Usado por pruebas de integración como sync/integration_test.go; `Setup` construye el `*repository.Queries` llamando a `repository.New`, definida en [[backend/internal/repository/db.go.md|repository/db.go]].

## Funciones

### migrationsDir
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la ruta absoluta del directorio `backend/migrations` resuelta de forma relativa al archivo fuente mediante `runtime.Caller`, asegurando independencia del directorio desde el cual se ejecuten las pruebas.

### Setup
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Punto de entrada principal para preparar la base de datos de pruebas. Valida `TEST_DATABASE_URL` (omitiendo la prueba si no está definida), abre la conexión administrativa, invoca la creación del esquema aislado y retorna una estructura `DB` con el repositorio e instancia SQL.

### setupSchema
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Crea un esquema con nombre aleatorio de forma segura, programa su eliminación automática mediante `t.Cleanup`, establece un pool de conexiones acotado a dicho esquema usando `search_path` y aplica las migraciones requeridas.

### applyMigrations
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene los archivos de migración `.up.sql`, los ordena de forma léxica para asegurar el orden de ejecución correcto y los ejecuta secuencialmente sobre la base de datos aislada.

### upOnly
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Filtra el contenido SQL eliminando la sección de reversión `-- +goose Down` si estuviera presente, retornando únicamente las sentencias de aplicación.

### toKeywordDSN
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Normaliza cadenas de conexión DSN en formato URI o de par clave-valor al formato normalizado de clave-valor para permitir la posterior inclusión de parámetros de conexión adicionales.

### quoteIdent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Escapa comillas dobles y envuelve el identificador entre comillas para prevenir inyecciones SQL o errores sintácticos en nombres de esquema dinámicos.

### randomHex
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera una secuencia hexadecimal aleatoria mediante `crypto/rand` para garantizan la creación de nombres de esquemas únicos e impredecibles.
