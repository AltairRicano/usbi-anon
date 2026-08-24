# 02 — Backend: qué se copia, qué se reescribe, qué firma habla con qué base

Documento de detalle de [`00_Plan_maestro.md`](00_Plan_maestro.md). Cubre las
fases **F2–F4**. Todo lo que sigue está verificado contra el código real de
`../usbi/backend` (9 593 líneas Go), no contra su documentación.

## 1. El cambio de fondo, en una frase

`main.go` de `../usbi` abre **un** `*sql.DB` y construye **un**
`repository.Queries` que se inyecta a los seis servicios. USBI-Anon abre
**dos** pools, con **dos credenciales distintas**, y los servicios se reparten:
los que tocan progreso reciben el pool principal, los que tocan credenciales
reciben el pool de identidad, y exactamente **dos** servicios reciben ambos.

## 2. Qué se copia y qué se ignora, repo por repo

### 2.1 De `../usbi/backend` — **copia verbatim** (≈62 % del backend)

Estos paquetes **no contienen una sola referencia a `users`** — verificado con
`grep -rn 'users\b' internal/levels/ internal/sync/ internal/devices/ internal/incidents/`,
que devuelve **cero** resultados. Operan solo con `uuid.UUID`, que es
precisamente lo que USBI-Anon quiere. Se copian sin editar una línea:

| Paquete | LOC | Qué hace |
|---|---|---|
| `internal/levels` | 1 874 | Secciones, niveles, `POST /levels/{id}/complete`, progreso de perfil |
| `internal/sync` | 441 | Sync offline con HMAC, idempotencia, merge aditivo de XP |
| `internal/devices` | 156 | Registro y listado de dispositivos |
| `internal/incidents` | 188 | Bitácora de incidentes de seguridad |
| `internal/crypto` | 181 | Argon2id, HMAC, índice ciego, JWT |
| `internal/dbmaint` | 95 | Planificador de particiones futuras |
| `internal/mailer` | 186 | Envío de correo (lo usa el doble opt-in del tutor) |
| `internal/audit` | 74 | Helper de bitácora |
| `internal/httpjson`, `httpproblem`, `httputil` | 96 | Decodificación, RFC 7807, IP de cliente |

Del repositorio se copian tal cual los archivos que solo tocan la base
principal: `content_queries.go` (526), `badge_queries.go` (82),
`device_queries.go` (114), `security_incident_queries.go` (35), `db.go`,
`doc.go`, y la parte de `query.sql.go`/`models.go` referida a progreso y
contenido.

### 2.2 De `../usbi/backend` — **se reescribe**

| Paquete | LOC | Por qué |
|---|---|---|
| `internal/auth` | 1 335 | Habla con las dos bases: credenciales en identidad, alta de `accounts` en principal |
| `internal/maintenance` | 166 | Lee candidatos en identidad, purga en principal |
| `internal/repository` (parte auth/privacy/tutor) | ~415 | Se muda íntegra al paquete nuevo `internal/identityrepo` |
| `internal/transport/router.go` | 543 | `jwtAuthMiddleware` valida `token_version` contra identidad |
| `internal/config` | 162 | Necesita un segundo DSN |
| `internal/domain/models.go` | 135 | `User.FullName` desaparece del DTO; entra `DisplayAlias string` |
| `main.go` | 267 | Dos pools, dos `Queries`, dos health checks |
| `cmd/create_admin` | 81 | Escribe en las dos bases |
| `internal/testdb` | 163 | Necesita dos bases de prueba |

### 2.3 De `../usbi/backend` — **se ignora**

- `internal/api/` y `cmd/server/`: directorios **vacíos**. No se replican.
- `migrations/0002`–`0012`: se consolidan en el baseline (ver
  [`01_Base_de_datos.md` §1](01_Base_de_datos.md)). No se copian como archivos.
- Las 5 funciones PL/pgSQL de `0006`: ya estaban muertas y `0012` las eliminó.
  **No resucitarlas.** Go es la única fuente de verdad de XP y ARCO.
