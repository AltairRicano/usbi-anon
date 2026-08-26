# Plan maestro — Migración de USBI a USBI-Anon (dos bases de datos, identidad por UUID)

> **Fecha:** 2026-08-25
> **Estado:** **F4 cerrada** — `auth` (dto/service/handler), `maintenance`
> (con el job de reconciliación ARCO), `transport/router.go`, `main.go` (dos
> pools, dos health checks), `cmd/create_admin` (escribe en las dos bases) y
> `internal/testdb` (dos esquemas desechables) reescritos para el reparto en
> dos bases. Saga ARCO de cancelación implementada como checkpoint reanudable
> en `arco_requests.status`, con reclamo bajo `FOR UPDATE` y dos pruebas de
> integración que matan el proceso entre cada paso, corridas contra un
> esquema desechable en `usbi-database`. Compilado, `go vet`, `go test ./...`
> y `gofmt -l` en verde (Go 1.22); servidor arrancado de extremo a extremo
> contra dos bases desechables reales (registro→login→alias→`/health/ready`
> detectando la caída de cada base por separado). Ver
> [`02_Backend.md` §8](02_Backend.md) para el detalle de los 7 criterios de
> aceptación, incluidas dos excepciones documentadas al diff-cero del
> criterio 4.
>
> **⚠️ Los scripts SQL que describe este documento ya NO existen en el repo.**
> `backend/migrations/{identity,main}/` y `backend/sql/00_roles_{identidad,
> principal}.sql` se retiraron al cerrar **F5** del nuevo plan, sustituidos por
> `backend/migrations/0001_esquema_unificado.{up,down}.sql` y
> `backend/sql/00_roles_unificado.sql` (una sola base, sin correo electrónico,
> sin flujo de tutor). Lo que este documento dice sobre el esquema vale como
> historia de por qué se llegó ahí, no como descripción del repo actual.
>
> **⚠️ Cambio de rumbo tras F4 (misma fecha, sesión posterior):** el usuario
> decidió abandonar el modelo de dos bases + email + tutor por correo que
> describen este documento y `01_Base_de_datos.md`/`02_Backend.md` para todo
> lo relativo a **identidad**. El nuevo diseño (una sola base, registro por
> cuestionario de gustos → nickname/password generados, sin PII directa en
> absoluto) está cerrado y documentado en
> **[`04_Rediseno_identidad_gustos.md`](04_Rediseno_identidad_gustos.md)**,
> que define las fases **F5–F11** (reemplazan la F5/F6 originales de la tabla
> de abajo) y su orden de dependencia para poder paralelizarlas entre varios
> agentes/sesiones. Este documento y `01_Base_de_datos.md`/`02_Backend.md`
> siguen siendo la referencia válida para todo lo que **no** cambia (progreso,
> contenido, niveles, sync, dispositivos, saga ARCO en su forma general) —
> pero para identidad/auth, **`04_Rediseno_identidad_gustos.md` manda sobre
> este documento**, no al revés.
>
> **Fuente de verdad de este plan:** el código real de `../usbi` (migraciones aplicadas
> y verificadas contra el contenedor `usbi-database`), **no** el `plan/` de `../usbi`.

Este documento es el índice y el contrato de fases. El detalle vive en:

| Documento | Contenido |
|---|---|
| [`01_Base_de_datos.md`](01_Base_de_datos.md) | Qué tabla/columna se conserva, se va, se crea o se mueve. Scripts SQL a producir. Mejoras y skill asociada. |
| [`02_Backend.md`](02_Backend.md) | Qué paquetes Go se copian tal cual, cuáles se reescriben, qué firmas cambian, contra qué base habla cada una. |
| [`03_Frontend.md`](03_Frontend.md) | Los 6 archivos que cambian y por qué el resto no se toca. |
| [`05_Contenido_maker_y_juego.md`](05_Contenido_maker_y_juego.md) | **F10.6–F10.12**: poda de roles, motor de juego, maker local, panel de contenido con purga real, vista de jugador, borrado duro de cuenta y navegación. Manda sobre `03_Frontend.md` y sobre la §5 de `04` en todo lo de contenido y juego. |

---

## 1. El problema en una frase

En `../usbi`, **17 tablas viven en una sola base** (`usbi_db` del contenedor
`usbi-database`, verificado con `\dt`) y **17 claves foráneas apuntan a
`users(id)`**, una tabla que contiene `full_name`, `email`, `phone` y —
vía `tutor_consents` — nombre y correo del tutor de un menor.

USBI-Anon parte esa base en dos:

