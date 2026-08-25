# Rediseño de identidad: de email+dos bases a nickname/password por cuestionario de gustos

> **Estado: diseño cerrado y aprobado por el usuario (2026-08-25). F5 cerrada;
> F6–F11 sin empezar.** El esquema unificado ya existe como
> `backend/migrations/0001_esquema_unificado.{up,down}.sql` +
> `backend/sql/00_roles_unificado.sql`, verificado contra Postgres real en una
> base desechable (ver §6 y la bitácora de `estado_proyecto.md`). Este documento
> reemplaza, para todo lo relativo a identidad y
> autenticación, las secciones equivalentes de `01_Base_de_datos.md` y
> `02_Backend.md` que describían el modelo de dos bases con email. Esos dos
> documentos siguen siendo la referencia válida para todo lo que **no** cambia
> (progreso, contenido, niveles, sync, dispositivos — ver §2 de este documento).
> `00_Plan_maestro.md` §3 apunta aquí para las fases F5 en adelante.

## Contexto

El backend hasta F4 (cerrada, compilada y probada) implementaba un modelo de
identidad con **dos bases de datos separadas**: una de credenciales (email
cifrado + password) y una principal de progreso, unidas por UUID, con un flujo
completo de consentimiento de tutor por correo electrónico para menores de
edad.

El usuario reconsideró la premisa que originó ese diseño: ya no está seguro de
que la universidad vaya a alojar una segunda base de datos separada, y —
independientemente de eso— prefiere no pagar la complejidad de dos bases para
una aplicación de este tamaño. La alternativa que propuso es más radical que
solo "una base": **eliminar el correo electrónico del sistema por completo**.
En su lugar, el registro se hace respondiendo un cuestionario de gustos no
sensibles (color favorito, animal preferido, materia favorita, número
favorito, cantidad de mascotas), y el sistema deriva de esas respuestas un
**nickname** (credencial de login, elegido por el usuario entre 4 opciones
generadas) y un **password** de 12 caracteres (generado, mostrado una sola
vez). El cuestionario es administrable: banco de preguntas con mínimo 4
activas, máximo 10 mostradas por registro, resto en reserva rotando.

Esto no es un ajuste: invalida el diseño central de F1–F4 (separación de
bases, todo el flujo de tutor por correo, ~500 líneas de `internal/auth`, dos
pools de conexión, la saga ARCO en dos fases). Se decidió con el usuario,
explícitamente, **descartar y reescribir** en vez de adaptar. El resultado
deja el sistema con una sola base de datos, sin ningún dato personal directo
(ni siquiera un correo cifrado), y con el alias visible del jugador
(`display_alias`, ya resuelto en F1 con `RandomAlias()`/vista
`account_aliases`) coexistiendo — pero sin fusionarse — con el nuevo
`nickname` de login.

**Decisiones ya cerradas con el usuario (no reabrir sin volver a preguntar):**

1. Una sola base de datos Postgres (se elimina `usbi_ident_db`, todo vive en
   la base principal).
2. Se descarta y reescribe desde cero `internal/auth`, `internal/identityrepo`,
   `main.go`, `transport/router.go`, `cmd/create_admin`, `internal/testdb`.
3. Se elimina por completo el flujo de tutor por correo (`tutor_consents`,
   `tutor_consent_tokens`, `internal/mailer`). Un menor autoreportado
   (`is_adult=false`) juega de inmediato, sin gate de aprobación.
4. Recuperación de cuenta: un admin resetea el password manualmente,
   comparando a ojo las respuestas que el usuario reingresa contra las
   guardadas — el sistema debe persistir las respuestas de texto libre
   asociadas a la cuenta.
5. Las respuestas del cuestionario son **texto libre validado** (no
   dropdown), con validación real anti-inyección/anti-JSON en frontend Y
   backend.
6. Bootstrap del primer admin: se inserta directamente en la base vía script
   SQL de seed (con password ya hasheado por un binario standalone). Desde la
   app, un admin puede crear otros admins con nickname+password explícitos.
   Un admin puede borrar su propia cuenta pero **nunca** la de otro admin.
