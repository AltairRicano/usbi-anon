# 01 — Base de datos: qué se conserva, qué se va, qué se crea

Documento de detalle de [`00_Plan_maestro.md`](00_Plan_maestro.md). Cubre la
fase **F1**, ya **cerrada**: los scripts están escritos en `backend/` y
validados contra PostgreSQL 15. Ninguna base del proyecto se levantó — ver §8.

## 0. Punto de partida real (verificado, no supuesto)

Estado comprobado en el contenedor `usbi-database` (PostgreSQL 15, corriendo):

- Bases existentes: `usbi_db`, `maker_db`, `postgres`.
- `usbi_db` tiene **17 tablas** y coincide con la cadena de migraciones
  `0001`–`0012` de `../usbi/backend/migrations/`. Hay 6 filas en `users`.
- `maker_db` es el esquema "Privacidad Zero" de `../maker` (sesiones efímeras,
  sin PII), útil aquí solo como **patrón de referencia**, no como origen de
  datos.

Origen de cada pieza de este plan: `../usbi/backend/migrations/0001_initial.up.sql`
… `0012_drop_dead_routines.up.sql` y `../maker/maker/backend/migrations/0001_esquema_inicial.up.sql`.

---

## 1. Los dos scripts a producir

```text
backend/
├── migrations/
│   ├── identity/
│   │   ├── 0001_esquema_identidad.up.sql      # 283 líneas
│   │   └── 0001_esquema_identidad.down.sql
│   └── main/
│       ├── 0001_esquema_principal.up.sql      # ~460 líneas
│       └── 0001_esquema_principal.down.sql
└── sql/
    ├── 00_roles_identidad.sql                 # fuera de la cadena de migración
    └── 00_roles_principal.sql                 # (exigen superusuario, ver §5)
```

Dos cadenas de migración **independientes**, cada una con su propia tabla de
control de versión. Ninguna sentencia de un script puede referirse a un objeto
del otro: no hay FK, no hay JOIN, no hay `dblink`, no hay FDW. Esa es la
propiedad que hace que la decisión §4.1 del plan maestro (misma instancia vs.
instancia separada) sea puramente de despliegue y no bloquee F1.