- **BD de identidad** (`usbi_ident_db`): correo + hash de contraseña + UUID.
  Único dato personal directo del sistema.
- **BD principal** (`usbi_anon_db`): todo el progreso, contenido y auditoría,
  indexado por ese UUID y **sin una sola columna identificable**.

PostgreSQL **no permite claves foráneas entre bases de datos distintas**. Esa
única restricción técnica es la que dispara casi todo lo demás en este plan:
las 17 FK a `users(id)` no pueden sobrevivir tal cual, la resolución ARCO deja
de ser una transacción única, y `main.go` deja de tener un solo `*sql.DB`.

## 2. Decisión de diseño central: tabla `accounts` en la BD principal

Se evaluaron dos formas de romper esas 17 FK:

**(A) Sin tabla ancla.** Cada tabla guarda `user_id UUID NOT NULL` sin FK.
Máximo desacoplamiento, pero se pierde integridad referencial, se pierden los
`ON DELETE CASCADE` que hoy implementan la purga ARCO, y nada impide filas
huérfanas de un UUID que nunca existió.

**(B) Tabla ancla local `accounts` (elegida).** La BD principal tiene su
propia tabla `accounts(id UUID PK, role, status, is_adult, created_at,
deleted_at)` — **sin ninguna columna identificable**. Las 17 FK sobreviven
cambiando únicamente su destino: `REFERENCES users(id)` →
`REFERENCES accounts(id)`.

Se elige **(B)** por cuatro razones concretas:

1. **Es el diff mínimo sobre el esquema existente**: cada FK cambia una
   palabra, no desaparece. Los `ON DELETE CASCADE`/`SET NULL` siguen
   funcionando exactamente igual.
2. **Preserva la maquinaria de purga ARCO** ya escrita y auditada en
   `../usbi/backend/internal/repository/privacy_queries.go`.
3. **Es el patrón que `../maker` ya usa y que funciona**: allí `sections` y
   `levels` referencian `admin_users(id)` — una tabla de credenciales sin PII.
   Aquí `accounts` es aún más pobre: ni siquiera tiene `username`.
4. **`role` y `status` deben estar en la BD principal de todos modos**: el
   panel admin filtra por rol y el motor de progreso rechaza cuentas
   suspendidas. Sin `accounts` habría que consultar la BD de identidad en cada
   request de progreso — justo el acoplamiento que este proyecto busca evitar.

**Costo aceptado:** `role`, `status` e `is_adult` quedan duplicados en ambas
bases. La **fuente de verdad es siempre la BD de identidad**; la BD principal
mantiene una réplica que se refresca en `login`/`refresh` (ver
[`02_Backend.md` §4](02_Backend.md)). La revocación inmediata no depende de esa
réplica: sigue funcionando por `token_version` contra la BD de identidad, igual
que hoy.

## 3. Fases

| # | Fase | Entregable | Bloquea a |
|---|---|---|---|
| **F0** | Decisiones abiertas | Respuestas a §4 de este documento | F1 |
| **F1** | **Scripts SQL** ✅ hecha, ⚠️ **retirada en F5** | `backend/migrations/{identity,main}/0001_*.{up,down}.sql` + `backend/sql/00_roles_*.sql` — **eliminados del repo**, ver `04_Rediseno_identidad_gustos.md` §1 | F2 |
| **F2** | **Esqueleto Go** ✅ **hecha** | Módulo nuevo, `config` con dos DSN, `internal/identityrepo` | F3 |
| **F3** | **Copia verbatim** ✅ **hecha** | `levels`, `sync`, `devices`, `incidents`, `crypto`, `httpjson`, `httpproblem`, `httputil`, `audit`, `mailer`, `dbmaint`, `domain` + `internal/repository` [P] | F4 |
| **F4** | **Reescritura** ✅ **hecha** | `auth`, `maintenance`, `transport`, `main.go`, `cmd/create_admin`, `internal/testdb` | F5 |
| ~~F5~~ | ~~Frontend~~ | Superseded — ver F5–F11 en [`04_Rediseno_identidad_gustos.md`](04_Rediseno_identidad_gustos.md) | — |
| ~~F6~~ | ~~Legal~~ | Renumerada a **F11** en el nuevo plan | — |

**Tabla de fases vigente para identidad/auth (reemplaza la fila F5/F6 de
arriba):** ver [`04_Rediseno_identidad_gustos.md` §6](04_Rediseno_identidad_gustos.md)
para F5 (esquema SQL unificado) a F11 (legal), con su grafo de dependencias
para trabajo en paralelo.