- `full_name`, `phone`, `phone_lookup_hash` y todo su plumbing en Go y SQL.
- `apps/`, `packages/`, `node_modules/`, `.pnpm-store/` de la raíz de `../usbi`:
  andamiaje de workspace, no aportan nada al backend.
- `AUDITORIA_PREPARACION_PRODUCCION.md`: útil como lectura, no como código.

### 2.4 De `../maker` — qué se toma y qué **no**

**Se toma únicamente el patrón, no el código:**

- La **convención de migraciones** (golang-migrate, `.up.sql`/`.down.sql`, SQL
  plano en español, baseline consolidado).
- La **forma de la tabla sin PII**: `maker.admin_users` (id + username + hash
  Argon2id, nada más) valida el patrón de `accounts` de
  [`01_Base_de_datos.md` §3.1](01_Base_de_datos.md), y `sections`/`levels`
  referenciando `admin_users(id)` valida que el contenido se ancle a una tabla
  local pobre en datos.

**No se toma nada de código.** `maker` no tiene login persistente, ni XP, ni
insignias, ni rachas, ni sync offline, ni ARCO. Su `internal/session` (token
opaco con TTL, progreso que cae en cascada al expirar) es **el modelo opuesto**
al que necesita USBI-Anon, que exige progreso persistente entre dispositivos.
Copiarlo sería un error de diseño, no un atajo.

---

## 3. Firmas que cambian y contra qué base habla cada una

Notación: **[I]** = base de identidad, **[P]** = base principal.

### 3.1 Repositorios

```text
ANTES  repository.New(db *sql.DB) *repository.Queries          // una sola base

DESPUÉS
       repository.New(mainDB *sql.DB)    *repository.Queries       // [P]
       identityrepo.New(identDB *sql.DB) *identityrepo.Queries     // [I]  ← paquete nuevo
```

`repository.DBTX` ya es una interfaz, así que el paquete nuevo reutiliza el
mismo contrato sin inventar nada.

### 3.2 Qué consulta se muda a `internal/identityrepo` **[I]**

Migran íntegras desde `../usbi/backend/internal/repository/`:

| Origen | Consultas |
|---|---|
| `auth_queries.go` | `InsertRefreshToken`, `GetRefreshTokenUser`, `RevokeRefreshToken`, `RevokeRefreshTokensForUser`, `PurgeExpiredRefreshTokens` |
| `query.sql.go` | `CreateUser` → `CreateIdentity`, `GetUserByEmailHash` → `GetIdentityByEmailHash`, `GetUserTokenVersion`, `IncrementTokenVersion`, `UpdateUserAdultStatus`, `IncrementAgeUpAttempts` |
| `tutor_consent_queries.go` | Las 5 consultas del doble opt-in, completas |
| `privacy_queries.go` (parte) | `InsertTutorConsent`, `ActivateTutorConsentUser`, `PseudonymizeUser`, `PseudonymizeTutorConsents`, y las 4 de `arco_requests` |
| `maintenance_queries.go` | `ListPendingTutorConsentUsers`, `SuspendInactivePlayers`, `ListSuspendedUsersForCancellation` — las tres leen `users` |

`GetRefreshTokenUser` **cambia de forma**: hoy devuelve `FullName` en su
`SELECT`. En USBI-Anon esa columna no existe; el struct `RefreshTokenUser`
pierde el campo.

**Nombres de columna:** en la base principal `user_id` **se conserva** (y
`actor_user_id` en `admin_audit_log`), exactamente como en `../usbi`.
Renombrarlo a `account_id` en 12 tablas obligaría a editar `content_queries.go`,
`badge_queries.go`, `device_queries.go` y `query.sql.go` —que este plan da por
copiados verbatim— a cambio de cero ganancia de seguridad. Lo que cambió no es
el nombre sino el contenido, y eso queda documentado con `COMMENT ON COLUMN` en
el propio esquema. Lo mismo aplica en la base de identidad.

