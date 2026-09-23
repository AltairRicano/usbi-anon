---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Centraliza la carga y validación de variables de entorno del servidor. Construye DSNs aislados de PostgreSQL para distintos roles (jugador, moderador, mantenimiento y migraciones), advirtiendo si la conexión no usa SSL en entornos remotos y exigiendo longitud mínima en secretos criptográficos. Lo consume `usbictl` ([[backend/cmd/usbictl/main.go.md#main|main.go]] y [[backend/cmd/usbictl/migrate.go.md#runMigrate|migrate.go]]) para resolver los cuatro DSN por rol y validar secretos antes de operar sobre la base de datos.

## Funciones

### LoadEnvironment
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Carga el archivo de entorno desde la ruta definida en USBI_BACKEND_ENV_FILE o el archivo .env por defecto, omitiendo errores fatales si no existe para permitir configuración nativa en producción.

### DatabaseURL
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Construye la cadena de conexión DSN a PostgreSQL para el pool de usuarios/jugadores (rol usbi_app) a partir de DATABASE_URL o variables DB_* individuales.

### ModeratorDatabaseURL
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera la cadena DSN para el pool de moderación (rol usbi_moderador), manteniendo aislamiento de credenciales respecto al pool estándar de jugadores.

### DBMaintDatabaseURL
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera la cadena DSN para el pool de mantenimiento de particiones (rol usbi_dbmaint), con permisos restringidos exclusivamente a la ejecución de particionado anual.

### MigrateDatabaseURL
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Construye la cadena DSN para el rol propietario del esquema (usbi_migrate), orientada únicamente al uso por herramientas CLI de migración de base de datos.

### databaseURLFromEnv
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera una URL DSN estructurada a partir del conjunto de variables provisto y advierte en logs si sslmode=disable es utilizado fuera de interfaces loopback.

### RequireEnv
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Lee una variable de entorno obligatoria y termina la ejecución de manera fatal si no está configurada o se encuentra vacía.

### RequireSecret
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida la existencia de una variable de entorno y exige que tenga una longitud mínima de 32 bytes para prevenir la ejecución del sistema con claves criptográficas débiles.

### CheckConnPoolBounds
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Comprueba que el número de conexiones inactivas configuradas no exceda el límite de conexiones abiertas del pool de base de datos, deteniendo el servidor si la regla se viola.