7. Cancelación de cuenta: autoservicio inmediato (`DELETE /auth/me`), sin
   aprobación de nadie.
8. Rotación del banco de preguntas cuando hay más activas que el máximo
   configurado: muestreo aleatorio puro por registro.
9. Estado intermedio del registro (entre "elegir preguntas" y "confirmar
   nickname"): token HMAC firmado con TTL de 10 min, sin tabla intermedia.
10. Password: relleno aleatorio con `crypto/rand` (nickname sigue usando
    `math/rand` sembrado por el instante exacto, tal como pidió el usuario,
    por ser solo un identificador público, no el secreto).
11. Nickname y `display_alias` **coexisten, no se fusionan**: nickname =
    credencial de login derivada de respuestas reales; display_alias =
    saludo 100% aleatorio (`crypto/rand`), ya construido, no deriva de nada
    del usuario.

---

## 1. Esquema SQL unificado

Reemplaza `backend/migrations/{identity,main}/` por un único
`backend/migrations/0001_esquema_unificado.{up,down}.sql`, y
`backend/sql/00_roles_{identidad,principal}.sql` por
`backend/sql/00_roles_unificado.sql`. Es un reemplazo de baseline, no una
migración incremental: F1–F4 nunca aplicaron estos scripts contra una base
persistente (siempre esquemas desechables), así que no hay dato real que
migrar.

Se elimina `CREATE EXTENSION pgcrypto` (ya no hay ningún campo cifrado en el
sistema).

**`accounts`** (fusión de `identities` + `accounts`): `id UUID PK`,
`nickname VARCHAR(32) UNIQUE NOT NULL` (minúsculas, `[a-z0-9]`, 6–20
caracteres, **sin** blind index HMAC — no es PII cifrada, se busca con
`WHERE nickname = $1` directo), `password_hash VARCHAR` (Argon2id, reutiliza
`crypto.HashPassword` tal cual), `token_version`, `role` (player/admin/
operator/director), `status` (active/suspended/deleted — **sin**
`pending_tutor_consent`), `is_adult BOOLEAN` (autorreporte puro),
`age_up_attempts`, `alias_adjective_id`/`alias_noun_id`/`alias_number` (se
conservan, ver `RandomAlias()`), `privacy_notice_version`/
`privacy_notice_accepted_at`/`privacy_acceptance_hash` (el sello HMAC ya no
incluye email, se calcula sobre `id + version + accepted_at`), timestamps,
`deleted_at`, `deletion_reason`.

**`refresh_tokens`**: igual que hoy, FK renombrada a `account_id` (solo en
esta tabla; las 11 FK de progreso/contenido conservan `user_id` literal para
no invalidar la capa Go verbatim, tal como ya documenta el comentario
existente del esquema principal).

**Se elimina sin reemplazo**: `tutor_consents`, `tutor_consent_tokens` y sus
índices.

**`registration_questions`** (nueva): `id UUID PK`,
`question_text VARCHAR(280)`, `is_active BOOLEAN`, `display_order SMALLINT`,
timestamps. La regla "mínimo 4 en el banco" se valida en Go dentro de una
transacción antes de confirmar un DELETE (no como CHECK/trigger — consistente
con el resto del proyecto, que mantiene la lógica de negocio en Go).

**`registration_settings`** (nueva, fila única): `id SMALLINT PK DEFAULT 1
CHECK (id=1)`, `max_questions_shown SMALLINT CHECK (BETWEEN 4 AND 10)`.

**`account_quiz_answers`** (nueva): `id UUID PK`,
`account_id FK → accounts ON DELETE CASCADE`,
`question_id FK → registration_questions ON DELETE SET NULL`,
`question_text_snapshot VARCHAR(280)` (congela el texto visto por el usuario,
sobrevive a ediciones futuras de la pregunta), `answer_text VARCHAR(200)`,
`created_at`. Sin cifrar (respuestas no sensibles por diseño), pero el
endpoint que las expone queda protegido por rol admin y auditado.

**`audit_log`** (fusión de `identity_audit_log` + `admin_audit_log`): ya no
hay razón para partirla en dos sin email que filtrar. Mismo trigger
append-only reutilizado, columna `actor_account_id`.

**`arco_requests`**: se simplifica — ya no hace falta la máquina de estados
de 3 pasos (no hay problema de 2PC entre bases). `status` se reduce a
`pending/resolved/rejected`. Sigue sirviendo para `acceso`/`rectificacion`/
`oposicion`; `cancelacion` pasa a resolverse en una transacción única e
inmediata desde `DELETE /auth/me`.

**Sin cambios**: `devices`, `sync_events`, `sections`, `levels`,
`player_progress`, `level_attempts` (particionada), `daily_streak`,
`badges`, `user_badges`, `experience_history`, `security_incidents`,
`alias_adjectives`, `alias_nouns`, vista `account_aliases`.

### 1.1 Detalles resueltos al escribir F5 (no estaban en el diseño)

Cinco puntos que el diseño no fijaba y que había que decidir para que el SQL
existiera. Ninguno reabre una decisión cerrada; se documentan aquí para que F7
y F9 no tengan que deducirlos del archivo.

1. **`accounts.nickname` lleva `CHECK (nickname ~ '^[a-z0-9]{6,20}$')`.** El
   diseño fijaba el formato pero solo como regla de aplicación. Escrito como
   CHECK, la base rechaza físicamente `'Juan Perez'`, `'juan@correo.com'` y
   cualquier cadena con espacios, acentos o mayúsculas — el mismo tipo de
   garantía estructural que ya daban los tres enteros del alias. Verificado con
   los tres INSERT rechazados.
2. **La cancelación sobrescribe el nickname con relleno aleatorio `[a-z0-9]` de
   20 caracteres.** `UNIQUE` sobre `nickname` es total (no parcial por
   `deleted_at IS NULL`), tal como pedía §1. Para que una cuenta cancelada no
   bloquee para siempre su nickname —y para no dejar rastro de las respuestas
   que lo originaron—, `privacy.CancelAccount` (F7) debe sobrescribirlo. El
   relleno aleatorio cumple el CHECK, así que no hace falta ninguna excepción
   al formato.
3. **Se conserva `accounts.crypto_key_version`.** El diseño no la listaba
   porque ya no hay cifrado, pero `privacy_acceptance_hash` sigue siendo un
   HMAC: sin versión de clave, rotar `HMAC_SECRET` invalidaría de golpe toda la
   evidencia de no repudio del aviso de privacidad.
4. **`arco_requests` conserva `user_id` / `handled_by`.** §1 solo autoriza
   renombrar a `account_id` en `refresh_tokens`, y `audit_log` nace con
   `actor_account_id` por ser tabla nueva. El resto se queda como está, igual
   que las 11 FK de progreso.
5. **El baseline siembra el banco de preguntas y la configuración.** Cinco
   preguntas activas (color, animal, materia, número, mascotas) y
   `max_questions_shown = 5`. No es dato de prueba: el sistema exige mínimo 4
   activas para poder registrar a nadie, así que una base recién migrada sin
   seed no permitiría ni el primer registro.

### 1.2 Rotación de niveles por temporadas (corrección posterior a F5)

Regla de negocio que el usuario tenía por implícita y no estaba en ningún
documento ni en el código. Se incorporó al esquema el mismo día que se cerró
F5, antes de escribir nada de F7/F9 contra él.

**La regla:** el almacenamiento del servidor es finito (~20 GB) y no se
ampliará pagando más. El contenido se **rota entre temporadas**: se retiran
niveles y secciones viejos para meter otros. Retirar un nivel debe liberar su
almacenamiento **sin quitarle a ningún jugador la experiencia que ganó
jugándolo**.

**Por qué el esquema no lo cumplía.** Tres cosas, todas heredadas de `../usbi`:

1. El "borrado" de niveles era **solo lógico** (`SET deleted_at = NOW()`, ver
   `repository/content_queries.go:273` y `:289`). No liberaba un byte: la base
   solo podía crecer.
2. Las tres FK a `levels(id)` eran `ON DELETE RESTRICT`, así que un borrado
   físico era **imposible** en cuanto un solo jugador tocara el nivel.
3. `experience_history.level_id` era `NOT NULL`, y el XP total del jugador sale
   de `SUM(experience_history.xp_gained)`
   (`repository/content_queries.go:403`). Forzar el borrado con CASCADE habría
   borrado justamente las filas que sostienen el XP.

**Decisiones cerradas con el usuario** (tres preguntas, las tres con la opción
recomendada): conservar XP **y** contadores; conservar el historial fila por
fila (no colapsarlo); y retiro en **dos pasos** (archivar reversible → purgar
irreversible).

**Cambios aplicados al esquema:**

| Referencia a `levels(id)` | Antes | Ahora | Por qué |
|---|---|---|---|
| `level_attempts.level_id` | RESTRICT | **CASCADE** | Es la tabla que más crece; aquí está el ahorro real de disco |
| `player_progress.level_id` | RESTRICT | **CASCADE** | Sin nivel no hay progreso "de" él; los contadores se preservan aparte |
| `experience_history.level_id` | `NOT NULL` RESTRICT | **NULLABLE, SET NULL** | La fila sobrevive con su `xp_gained` intacto: el XP no se pierde |

Más una tabla nueva, **`account_retired_progress`** (`account_id`,
`levels_completed`, `attempts_total`), acumulativa, que preserva los contadores
de "niveles completados" e "intentos totales" de los niveles ya purgados. No
guarda XP: eso lo sostiene `experience_history` y duplicarlo lo contaría dos
veces. `levels.section_id` **sigue siendo RESTRICT** a propósito: purgar una
sección exige purgar antes sus niveles, para que un clic en la sección no
dispare un borrado masivo silencioso.

El trigger `enforce_append_only_ledgers()` ganó una segunda rama permitida
(`level_id → NULL`) junto a la de seudonimización ARCO. Ninguna de las dos
puede tocar `xp_gained`: "rotar niveles no quita XP" queda garantizado por la
base, no por el código.

**Qué le toca a cada fase pendiente:**

- **F7** (`internal/repository`): `GetUserProgressTotals` debe sumar las dos
  fuentes — `player_progress` vivos **más** `account_retired_progress`. Hoy
  solo cuenta la primera. `ListUserProgressLevels` y cualquier consulta de
  historial deben pasar a `LEFT JOIN` contra `levels` y tolerar `level_id`
  nulo.
- **F9** (`internal/levels` + transport): implementar la purga física como
  acción de admin separada del archivado, en una sola transacción y en este
  orden — acumular en `account_retired_progress` → `DELETE FROM levels`. La
  consulta exacta del UPSERT está en el comentario de esa tabla en el `.up.sql`.
  Validar en Go que solo se purga lo ya archivado (`deleted_at IS NOT NULL`).
- **F10** (frontend): la confirmación de purga debe decir cuántos intentos se
  van a borrar y que el XP se conserva. El listado de progreso debe tolerar
  niveles ausentes.

**Nota de magnitud, para no sobreestimar el ahorro.** Con ~500 jugadores y ~40
niveles, estas tablas rondan decenas de MB, no GB. Si los 20 GB se llenan, el
consumidor dominante serán los **assets de los minijuegos** (imágenes, audio),
no la base: `levels.content` guarda solo estructura. Lo que esta corrección
desbloquea de forma crítica no es tanto el disco como la **capacidad misma de
rotar**, que antes era imposible.

## 2. Paquetes Go

**Se elimina**: `internal/identityrepo/` completo, `internal/mailer/`
(único consumidor era el correo de tutor), `cmd/create_admin/`, y de
`internal/config`: `IdentityDatabaseURL()`, `PGP_ENCRYPTION_KEY`, `SMTP_*`,
`BLIND_INDEX_SECRET`. `HMAC_SECRET` se conserva (lo sigue usando
`privacy_acceptance_hash` y las bitácoras).

**Reutilizable sin cambios**: `internal/crypto` completo (Argon2id, HMAC,
JWT — nada depende de email), `internal/levels`, `internal/sync`,
`internal/devices`, `internal/incidents`, `internal/dbmaint`,
`internal/audit`, `internal/httpjson`, `internal/httpproblem`,
`internal/httputil`, todas las tablas/queries de progreso-contenido
existentes, `RandomAlias()`/`account_aliases` (pero ahora se llama **una sola
vez, en el INSERT de registro**, ya no en cada login/refresh — no hay dos
filas que reconciliar).

**Se reescribe**: `internal/domain/models.go` (pierde
`pending_tutor_consent`, gana `Nickname` en `domain.User`),
`internal/repository` (absorbe de identityrepo lo reutilizable:
`FindAccountByNickname`, `CreateAccount`, refresh tokens, audit unificado),
`internal/privacy/privacy.go` (colapsa la saga de 3 fases a una función
`CancelAccount(ctx, q, params) error` de ~40 líneas en una sola `sql.Tx` — se
elimina toda la maquinaria de checkpoint/reconciliación, incluido el job de
`internal/maintenance` que la alimentaba), `internal/transport/router.go` y
`main.go` (un solo pool, un solo health check), `internal/testdb` (un solo
esquema desechable).

**Se crea nuevo — `internal/quiz`** (dos archivos, responsabilidades
separadas como pidió el usuario):

- `bank.go`: CRUD del banco de preguntas (`ListAllQuestions`,
  `CreateQuestion`, `UpdateQuestion`, `DeleteQuestion` con guard de mínimo 4,
  `SetMaxQuestionsShown` con guard 4–10, `SelectQuestionsForRegistration`
  con muestreo aleatorio puro vía `math/rand`), más su propio `Handler` HTTP
  para las rutas admin del banco.
- `credentials.go`: `GenerateNicknameCandidates(answers, existsFn) ([4]string,
  error)` y `GeneratePassword(answers) (string, error)` — sin HTTP, sin
  acceso a banco de preguntas, orquestadas por `internal/auth`.

**`internal/auth`** reescrito: se elimina `SubmitTutorConsent`/
`VerifyTutorConsent` (~500 líneas). Se conserva `AgeUp` simplificado (solo
`age_up_attempts` + `is_adult=true`, sin tutor) y `Arco`/`ListPendingArco`/
`ResolveArco` para acceso/rectificación/oposición. Nuevo flujo de registro en
3 pasos (ver §3) y `DELETE /auth/me` para cancelación autoservicio.

**Bootstrap de admin**: `cmd/hash_password/main.go` (nuevo, standalone, sin
DB) imprime el hash Argon2id de un password dado; `backend/sql/
01_seed_primer_admin.sql` (plantilla comentada) para el INSERT manual del
primer admin. Endpoint `POST /admin/accounts` (nickname+password explícitos)
para que un admin cree a otros. `DELETE /admin/accounts/{id}` responde `403`
si el objetivo tiene `role='admin'`, sin excepción.

## 3. Contrato de API (nuevo/cambiado)

```
POST /api/v1/auth/register/questions   → { questions: [{id, text}], max_questions_shown }
POST /api/v1/auth/register/answers     body: { answers: [{question_id, answer_text}], is_adult, privacy_notice_version }
                                        → valida anti-inyección, genera 4 candidatos
                                        → { registration_token (HMAC, TTL 10min), nickname_candidates: [4] }
POST /api/v1/auth/register/confirm     body: { registration_token, chosen_nickname }
                                        → recheck de colisión, genera password, INSERT accounts+respuestas+alias
                                        → 201 { account_id, nickname, password, display_alias }  (se muestran UNA vez)

POST /api/v1/auth/login                { nickname, password }
POST /api/v1/auth/refresh              { refresh_token }
POST /api/v1/auth/logout               (auth)
GET  /api/v1/auth/me                   (auth)
POST /api/v1/auth/age-up               (auth)
DELETE /api/v1/auth/me                 (auth) → cancelación autoservicio inmediata

POST /api/v1/arco                      (auth) → acceso/rectificacion/oposicion
GET  /api/v1/arco/pending              (auth, admin)
POST /api/v1/arco/{id}/resolve         (auth, admin)

GET    /api/v1/admin/registration-questions
POST   /api/v1/admin/registration-questions
PATCH  /api/v1/admin/registration-questions/{id}
DELETE /api/v1/admin/registration-questions/{id}     → 409 si quedarían < 4
PUT    /api/v1/admin/registration-settings           { max_questions_shown }

POST   /api/v1/admin/accounts                         { nickname, password, role }
DELETE /api/v1/admin/accounts/{account_id}             → 403 si target.role == 'admin'
GET    /api/v1/admin/accounts/{account_id}/quiz-answers
POST   /api/v1/admin/accounts/{account_id}/reset-password → { new_password } (una vez)
```

## 4. Algoritmo de generación (Go)

**Nickname** (`internal/quiz/credentials.go`): normaliza cada respuesta
(minúsculas, mapa manual de acentos —`golang.org/x/text` no está en
`go.mod`, no vale la pena la dependencia—, solo `[a-z0-9]`), recorta a 4
caracteres por fragmento. Semilla `time.Now().UnixNano()` en `math/rand`
(pseudoaleatoriedad del instante exacto, tal como pidió el usuario,
deliberadamente distinta del `crypto/rand` de `RandomAlias()`). Combina 2
fragmentos + 3 dígitos por candidato; verifica colisión contra la base vía la
función `exists` inyectada, regenerando solo el candidato que choque, hasta
tener 4 disponibles.

**Password** (12 caracteres exactos): toma un fragmento de una respuesta
(3–5 caracteres, capitalizado), rellena el resto con `crypto/rand` sobre un
charset sin caracteres ambiguos (`0/O/1/l/I` excluidos), e intercala
fragmento y relleno (no "fragmento + relleno" plano) para que el patrón no
sea trivialmente adivinable conociendo el algoritmo. Se hashea con
`crypto.HashPassword` antes de guardar; el texto plano se devuelve una sola
vez en la respuesta de `register/confirm`.

## 5. Frontend (greenfield — no existe `frontend/` todavía)

Sigue el stack y patrones ya confirmados en `../usbi/frontend` (React 18 +
Vite 6 + TS + Tauri v2 + Zustand con persist en `sessionStorage` + axios +
Zod), pero **greenfield**, no copiado:

```
frontend/src/
  features/
    auth/
      RegisterPage.tsx        # wizard de 3 pasos (useState step 1|2|3), sin router extra
      LoginPage.tsx            # nickname+password
      useAuthStore.ts          # Zustand persist, User.nickname
      schemas.ts                # Zod anti-inyección (ver abajo) — excepción deliberada:
                                 # aquí Zod SÍ valida input de formulario, a diferencia del resto
                                 # del proyecto hermano donde Zod solo valida respuestas del backend
    admin-quiz-bank/
      AdminQuizBankPage.tsx     # CRUD acordeón, mismo patrón que AdminContentPage.tsx (../usbi)
      MinQuestionsModal.tsx     # el ÚNICO modal de toda la UI — fondo difuminado, centrado,
                                 # excepción única y deliberada a la convención "sin modales"
    admin-accounts/
      AdminAccountsPage.tsx     # crear admin, borrar cuenta (deshabilitado si target es admin),
                                 # ver respuestas de cuestionario, resetear password
  shared/
    apiClient.ts                # interceptor de refresh, reutilizado literal del patrón ../usbi
    errorMessage.ts             # helper de RFC7807, reutilizado literal
```

Validación Zod (duplicada en Go, el frontend nunca es la única barrera):

```ts
z.string().trim().min(1).max(200)
  .refine(v => !/[ -]/.test(v))     // sin caracteres de control
  .refine(v => !/^[\[{]/.test(v))                    // sin JSON
  .refine(v => !/<[a-z][\s\S]*>/i.test(v))           // sin HTML
```

## 6. Fases de entrega y dependencias entre ellas

| # | Fase | Entregable | Depende de |
|---|---|---|---|
| F5 ✅ **hecha** | Esquema SQL unificado | `0001_esquema_unificado.{up,down}.sql`, `00_roles_unificado.sql`, retiro de los scripts antiguos | — (primera, bloquea todo lo demás) |
| F6 | Poda de paquetes/config | Elimina identityrepo/mailer/create_admin, poda config, reescribe `domain` | — (independiente de F5: solo borra/limpia código viejo) |
| F7 | Repositorio + privacidad | `internal/repository` unificado, `internal/privacy.CancelAccount`, `internal/testdb` con un esquema | F5, F6 |
| F8 | `internal/quiz` | Banco de preguntas + generación nickname/password, con pruebas unitarias de determinismo/colisión | F5 (solo necesita el esquema, no F6/F7) |
| F9 | `internal/auth`, transport, main.go, bootstrap admin | Registro en 3 pasos, login, age-up, ARCO simplificado, gestión de admins | F7, F8 |
| F10 | Frontend | Wizard de registro, login, panel de banco de preguntas con modal, panel de admins | Ninguna para maquetar (el contrato de API de §3 ya está fijo); F9 para probar contra backend real |
| F11 | Legal | Reescritura completa (ya prevista como F6 original, renumerada) | — (totalmente independiente) |

**Cómo paralelizar con varios agentes o sesiones simultáneas:**

```
F5 ──┬──► F6 ──► F7 ──┐
     │                ├──► F9 ──► (F10 integración real)
     └──────► F8 ─────┘

F10 (maquetado/estructura) ── en paralelo desde el día 1, sin esperar nada
F11 (legal) ── en paralelo desde el día 1, sin esperar nada
```

- **Ronda 1 (secuencial, corta):** un solo agente cierra **F5** primero — todo
  lo demás en backend depende de que el esquema exista, aunque sea en papel.
- **Ronda 2 (paralela):** una vez cerrada F5, tres agentes pueden trabajar a
  la vez sin pisarse: uno en **F6** (poda, no toca SQL), otro en **F8**
  (`internal/quiz`, paquete nuevo y aislado), y el frontend puede arrancar
  **F10** de maquetado (estructura de carpetas, componentes, esquemas Zod)
  usando el contrato de API de §3 como contrato fijo, sin esperar a que el
  backend real exista.
- **Ronda 3:** **F7** solo puede empezar cuando F6 esté cerrada (depende de
  `domain` reescrito). **F9** solo puede empezar cuando F7 y F8 estén
  cerradas (auth orquesta repository + quiz).
- **F11** (legal) no tiene dependencia técnica con nada de lo anterior — puede
  asignarse a un agente distinto desde ya.
- Cada fase sigue cerrando en su propio commit y con su propia entrada en
  `estado_proyecto.md`, igual que F1–F4, para que cualquier sesión nueva
  pueda leer qué está hecho sin tener que releer esta conversación.

## Verificación

Por fase, siguiendo el mismo método ya usado en F1–F4 (contenedor
`golang:1.22-bookworm`, nunca contra `usbi-database` persistente):

- F5: aplicar a base desechable, `\d accounts`/`\d+` confirma columnas y FKs
  fusionadas correctamente.
- F6–F9: `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...` en
  verde en cada fase; pruebas de integración contra esquema único desechable
  (registro→login→alias, admin no puede borrar admin, banco de preguntas no
  baja de 4).
- F10: build de Vite en verde, flujo manual registro→login→dashboard contra
  backend real de F9, casos de rechazo Zod/Go coincidentes.
- Arranque end-to-end del binario compilado contra base desechable real antes
  de cerrar F9, igual que se hizo en F4.

## Archivos críticos de referencia

- `backend/migrations/main/0001_esquema_principal.up.sql`,
  `backend/migrations/identity/0001_esquema_identidad.up.sql`
- `backend/internal/repository/account_queries.go` (RandomAlias, patrón a
  replicar)
- `backend/internal/privacy/privacy.go`, `backend/internal/transport/
  router.go`
- `backend/internal/crypto/hash.go` (reutilizar tal cual)
- `backend/internal/domain/models.go`, `backend/cmd/create_admin/main.go`
- `plan/00_Plan_maestro.md` (convención de fases y verificación a seguir)