### 3.3 Qué se queda en `internal/repository` **[P]**

`InsertExperienceHistory`, `InsertLevelAttempt`, `UpsertPlayerProgress`,
`UpsertDailyStreak`, `GetLevelAttemptsByDate`, `InsertSyncEvent`,
`UpdateSyncEventStatus`, `CreateLevel`, `CreateSection`, `ListPublishedLevels`,
`LogAdminAudit`, y las de insignias, dispositivos e incidentes.

De `privacy_queries.go` se quedan en **[P]** exactamente dos:
`NullUserInPseudonymizableLedgers` y `PurgeUserProgressData`.

**Nuevas en [P]:**

- `UpsertAccount(ctx, UpsertAccountParams{ID, Role, Status, IsAdult, AliasAdjectiveID, AliasNounID, AliasNumber})`
  — alta/refresco de la réplica. En el `ON CONFLICT DO UPDATE` **no debe tocar
  las tres columnas de alias**: el alias se sortea una sola vez, en el alta, y
  debe permanecer estable para la persona usuaria.
- `GetAccountAlias(ctx, userID) (string, error)` — lee la vista
  `account_aliases`, que compone `"Jaguar Azul 42"` sin que ninguna tabla
  almacene la cadena.
- `MarkDevicesForWipe(ctx, userID)`.

Los vocabularios `alias_adjectives` / `alias_nouns` tienen 24 palabras cada uno
(576 000 combinaciones con el número). El rol de aplicación **no tiene permiso
de escritura sobre ellos** (ver `backend/sql/00_roles_principal.sql`), así que
Go sortea índices dentro de rangos conocidos, nunca inserta palabras. El alias
no es único ni es un identificador: no debe usarse jamás como clave de
búsqueda.

### 3.4 Constructores de servicio

```text
ANTES                                          DESPUÉS
auth.NewService(q, cfg)                        auth.NewService(ident, main, cfg)   [I]+[P]
maintenance.NewService(q, cfg)                 maintenance.NewService(ident, main, cfg)  [I]+[P]
levels.NewService(q)                           levels.NewService(main)      [P]  — sin cambios de código
sync.NewService(q, hmacSecret)                 sync.NewService(main, hmacSecret)   [P]  — sin cambios
devices.NewService(q)                          devices.NewService(main)     [P]  — sin cambios
incidents.NewService(q, hmacSecret)            incidents.NewService(main, hmacSecret)  [P]  — sin cambios
```

Los cuatro últimos **no cambian de código**: solo reciben un pool distinto en
`main.go`. Esa es la ganancia de haber verificado que no tocan `users`.

### 3.5 Router y middleware

```text
ANTES   transport.RouterDependencies{ ..., Queries *repository.Queries }
        jwtAuthMiddleware(cfg crypto.TokenConfig, queries *repository.Queries)

DESPUÉS transport.RouterDependencies{ ..., IdentityQueries *identityrepo.Queries }
        jwtAuthMiddleware(cfg crypto.TokenConfig, ident *identityrepo.Queries)
```

El middleware valida `token_version` contra **[I]**, no contra **[P]**.

> **Decisión de rendimiento deliberada.** Se evaluó que el middleware
> garantizara también la existencia de la fila en `accounts` en cada request:
> eso duplicaría las consultas por request (una por base). En vez de eso, el
> alta/refresco de `accounts` ocurre **solo en `Login` y `Refresh`**, que ya
> escriben. Resultado: **una consulta por request autenticado, igual que hoy**,
> y la fila de `accounts` se autorrepara en el siguiente login si faltara.

### 3.6 Configuración y arranque

```text
config.DatabaseURL()          → sigue existiendo, ahora es la base PRINCIPAL   [P]
config.IdentityDatabaseURL()  → NUEVA                                          [I]
```