**F1 fue deliberadamente solo escritura de SQL.** Los scripts se validaron
aplicándolos a dos bases desechables (`sqlcheck_ident`, `sqlcheck_main`)
creadas y **eliminadas** en la misma sesión: `usbi_db` (17 tablas, 6 filas en
`users`) y `maker_db` quedaron intactas y el contenedor sigue con exactamente
las bases que tenía antes. Ninguna base del proyecto se levantó.

**F2 migró literalmente las consultas de identidad de `../usbi/backend/internal/repository`
a `backend/internal/identityrepo`** (auth, tutor consent, ARCO, mantenimiento —
ver [`02_Backend.md` §3.2](02_Backend.md)), adaptando solo lo que el esquema de
F1 cambió: tabla `users`→`identities`, y la eliminación de `full_name`/`phone`
en `CreateIdentity`, `GetIdentityByEmailHash`, `RefreshTokenUser` y
`PseudonymizeUser`. `internal/config` ganó `IdentityDatabaseURL()` junto a la
`DatabaseURL()` existente, factorizando el parseo común. Verificado con
`go build ./...`, `go vet ./...` y `gofmt -l .` (Go 1.22, contenedor
`golang:1.22-bookworm`, sin errores) y con los criterios 1 y 3 de
[`02_Backend.md` §8](02_Backend.md) (`grep` de `FullName`/`full_name`/`phone` y
de tablas de la base principal). No se tocó ningún paquete de `../usbi` ni se
levantó ninguna base de datos.

**F3 copió verbatim los 11 paquetes que el `grep` de
[`02_Backend.md` §2.1](02_Backend.md) confirma sin ninguna referencia a
`users`**, más `internal/domain` (necesario para que esos paquetes compilen;
su reescritura real — quitar `FullName`, añadir `DisplayAlias` — sigue siendo
tarea de F4, tal como dice [`02_Backend.md` §2.2](02_Backend.md)). Se
reconstruyó `internal/repository` para la base principal a partir de la
partición de `query.sql.go`/`models.go` (11 de 18 declaraciones eran [P]; las
otras 7 ya habían migrado a `identityrepo` en F2) más los cuatro archivos
puros (`content_queries.go`, `badge_queries.go`, `device_queries.go`,
`security_incident_queries.go`) y `db.go`, copiados sin editar una línea.
`internal/sync/integration_test.go` se difirió a F4 por depender de `auth` y
`testdb`, que todavía no existen. Verificado con `go build ./...`,
`go vet ./...` y `go test ./...` en verde (Go 1.22, contenedor
`golang:1.22-bookworm`), `diff` contra `../usbi` (cero diferencias salvo la
línea `module`) en los 11 paquetes verbatim, y los criterios 1 y 2 de
[`02_Backend.md` §8](02_Backend.md). No se tocó ningún paquete de `../usbi` ni
se levantó ninguna base de datos.

**F4 reescribió los cinco paquetes que solo tienen sentido hablando con dos
bases** (`auth`, `maintenance`, `transport`, `main.go`, `cmd/create_admin`),
más `internal/testdb` (dos esquemas desechables en vez de uno) y
`internal/domain` (quitó `FullName`, añadió `DisplayAlias`, pendiente desde
F3). `auth.Service`/`maintenance.Service` reciben ahora `(ident, main, cfg)`
en vez de `(q, cfg)`; `Login`/`Refresh` sincronizan la réplica `accounts` [P]
(alias sorteado una sola vez, nunca sobrescrito) y son la única escritura
cruzada a las dos bases fuera de la cancelación. La saga de cancelación
(`internal/privacy.PurgeMain`/`PseudonymizeIdentity`) queda persistida como
checkpoint reanudable en `arco_requests.status` (`pending → purging_main →
identity_pseudonymized → resolved/rejected`), reclamado bajo `SELECT ... FOR
UPDATE` dentro de una transacción corta para que dos llamadas concurrentes
sobre el mismo trámite no corran la purga dos veces en paralelo; el mismo
mecanismo lo comparten `auth.Service.ResolveArcoRequest` (reanuda si lo
vuelven a llamar) y el nuevo job de reconciliación de `maintenance` (retoma
trámites que nadie volvió a tocar, vía `identityrepo.ListStuckArcoRequests`).
Se detectó y corrigió un bug heredado de F3 (`device_label` en vez de
`device_kind`, columna ya renombrada en el esquema de F1) y se descubrió que
`cmd/create_admin` necesita escribir también en la base principal
(`admin_audit_log.actor_user_id` tiene FK a `accounts(id)`, así que un admin
recién creado no podría auditar ninguna acción de contenido sin esa réplica).
Verificado con `go build ./...`, `go vet ./...`, `go test ./...` y
`gofmt -l .` en verde (Go 1.22), dos pruebas de integración que matan el
proceso entre cada paso de la saga (una por cada estado intermedio) corridas
contra un esquema desechable en `usbi-database`, y un arranque real del
binario compilado contra dos bases Postgres desechables (registro, login,
alias generado, y `/health/ready` devolviendo 503 con el nombre de la base
caída al derribar cada una por separado). Los 7 criterios de
[`02_Backend.md` §8](02_Backend.md) se verificaron uno a uno; el criterio 4
(diff-cero) quedó con dos excepciones documentadas ahí mismo. Ninguna base
del proyecto (`usbi_db`, `maker_db`) se modificó — todas las verificaciones
usaron bases/esquemas desechables creados y eliminados en la misma sesión.