**Consolidación:** ambos scripts son un **baseline único (`0001`)**, no la
copia de las 12 migraciones de `../usbi`. Ese es el mismo movimiento que
`../maker` ya hizo ("Sustituye a la cadena histórica de migraciones del sistema
anterior"). Lo que se gana al consolidar:

| Deuda de `../usbi` | Cómo desaparece en el baseline |
|---|---|
| `0002` crea `levels_template_type_check` como `NOT VALID`; `0011` lo valida 9 migraciones después | Se declara válido desde el `CREATE TABLE` |
| `0007` crea `level_attempts`/`daily_streak` sin particionar y luego las reconstruye particionadas con `DROP TABLE ... CASCADE` | Se crean **ya particionadas** de origen |
| `0008` parcha la ausencia de partición `DEFAULT` de `0007` | La partición `DEFAULT` viene desde el inicio |
| `0010` añade 5 índices de FK que faltaban | Los índices van junto a cada FK |
| `0006` define 5 funciones PL/pgSQL que `0012` elimina por muertas | **No se recrean.** Go es la única fuente de verdad de XP/ARCO |
| `0001` tiene anotaciones `-- +goose`, `0002`–`0012` no, y no existe ningún `.down.sql` | Convención única golang-migrate con `.down.sql` real |

**Lo único de `0006` que SÍ se conserva:** la función
`usbi_enforce_append_only_ledgers()` y sus dos triggers sobre
`experience_history` y `admin_audit_log`. `0012` las dejó vivas
deliberadamente y con razón: son invariantes que deben vivir en la base
independientemente de qué código escriba en esas tablas.

---

## 2. BD de identidad (`usbi_ident_db`) — qué contiene

Extensión requerida: **`pgcrypto`** (aquí sí, para `pgp_sym_encrypt` del correo
y del contacto del tutor).

### 2.1 `identities` — reemplaza a `users`

| Columna de `users` | Destino | Motivo |
|---|---|---|
| `id UUID PK` | **se conserva** | Es *el* UUID. Sujeto del JWT y única cosa que viaja a la BD principal |
| `email BYTEA` (cifrado) | **se conserva** | Credencial de login |
| `email_lookup_hash BYTEA` | **se conserva** | Índice ciego único parcial (`WHERE deleted_at IS NULL`) |
| `password_hash VARCHAR` | **se conserva** | Argon2id |
| `token_version INTEGER` | **se conserva** | Revocación inmediata; se valida en cada request |
| `is_adult BOOLEAN` | **se conserva** (y se replica) | Decide el flujo de consentimiento de tutor |
| `role VARCHAR` + CHECK | **se conserva** (y se replica) | Fuente de verdad de autorización |
| `status VARCHAR` + CHECK | **se conserva** (y se replica) | `active`/`suspended`/`pending_tutor_consent`/`deleted` |
| `privacy_notice_version`, `privacy_notice_accepted_at`, `privacy_acceptance_hash` | **se conserva** | Evidencia de no repudio del consentimiento |
| `crypto_key_version SMALLINT` | **se conserva** | Rotación de llave |
| `age_up_attempts SMALLINT` | **se conserva** | Contador Ley 251 (máx. 3) |
| `created_at`, `updated_at`, `last_login_at`, `deleted_at`, `deletion_reason` | **se conserva** | Ciclo de vida y retención |
| **`full_name VARCHAR(120)`** | **SE VA** | PII pura. No participa en autenticación. Ver decisión §4.2 del plan maestro |
| **`phone BYTEA` + `phone_lookup_hash BYTEA`** | **SE VA** (sujeto a decisión §4.3) | Almacenado y jamás consultado: `Login` solo usa correo, y `RegisterPage.tsx` ni siquiera lo envía |

### 2.2 Tablas que **se mudan completas** a la BD de identidad

| Tabla | Por qué aquí y no en la principal |
|---|---|
| `refresh_tokens` | Material de sesión atado a la credencial. FK a `identities(id)` |
| `tutor_consents` | `tutor_name` y `tutor_email` cifrados = PII de un tercero |
| `tutor_consent_tokens` | Ídem, más el `token_hash` del doble opt-in |
| `arco_requests` | El registro de la solicitud pertenece al titular del dato. `user_id` y `handled_by` con FK a `identities(id)` |

### 2.3 Tabla **nueva**: `identity_audit_log`

**Hallazgo del análisis, no estaba en `../usbi`.** Hoy `admin_audit_log` guarda
`before_state`/`after_state` en JSONB para *cualquier* acción administrativa,
incluidas las de identidad (resolver ARCO, suspender cuenta). Si esa tabla se
queda entera en la BD principal, un `before_state` de una edición de identidad
**filtra PII a la base que juramos que no la tiene**.

Solución: partir la bitácora en dos, con el mismo esquema y las mismas
garantías append-only:

- `identity_audit_log` (BD identidad) — auth, ARCO, consentimiento de tutor,
  suspensión, cancelación.
- `admin_audit_log` (BD principal) — contenido, secciones, niveles, incidentes.

Ambas conservan el trigger `enforce_append_only` y la semántica de
seudonimización por `SET NULL` del actor.

---

## 3. BD principal (`usbi_anon_db`) — qué contiene

**No requiere `pgcrypto`.** Al no quedar ni un solo campo cifrado en esta base,
la dependencia desaparece por completo — resultado limpio y verificable del
rediseño.

### 3.1 Tabla **nueva**: `accounts` (el ancla)

```text
accounts(
  id                 UUID PRIMARY KEY,   -- = identities.id, SIN FK (otra base)
  role               VARCHAR NOT NULL CHECK (...),
  status             VARCHAR NOT NULL CHECK (...),
  is_adult           BOOLEAN NOT NULL,
  alias_adjective_id SMALLINT NOT NULL REFERENCES alias_adjectives(id),
  alias_noun_id      SMALLINT NOT NULL REFERENCES alias_nouns(id),
  alias_number       SMALLINT NOT NULL CHECK (alias_number BETWEEN 0 AND 999),
  created_at, updated_at, deleted_at
)
```

Réplica no autoritativa de `identities`, **sin una sola columna de texto
libre**. Se refresca en `login`/`refresh` (ver
[`02_Backend.md` §4](02_Backend.md)).

**El alias visible se guarda como tres enteros, no como texto.** La decisión
§4.2 del plan maestro conservó el saludo del dashboard con un alias generado
por el sistema. Una columna `display_alias VARCHAR` habría bastado hoy, pero
cualquier cambio futuro de código podría escribir en ella un nombre real y nada
en la base lo impediría — ni un `CHECK` de formato, porque *"Juan Pérez 12"*
encaja en cualquier patrón razonable de "Adjetivo Sustantivo Número". Con tres
enteros contra vocabularios curados (`alias_adjectives`, `alias_nouns`, 24×24
palabras → 576 000 combinaciones), la base **físicamente no puede** almacenar un
nombre. Es una garantía estructural verificable de un vistazo en una auditoría,
no una convención. Los adjetivos son invariantes en género (*azul, veloz,
feliz…*) para concordar con cualquier sustantivo.

La vista `account_aliases` compone la cadena legible (`"Jaguar Azul 42"`) sin
que ninguna tabla la almacene. **El alias no es un identificador**: no es único
y jamás debe usarse como clave de búsqueda. La llave siempre es `accounts.id`.

> **Regla dura para cualquier cambio futuro:** si alguien propone añadir a
> `accounts` una columna que pueda contener texto escrito por una persona, es
> incompatible con el diseño (`SKILLS.md`, regla transversal 2).

### 3.2 Tablas que **se conservan sin cambios de forma**

Solo cambia el destino de su FK: `REFERENCES users(id)` → `REFERENCES accounts(id)`.
El nombre de columna `user_id` **se conserva** deliberadamente: renombrarlo a
`account_id` en 12 tablas es un diff enorme con cero ganancia de seguridad. Se
documenta con `COMMENT ON COLUMN` que contiene un UUID anónimo.

| Tabla | FK que cambia de destino | Nota |
|---|---|---|
| `devices` | `user_id` | ⚠️ ver §3.4 — `device_label` es texto libre |
| `sync_events` | `user_id` + FK compuesta a `devices(id,user_id)` | El `payload` JSONB ya está documentado como "No PII" |
| `sections` | `created_by_admin_id` | `ON DELETE SET NULL` |
| `levels` | `created_by_admin_id`, `deleted_by` | |
| `player_progress` | `user_id` | PK compuesta `(user_id, level_id)` |
| `level_attempts` | `user_id` | Particionada por rango sobre `attempt_date` |
| `daily_streak` | `user_id` | Particionada por rango sobre `activity_date` |
| `user_badges` | `user_id` | |
| `experience_history` | `user_id` (`SET NULL`) | Append-only + trigger |
| `admin_audit_log` | `actor_user_id` (`SET NULL`) | Append-only + trigger; ahora solo acciones de contenido (§2.3) |
| `badges` | — (sin FK a usuario) | Se conserva íntegra, incluido el seed de `0004` |
| `security_incidents` | — (sin FK a usuario) | ⚠️ ver §3.4 |

### 3.3 Particionado: se conserva, corregido

`level_attempts` y `daily_streak` siguen particionadas por rango
(`attempt_date` / `activity_date`), con particiones anuales explícitas **más
una partición `DEFAULT`** desde el minuto cero, y el planificador
`internal/dbmaint` sigue creando particiones futuras (año actual + 2). Se
conserva porque la escala documentada (10k usuarios / 5k niveles) lo justifica
y porque el código Go que las mantiene se copia verbatim.

### 3.4 Dos fugas de PII latentes que hay que cerrar en el SQL

Ninguna de las dos existe como bug en `../usbi` (allí la base ya tenía PII, así
que no importaba). Aquí sí importan, porque la promesa de la BD principal es
absoluta.

1. ✅ **`devices.device_label VARCHAR NOT NULL`** era texto libre escrito por
   la persona usuaria. En la vida real eso se llena con *"iPad de Sofía"*.
   **Resuelto:** sustituida por `device_kind` con vocabulario cerrado
   (`movil`, `tablet`, `laptop`, `escritorio`, `otro`). Basta para distinguir
   dispositivos en pantalla y el `CHECK` rechaza cualquier otra cosa —
   verificado: insertar `'iPad de Sofia'` falla.
2. ⚠️ **`security_incidents.affected_scope` y `.description`** son `TEXT` que
   redacta un operador durante un incidente — el momento exacto en que alguien
   escribe *"se filtró la cuenta de Juan Pérez"*. **Acción:** no se puede
   resolver con `CHECK`; se resuelve con `COMMENT ON COLUMN` explícito
   prohibiendo nombrar titulares y con validación en `internal/incidents`.
   **Aplicado:** los tres campos de texto llevan `COMMENT ON COLUMN`
   prohibiendo nombrar titulares. Queda como control documental, no técnico —
   es el único texto libre que sobrevive en esta base.

---

## 4. Lo que cambia en las operaciones que cruzan las dos bases

El SQL debe escribirse sabiendo que estas cuatro rutinas **ya no son una
transacción única**. El detalle de orquestación está en
[`02_Backend.md` §5](02_Backend.md); aquí queda lo que el esquema debe ofrecer:

| Operación | Lado identidad | Lado principal |
|---|---|---|
| Registro | `INSERT identities` | *(nada — `accounts` se crea al primer login)* |
| Login | lectura por índice ciego + `INSERT accounts ... ON CONFLICT DO UPDATE` | upsert de la réplica |
| Cancelación ARCO | seudonimizar `identities`, `tutor_consents`; revocar `refresh_tokens`; `token_version + 1` | purgar `player_progress`/`level_attempts`/`daily_streak`/`user_badges`; `SET NULL` en `experience_history` y `admin_audit_log`; `devices.wipe_local_data = TRUE` |
| Retención automática | listar `pending_tutor_consent` vencidos y `suspended` a cancelar | ejecutar la purga anterior por UUID |

`arco_requests.status` gana estados intermedios para que la saga sea
reanudable: `pending → purging_main → identity_pseudonymized → resolved`
(hoy el `CHECK` de `0011` solo admite `pending`/`resolved`/`rejected`).

---

## 5. Aislamiento real: roles de base de datos (no negociable)

`SKILLS.md` §2 marca como hallazgo de severidad elevada cualquier vía no
autorizada para re-vincular un UUID con su identidad. Dos bases con el **mismo
usuario de conexión** no aíslan nada. Los scripts deben crear:

| Rol | Sobre | Permisos |
|---|---|---|
| `usbi_ident_app` | `usbi_ident_db` | `SELECT/INSERT/UPDATE` en las tablas de identidad. Sin `DELETE` en bitácoras |
| `usbi_main_app` | `usbi_anon_db` | `SELECT/INSERT/UPDATE/DELETE` en progreso; sin acceso alguno a la otra base |
| `usbi_*_migrate` | cada una | Solo para aplicar migraciones; no lo usa el backend en runtime |

El backend abre **dos pools con dos credenciales distintas**. Que el proceso Go
tenga ambas es inevitable; que la base las trate como la misma, no.

---

## 6. Mejoras propuestas y con qué skill se atacan

| # | Mejora | Skill / agente | Cuándo |
|---|---|---|---|
| 1 | Escribir y revisar los dos baselines (particionado, índices, `CHECK`, `COMMENT`) | **`postgres-patterns`** | F1, al redactar el SQL |
| 2 | Revisar plan de ejecución de las consultas de progreso tras perder los JOIN a `users` | agente **`database-optimizer`** | F1→F2, tras el baseline |
| 3 | Verificar que el aislamiento entre bases es real (roles, red, respaldos) y que ninguna ruta re-vincula UUID↔identidad | **`cyber-neo`** + agente **`security-engineer`** | Cuando exista código (F4) |
| 4 | Auditar la saga ARCO de dos fases contra Ley 251 / LGPDPPSO / GDPR-K | agente **`legal-compliance-checker`** | F6, junto con la reescritura legal |
| 5 | Segunda opinión sobre el modelo de dos bases sin gastar cuota propia | **`agy-delegate`** (modo `strict`, solo lectura) | Cualquier momento |
| 6 | Puerta de revisión del diff SQL antes de dar F1 por cerrada | **`/security-review`** + **`/code-review`** | Cierre de F1 |

> Al invocar cualquiera de estas skills o agentes: instruirlos explícitamente a
> escribir su reporte en su propia sección de `estado_proyecto.md`
> (regla 5 del `CLAUDE.md` global).

---

## 7. Criterios de aceptación de F1 — resultado

| # | Criterio | Resultado |
|---|---|---|
| 1 | Sin columnas `full_name`/`email`/`phone`/`tutor` en `main/` | ✅ ninguna |
| 2 | `main/` no nombra `identities`; `identity/` no nombra `accounts` ni progreso | ✅ |
| 3 | Sin `dblink` ni `postgres_fdw` | ✅ |
| 4 | `main/0001` no crea `pgcrypto` | ✅ verificado: `pg_extension` sin `pgcrypto` |
| 5 | Las 17 FK sobreviven con el mismo modo (`CASCADE`/`RESTRICT`/`SET NULL`) | ✅ 11 hacia `accounts` + 6 hacia `identities` |
| 6 | `level_attempts` y `daily_streak` nacen particionadas y con partición `DEFAULT` | ✅ 2026-2028 + `DEFAULT`; un insert de 2026-08-24 cae en `level_attempts_2026` |
| 7 | Triggers append-only, uno por base | ✅ ver §8 |
| 8 | Cada `.up.sql` tiene su `.down.sql` reversible | ✅ ambas bases quedan en 0 tablas y el `up` reaplica limpio |
| 9 | No se aplicaron migraciones a `usbi-database` | ✅ ver §8 |

## 8. Cómo se validó (y por qué no se levantó ninguna base)

Un script SQL de 460 líneas con particiones, claves foráneas compuestas y
triggers PL/pgSQL que nunca se ha ejecutado no es un entregable: es una
hipótesis. Para verificarlo sin levantar las bases del proyecto se crearon dos
bases **desechables** (`sqlcheck_ident`, `sqlcheck_main`), se aplicaron los
cuatro scripts, se corrieron las pruebas de abajo y se **eliminaron ambas** en
la misma sesión.

Estado del contenedor `usbi-database` antes y después: idéntico —
`usbi_db` (17 tablas, 6 filas en `users`), `maker_db`, `postgres`.

Pruebas ejecutadas y su resultado:

| Prueba | Esperado | Resultado |
|---|---|---|
| Alta de cuenta + vista de alias | `"Jaguar Azul 42"` | ✅ |
| Columnas de texto libre en `accounts` | ninguna | ✅ 0 filas |
| `DELETE FROM experience_history` | falla | ✅ `experience_history es append-only` |
| `UPDATE ... SET xp_gained = 999` | falla | ✅ error del trigger |
| `UPDATE ... SET user_id = NULL` (seudonimización ARCO) | pasa, conservando el resto de la fila | ✅ |
| Insert en `level_attempts` con fecha de 2026 | cae en la partición anual | ✅ `level_attempts_2026` |
| `device_kind = 'iPad de Sofia'` | falla | ✅ viola `devices_device_kind_check` |
| `down` con datos dentro, y `up` de nuevo | ambas limpias | ✅ |

Lo que **no** se ha hecho, deliberadamente: crear `usbi_ident_db` ni
`usbi_anon_db`, aplicar los scripts de roles (exigen superusuario y decisiones
de despliegue), y tocar `usbi_db` o `maker_db`.