Variables de entorno nuevas: `IDENT_DATABASE_URL` (o el juego
`IDENT_DB_USER`/`IDENT_DB_PASSWORD`/`IDENT_DB_HOST`/`IDENT_DB_PORT`/`IDENT_DB_NAME`/`IDENT_DB_SSLMODE`).
Se reutilizan `RequireSecret`, `CheckConnPoolBounds` y el aviso de
`DB_SSLMODE=disable` con host no-loopback tal cual — ese aviso **importa más
aquí**, porque la base de identidad razonablemente estará en otra máquina.

`main.go` abre dos pools, hace `Ping` a los dos, y `/health/ready` **debe
comprobar ambos** y reportar cuál falló.

`PGP_ENCRYPTION_KEY` y `BLIND_INDEX_SECRET` pasan a ser secretos **solo del
lado identidad**; el pool principal ya no cifra nada.

---

## 4. Ciclo de vida de `accounts` (la réplica)

| Evento | Qué pasa |
|---|---|
| Registro | Escribe solo **[I]**. No se crea nada en **[P]** |
| Primer login | **[I]** valida credenciales → **[P]** `INSERT INTO accounts ... ON CONFLICT (id) DO UPDATE SET role, status, is_adult, updated_at` |
| Login/refresh posteriores | El mismo upsert refresca la réplica |
| Cambio de rol o suspensión | **[I]** es la verdad. El bloqueo es inmediato vía `token_version + 1` (el JWT muere en el siguiente request). La réplica se pone al día en el siguiente login |
| Cancelación ARCO | Ver §5 |

El upsert es idempotente por construcción, así que un fallo parcial nunca deja
el sistema en estado inválido: se corrige solo en el siguiente login.

---

## 5. La saga ARCO — el único punto genuinamente difícil

Hoy `auth.Service.ResolveArcoRequest` (`service.go:571`) hace todo esto dentro
de **una sola `sql.Tx`**: leer la solicitud `FOR UPDATE`, seudonimizar el
usuario y los consentimientos de tutor, revocar refresh tokens, poner a NULL el
actor en las dos bitácoras, purgar el progreso, marcar dispositivos para wipe y
registrar la acción en auditoría.

Con dos bases eso **deja de ser posible**. PostgreSQL no ofrece transacción
distribuida usable aquí (`PREPARE TRANSACTION` existe, pero exige
`max_prepared_transactions > 0`, un coordinador y limpieza manual de
transacciones colgadas — inasumible sobre un contenedor de ~256 MB).

**Diseño elegido: saga reanudable con estado persistido.**

```text
pending
  └─► purging_main            [P]  purga progreso, SET NULL en bitácoras,
  │                                marca dispositivos con wipe_local_data
  └─► identity_pseudonymized  [I]  seudonimiza identidad y tutores,
  │                                revoca refresh tokens, token_version + 1
  └─► resolved                [I]  cierra arco_requests
```

Reglas que hacen esto correcto:

1. **La base principal va primero.** La seudonimización de identidad es la que
   bloquea a la persona usuaria; si se hiciera primero y el proceso muriera, la
   cuenta quedaría muerta con su progreso intacto y sin sesión para reintentar.
   Al revés, un fallo deja una cuenta viva con progreso ya purgado: recuperable
   y reintentable.
2. **Cada paso es idempotente.** `PurgeUserProgressData` son `DELETE ... WHERE
   user_id = $1`; `PseudonymizeUser` lleva `AND deleted_at IS NULL`. Reintentar
   cualquiera de los dos no hace daño.
3. **El estado se persiste en `arco_requests.status` [I]** después de cada
   paso. Un reintento arranca desde donde quedó.
4. **La seudonimización conserva `identities.id`**, así que el UUID siempre
   sigue disponible para reintentar el lado principal. La saga nunca se queda
   sin la llave para terminar.
5. **Un job de reconciliación** (dentro del planificador de `maintenance`, que
   ya existe y ya corre cada 24 h) rebarre solicitudes atascadas en un estado
   intermedio.