## 4. Decisiones que hay que cerrar antes de escribir el SQL definitivo

Estas cuatro no se asumen (regla 4 de `SKILLS.md`). Cada una lleva
recomendación, pero la decisión es del usuario.

1. **¿Dónde vive la BD de identidad?** El contenedor `usbi-database` ya aloja
   dos bases (`usbi_db` y `maker_db`), así que "dos bases en la misma
   instancia" es un patrón ya en uso aquí.
   *Recomendación:* **instancia PostgreSQL separada** para identidad. Dos bases
   en la misma instancia comparten proceso, `pg_hba.conf`, superusuario y
   volumen de respaldo — el aislamiento sería nominal, no real, y es
   exactamente lo que `SKILLS.md` §2 marca como hallazgo de severidad elevada.
   Los scripts SQL de F1 se escriben para funcionar en cualquiera de los dos
   escenarios (cero dependencias cruzadas), así que esta decisión **no bloquea
   F1**, solo el despliegue.

2. ✅ **CERRADA — se conserva un alias generado automáticamente.** El saludo
   del dashboard sobrevive como *"Bienvenido, Jaguar Azul 42"*.
   **Implementado más estricto de lo planteado:** en vez de una columna
   `display_alias VARCHAR`, el alias se guarda como **tres enteros** contra dos
   vocabularios curados (`alias_adjectives`, `alias_nouns`) más un número. Una
   columna de texto cumpliría hoy, pero cualquier cambio futuro de código
   podría escribir en ella un nombre real y nada en la base lo impediría — ni
   siquiera un `CHECK` de formato, porque *"Juan Pérez 12"* encaja en cualquier
   patrón razonable de "Adjetivo Sustantivo Número". Con tres enteros la base
   **físicamente no puede** almacenar un nombre, y el rol de aplicación no tiene
   permiso de escritura sobre los vocabularios. Ver
   [`01_Base_de_datos.md` §3.1](01_Base_de_datos.md).

3. ✅ **CERRADA — se elimina `phone`.** No existe en `identities` ni en ninguna
   otra tabla. Se almacenaba cifrada con índice ciego pero el login nunca la
   consultaba y `RegisterPage.tsx` nunca la enviaba.

4. **¿Herramienta de migraciones?** `../usbi` no usa ninguna: `0001` tiene
   anotaciones `-- +goose`, `0002`–`0012` no las tienen y no existe ningún
   `.down.sql`; `DEPLOYMENT.md` dice que se aplican "en orden de nombre de
   archivo" (o sea, `psql` a mano).
   *Recomendación:* adoptar **golang-migrate** con pares `.up.sql`/`.down.sql`,
   que es exactamente lo que `../maker` ya hace bien.

## 5. Qué NO cambia (para acotar expectativas)

- Las reglas de negocio de XP, antitrampas, idempotencia de `/sync` y HMAC
  offline: idénticas. El motor de progreso nunca supo el nombre de nadie.
- La identidad visual UV (`plan/Convenciones_de_color_UV.md`).
- El stack: Go 1.22 + chi + lib/pq + PostgreSQL 15; React 18 + Vite + Tauri 2.
- El límite de RAM de los contenedores (~256 MB para Postgres): **compilar
  fuera del contenedor**, regla heredada de `../usbi/CLAUDE.md`.

## 6. Referencias cruzadas rápidas

- Detalle tabla por tabla → [`01_Base_de_datos.md` §2 y §3](01_Base_de_datos.md)
- Firmas Go que cambian y contra qué base hablan → [`02_Backend.md` §3](02_Backend.md)
- Qué se copia y qué se ignora de cada repo → [`02_Backend.md` §2](02_Backend.md)
- Frontend → [`03_Frontend.md`](03_Frontend.md)