El mismo patrón aplica a la cancelación automática por inactividad de
`maintenance.cancelUser` (`service.go:112`).

---

## 6. Orden de trabajo por eficiencia

De mayor a menor relación resultado/esfuerzo:

1. ~~**F1 — SQL**~~ ✅ **hecha y validada.** Ver
   [`01_Base_de_datos.md` §7 y §8](01_Base_de_datos.md).
2. ~~**F3 — copia verbatim.**~~ ✅ **hecha.** ≈3 300 líneas de Go movidas sin
   editar una línea (verificado con `diff`), más `internal/repository`
   reconstruido para la base principal. Da un backend que compila y pasa sus
   pruebas unitarias antes de tocar lo difícil.
3. ~~**F2 — `internal/identityrepo`.**~~ ✅ **hecha.** ~415 líneas, casi todas
   mudanza literal de consultas ya escritas y auditadas.
4. **F4 — reescritura de `auth`, `maintenance`, `transport`, `main.go`.** Aquí
   está el trabajo real. Dentro de esta fase, la saga ARCO (§5) es el único
   punto que merece diseño cuidadoso; el resto es fontanería de dos pools.
5. **F5 — frontend.** 6 archivos. Ver [`03_Frontend.md`](03_Frontend.md).

**Anti-patrón a evitar:** empezar por `auth` porque "es lo primero que corre el
usuario". Es el paquete más acoplado a las dos bases y el más caro de rehacer
sin tener antes el esquema cerrado y el repositorio partido.

---

## 7. Mejoras propuestas y con qué skill se atacan

| # | Mejora | Skill / agente | Cuándo |
|---|---|---|---|
| 1 | Diseñar el corte de paquetes y los contratos entre los dos repositorios antes de escribir código | agente **`backend-architect`** o **`the-architect:architect-brownfield`** | Inicio de F2 |
| 2 | Blueprint con criterios de aceptación por paso, validado | **`the-architect:architect-brownfield`** + **`the-architect:architect-audit`** | F2 |
| 3 | Evitar que la migración se convierta en refactor general del backend | agente **`minimal-change-engineer`** | Durante F3–F4 |
| 4 | Auditoría de seguridad del aislamiento real entre pools, roles y secretos | **`cyber-neo`** + agente **`security-engineer`** | Cierre de F4 |
| 5 | Revisión de la saga ARCO (idempotencia, reanudabilidad, orden de pasos) | agente **`code-reviewer`** con foco explícito en §5 | Cierre de F4 |
| 6 | Puerta de diff antes de cada commit de fase | **`/security-review`** y **`/code-review`** | Cada fase |
| 7 | Resumen masivo o segunda opinión sin gastar cuota propia | **`agy-delegate`** (modo `strict`) | Cualquier momento |

> Regla 5 del `CLAUDE.md` global: a cualquier agente que se lance hay que
> pedirle explícitamente que escriba su reporte en su propia sección de
> `estado_proyecto.md`.

---

## 8. Criterios de aceptación del backend

1. `grep -rn 'FullName\|full_name\|phone' backend/ --include=*.go` no devuelve
   nada fuera de comentarios.
2. Ningún archivo de `internal/repository` menciona `identities`,
   `refresh_tokens`, `tutor_consent*` ni `arco_requests`.
3. Ningún archivo de `internal/identityrepo` menciona `player_progress`,
   `level_attempts`, `experience_history`, `daily_streak` ni `accounts`.
4. `levels`, `sync`, `devices`, `incidents`, `crypto`, `dbmaint`, `mailer`,
   `audit`, `httpjson`, `httpproblem`, `httputil` tienen **diff cero** contra
   `../usbi` salvo la línea `module` de los imports.
5. `/health/ready` falla si **cualquiera** de las dos bases no responde, e
   indica cuál.
6. La saga ARCO reanuda correctamente desde cada estado intermedio (prueba de
   integración que mata el proceso entre pasos).
7. El número de consultas por request autenticado sigue siendo **una**.
