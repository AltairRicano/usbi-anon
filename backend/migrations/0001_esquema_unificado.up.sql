-- Migration: 0001_esquema_unificado.up.sql
-- Proyecto:  USBI-Anon — BASE DE DATOS ÚNICA (usbi_anon_db)
--
-- Reemplaza como BASELINE a los dos esquemas de F1 (migrations/identity/ y
-- migrations/main/). No es una migración incremental: aquellos scripts nunca se
-- aplicaron contra una base persistente (F1–F4 siempre usaron bases/esquemas
-- desechables), así que no hay un solo dato real que migrar. Ver
-- plan/04_Rediseno_identidad_gustos.md §1.
--
-- QUÉ CAMBIÓ Y POR QUÉ. El diseño anterior partía la persistencia en dos bases
-- porque una de ellas contenía el ÚNICO dato personal directo del sistema: el
-- correo electrónico. Al eliminarse el correo del producto (el registro pasa a
-- ser un cuestionario de gustos no sensibles del que se derivan nickname y
-- password), esa base deja de tener razón de existir: ya no hay nada que
-- aislar. La separación física se sustituye por algo más fuerte —— la ausencia
-- del dato.
--
-- PROMESA CENTRAL DEL PROYECTO, ahora para toda la base: ninguna columna
-- contiene nombre, correo, teléfono ni dato alguno de tutor. El único texto
-- libre escrito por una persona usuaria son las respuestas del cuestionario
-- (account_quiz_answers.answer_text), deliberadamente acotadas a gustos no
-- sensibles y validadas en Go antes de llegar aquí.
--
-- Nótese lo que NO se crea: la extensión pgcrypto. Al no quedar un solo campo
-- cifrado en el sistema, la dependencia desaparece. Que vuelva a hacer falta es
-- la señal de que se coló un dato sensible.
--
-- Convención: golang-migrate (pares .up.sql / .down.sql, SQL plano, sin
-- anotaciones de goose). Los UUID los provee la aplicación (Go, google/uuid
-- v7); gen_random_uuid() no se usa como valor por defecto en ninguna PK.

-- ═══════════════════════════════════════════════════════════════════════════
-- 1. IDENTIDAD ANÓNIMA
-- ═══════════════════════════════════════════════════════════════════════════

-- ── alias_adjectives / alias_nouns ───────────────────────────────────────────
-- Vocabulario cerrado para generar el alias visible del jugador.
--
-- POR QUÉ ASÍ Y NO UNA COLUMNA DE TEXTO: se decidió conservar un alias visible
-- ("Bienvenido, Jaguar Azul 42") generado por el sistema. Una columna
-- `display_alias VARCHAR` cumpliría hoy, pero cualquier cambio futuro de código
-- podría escribir en ella un nombre real, y nada en la base lo impediría — un
-- CHECK de formato tampoco, porque "Juan Pérez 12" encaja en cualquier patrón
-- razonable de "Adjetivo Sustantivo Número".
--
-- Guardando el alias como TRES ENTEROS contra vocabularios curados, la base
-- FÍSICAMENTE NO PUEDE almacenar un nombre. Es una garantía estructural, no una
-- convención, y se verifica de un vistazo en una auditoría.
--
-- Los adjetivos son invariantes en género (azul, verde, feliz, veloz…) para que
-- la concordancia funcione con cualquier sustantivo de la otra lista.
CREATE TABLE alias_adjectives (
    id   SMALLINT    PRIMARY KEY,
    word VARCHAR(24) NOT NULL UNIQUE
);

CREATE TABLE alias_nouns (
    id   SMALLINT    PRIMARY KEY,
    word VARCHAR(24) NOT NULL UNIQUE
);

INSERT INTO alias_adjectives (id, word) VALUES
    (1,'Azul'), (2,'Verde'), (3,'Feliz'), (4,'Veloz'), (5,'Audaz'), (6,'Fuerte'),
    (7,'Valiente'), (8,'Amable'), (9,'Alegre'), (10,'Noble'), (11,'Ágil'),
    (12,'Brillante'), (13,'Radiante'), (14,'Firme'), (15,'Leal'), (16,'Gentil'),
    (17,'Hábil'), (18,'Libre'), (19,'Elegante'), (20,'Constante'),
    (21,'Prudente'), (22,'Paciente'), (23,'Sutil'), (24,'Fiel');

-- Fauna de la región, coherente con la identidad institucional del proyecto.
INSERT INTO alias_nouns (id, word) VALUES
    (1,'Jaguar'), (2,'Quetzal'), (3,'Tucán'), (4,'Manatí'), (5,'Colibrí'),
    (6,'Tapir'), (7,'Ocelote'), (8,'Delfín'), (9,'Halcón'), (10,'Venado'),
    (11,'Coatí'), (12,'Tortuga'), (13,'Iguana'), (14,'Nutria'), (15,'Cenzontle'),
    (16,'Armadillo'), (17,'Mapache'), (18,'Búho'), (19,'Garza'), (20,'Pelícano'),
    (21,'Cangrejo'), (22,'Mariposa'), (23,'Libélula'), (24,'Zorro');

-- ── accounts ─────────────────────────────────────────────────────────────────
-- Fusión de las antiguas `identities` (base de identidad) y `accounts` (base
-- principal). Ancla de todas las claves foráneas de la base, y a la vez la
-- tabla de credenciales — algo que antes era impensable porque credencial
-- significaba correo electrónico. Ahora la credencial es un nickname derivado
-- de respuestas fragmentadas sobre gustos, y no identifica a nadie por sí solo.
--
-- Diferencias frente a la antigua `identities`:
--   · SE ELIMINA email / email_lookup_hash → no hay correo en el sistema.
--   · SE ELIMINA full_name y phone         → ya eliminados en F1.
--   · SE AÑADE  nickname                   → credencial de login, en claro,
--     buscada con `WHERE nickname = $1`. No lleva blind index HMAC porque no es
--     PII cifrada: es un identificador público generado por el sistema.
--   · status pierde 'pending_tutor_consent' → sin correo no hay doble opt-in de
--     tutor; un menor autoreportado juega de inmediato
--     (plan/04_Rediseno_identidad_gustos.md, decisión 3).
CREATE TABLE accounts (
    id                         UUID        PRIMARY KEY,
    -- Credencial de login. Formato cerrado: minúsculas y dígitos, 6–20
    -- caracteres. El CHECK no es cosmético: impide que un cambio futuro de
    -- código escriba aquí un nombre propio, un correo o cualquier texto con
    -- espacios, acentos o arroba.
    nickname                   VARCHAR(32) NOT NULL UNIQUE
        CHECK (nickname ~ '^[a-z0-9]{6,20}$'),
    password_hash              VARCHAR     NOT NULL, -- Argon2id, irreversible
    token_version              INTEGER     NOT NULL DEFAULT 1,
    is_adult                   BOOLEAN     NOT NULL, -- autorreporte puro
    role                       VARCHAR     NOT NULL
        CHECK (role IN ('player', 'admin', 'operator', 'director')),
    status                     VARCHAR     NOT NULL
        CHECK (status IN ('active', 'suspended', 'deleted')),
    -- Contador de intentos de transición a mayoría de edad (Ley 251, máx. 3).
    age_up_attempts            SMALLINT    NOT NULL DEFAULT 0,
    -- Alias visible, generado por el sistema. Ver alias_adjectives/alias_nouns.
    -- Se sortea UNA SOLA VEZ, en el INSERT de registro: al no haber dos bases
    -- que reconciliar, desaparece el upsert de login/refresh de F4.
    alias_adjective_id         SMALLINT    NOT NULL REFERENCES alias_adjectives(id) ON DELETE RESTRICT,
    alias_noun_id              SMALLINT    NOT NULL REFERENCES alias_nouns(id)      ON DELETE RESTRICT,
    alias_number               SMALLINT    NOT NULL CHECK (alias_number BETWEEN 0 AND 999),
    privacy_notice_version     VARCHAR     NOT NULL,
    privacy_notice_accepted_at TIMESTAMPTZ NOT NULL,
    -- Sello HMAC de no repudio del aviso de privacidad. Antes se calculaba
    -- sobre el correo; ahora sobre (id + version + accepted_at).
    privacy_acceptance_hash    BYTEA       NOT NULL,
    -- Versión de la clave HMAC con la que se selló la fila. Se conserva pese a
    -- que ya no hay cifrado: sin ella, rotar HMAC_SECRET invalidaría todos los
    -- sellos existentes y con ellos la evidencia de no repudio.
    crypto_key_version         SMALLINT    NOT NULL,
    created_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at              TIMESTAMPTZ,
    deleted_at                 TIMESTAMPTZ,
    deletion_reason            VARCHAR
);

COMMENT ON TABLE accounts IS
    'Cuenta de jugador o administrador. No contiene ningún dato personal '
    'directo: el nickname es un identificador generado por el sistema a partir '
    'de fragmentos de respuestas sobre gustos, no un nombre. Añadir aquí una '
    'columna que pueda contener un nombre, un correo o un teléfono es '
    'incompatible con el diseño del proyecto.';
COMMENT ON COLUMN accounts.id IS
    'UUID v7 generado por la aplicación. Es el subject del JWT y la llave de '
    'todo el progreso. La cancelación de cuenta conserva la fila con el '
    'nickname sobrescrito por relleno aleatorio, para no romper el no repudio '
    'de las bitácoras.';
COMMENT ON COLUMN accounts.nickname IS
    'Credencial de login. Al cancelar una cuenta debe sobrescribirse con '
    'relleno aleatorio [a-z0-9] de 20 caracteres — cumple el CHECK, libera el '
    'valor original para reuso y no deja rastro de las respuestas que lo '
    'originaron.';

CREATE INDEX accounts_role_status_idx ON accounts (role, status);
CREATE INDEX accounts_status_idx      ON accounts (status);

-- Índices parciales para las dos rutinas de retención automática que
-- sobreviven en internal/maintenance (la tercera, la purga de registros
-- atascados en 'pending_tutor_consent', desaparece con el flujo de tutor).
CREATE INDEX accounts_inactive_players_idx
    ON accounts (COALESCE(last_login_at, created_at))
    WHERE role = 'player' AND status = 'active' AND deleted_at IS NULL;

CREATE INDEX accounts_pending_cancel_idx
    ON accounts (updated_at)
    WHERE role = 'player'
      AND status = 'suspended'
      AND deletion_reason = 'inactive_suspension_pending_cancel'
      AND deleted_at IS NULL;

-- Vista de conveniencia: compone el alias legible sin que ninguna tabla
-- almacene la cadena. El alias NO es un identificador: no es único y jamás debe
-- usarse como clave de búsqueda. La llave siempre es accounts.id.
CREATE VIEW account_aliases AS
SELECT a.id,
       n.word || ' ' || adj.word || ' ' || a.alias_number::text AS display_alias
FROM accounts a
JOIN alias_adjectives adj ON adj.id = a.alias_adjective_id
JOIN alias_nouns      n   ON n.id   = a.alias_noun_id;

-- ── refresh_tokens ───────────────────────────────────────────────────────────
-- Tokens opacos de sesión (7 días) con revocación del lado servidor.
CREATE TABLE refresh_tokens (
    id         UUID        PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    token_hash BYTEA       NOT NULL UNIQUE,
    issued_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ
);

CREATE INDEX refresh_tokens_user_active_idx
    ON refresh_tokens (user_id)
    WHERE revoked_at IS NULL;

-- Soporta la purga periódica de tokens vencidos o revocados hace más de 7 días.
CREATE INDEX refresh_tokens_expires_at_idx ON refresh_tokens (expires_at);

-- ═══════════════════════════════════════════════════════════════════════════
-- 2. CUESTIONARIO DE REGISTRO
-- ═══════════════════════════════════════════════════════════════════════════

-- ── registration_questions ───────────────────────────────────────────────────
-- Banco de preguntas administrable. El registro muestra un subconjunto
-- aleatorio de las activas (máximo registration_settings.max_questions_shown);
-- el resto queda en reserva y rota entre registros.
--
-- LA REGLA "MÍNIMO 4 PREGUNTAS ACTIVAS" NO SE IMPLEMENTA AQUÍ. Podría hacerse
-- con un trigger, pero este proyecto mantiene la lógica de negocio en Go de
-- forma consistente (ver la nota de las funciones PL/pgSQL muertas en §5). Se
-- valida en internal/quiz dentro de la misma transacción que el DELETE/UPDATE,
-- que además necesita devolver un 409 con mensaje entendible.
CREATE TABLE registration_questions (
    id            UUID         PRIMARY KEY,
    question_text VARCHAR(280) NOT NULL CHECK (btrim(question_text) <> ''),
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    display_order SMALLINT     NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Soporta el muestreo aleatorio del registro, que solo mira las activas.
CREATE INDEX registration_questions_active_idx
    ON registration_questions (display_order)
    WHERE is_active;

COMMENT ON TABLE registration_questions IS
    'Preguntas sobre gustos NO SENSIBLES. Añadir aquí una pregunta que pida un '
    'nombre, una escuela, una dirección, una fecha de nacimiento o cualquier '
    'dato identificable rompe la promesa central del proyecto: las respuestas '
    'se guardan en claro en account_quiz_answers.';

-- Banco inicial. Son las cinco preguntas con las que se cerró el diseño; el
-- sistema necesita al menos cuatro activas para poder registrar a nadie, así
-- que sembrarlas aquí es parte del baseline, no un dato de prueba.
INSERT INTO registration_questions (id, question_text, is_active, display_order) VALUES
    ('018fd2b4-3f0d-7c00-8000-000000000201', '¿Cuál es tu color favorito?',              TRUE, 1),
    ('018fd2b4-3f0d-7c00-8000-000000000202', '¿Cuál es tu animal favorito?',             TRUE, 2),
    ('018fd2b4-3f0d-7c00-8000-000000000203', '¿Cuál es tu materia escolar favorita?',    TRUE, 3),
    ('018fd2b4-3f0d-7c00-8000-000000000204', '¿Cuál es tu número favorito?',             TRUE, 4),
    ('018fd2b4-3f0d-7c00-8000-000000000205', '¿Cuántas mascotas tienes?',                TRUE, 5);

-- ── registration_settings ────────────────────────────────────────────────────
-- Fila única (id = 1, forzado por CHECK). Un ajuste de configuración editable
-- desde el panel admin no justifica una tabla clave-valor genérica.
CREATE TABLE registration_settings (
    id                  SMALLINT    PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    max_questions_shown SMALLINT    NOT NULL
        CHECK (max_questions_shown BETWEEN 4 AND 10),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO registration_settings (id, max_questions_shown) VALUES (1, 5);

-- ── account_quiz_answers ─────────────────────────────────────────────────────
-- Respuestas de texto libre del cuestionario, en claro y por diseño.
--
-- POR QUÉ SE GUARDAN. Sin correo electrónico no hay forma de recuperar una
-- cuenta olvidada por el canal habitual. La recuperación acordada es que un
-- administrador compare a ojo las respuestas que la persona reingresa contra
-- las guardadas y resetee el password manualmente (decisión 4 del rediseño).
-- Eso exige persistirlas.
--
-- POR QUÉ NO VAN CIFRADAS. Son gustos no sensibles (color, animal, materia) y
-- el admin necesita leerlas para compararlas. Cifrarlas con una clave que el
-- propio backend posee no añadiría protección real frente al escenario que
-- importa aquí, y reintroduciría pgcrypto en un esquema que se limpió a
-- propósito. La protección es de acceso: el endpoint que las expone exige rol
-- admin y queda auditado en audit_log.
CREATE TABLE account_quiz_answers (
    id                     UUID         PRIMARY KEY,
    user_id                UUID         NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    -- SET NULL: borrar una pregunta del banco no puede borrar la respuesta que
    -- alguien dio, porque es la única vía de recuperación de esa cuenta.
    question_id            UUID         REFERENCES registration_questions(id) ON DELETE SET NULL,
    -- Congela el texto que la persona vio realmente. Si la pregunta se edita
    -- después ("¿tu color favorito?" → "¿tu color menos favorito?"), la
    -- comparación del admin seguiría siendo válida.
    question_text_snapshot VARCHAR(280) NOT NULL,
    answer_text            VARCHAR(200) NOT NULL CHECK (btrim(answer_text) <> ''),
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX account_quiz_answers_user_idx     ON account_quiz_answers (user_id);
CREATE INDEX account_quiz_answers_question_idx ON account_quiz_answers (question_id);

COMMENT ON COLUMN account_quiz_answers.answer_text IS
    'Texto libre escrito por la persona usuaria — el único de toda la base. '
    'Validado en Go (y en el frontend) contra control chars, JSON y HTML antes '
    'de llegar aquí. Nunca se usa para construir SQL ni se devuelve a nadie '
    'salvo a un admin autenticado.';

-- ═══════════════════════════════════════════════════════════════════════════
-- 3. OPERACIÓN Y ARQUITECTURA OFFLINE
-- ═══════════════════════════════════════════════════════════════════════════

-- ── devices ──────────────────────────────────────────────────────────────────
-- CAMBIO FRENTE A ../usbi: allí existía `device_label VARCHAR NOT NULL`, texto
-- libre escrito por la persona usuaria. En la práctica eso se llena con "iPad
-- de Sofía". Aquí se sustituye por un vocabulario cerrado, que basta para
-- distinguir dispositivos en la pantalla de sesiones y no admite PII.
CREATE TABLE devices (
    id              UUID        PRIMARY KEY,
    user_id         UUID        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    device_kind     VARCHAR     NOT NULL
        CHECK (device_kind IN ('movil', 'tablet', 'laptop', 'escritorio', 'otro')),
    platform        VARCHAR     NOT NULL CHECK (platform IN ('web', 'tauri')),
    registered_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Bandera de borrado local: cuando es TRUE (la cancelación de cuenta la
    -- activa), la siguiente respuesta de sync DEBE inyectar wipe_local_data =
    -- true para que el dispositivo borre su SQLite local.
    wipe_local_data BOOLEAN     NOT NULL DEFAULT FALSE,
    revoked_at      TIMESTAMPTZ,
    -- Única compuesta: habilita la FK compuesta desde sync_events, que impide
    -- que un usuario reclame el dispositivo de otro.
    CONSTRAINT uq_devices_id_user UNIQUE (id, user_id)
);

CREATE INDEX devices_user_id_idx ON devices (user_id);

-- ── sync_events ──────────────────────────────────────────────────────────────
CREATE TABLE sync_events (
    id                 UUID        PRIMARY KEY,
    user_id            UUID        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    device_id          UUID        NOT NULL,
    -- Sin PII: solo eventos técnicos de progreso. Lo garantiza el tipado
    -- estricto de domain.SyncPayload en Go, que rechaza campos desconocidos.
    payload            JSONB       NOT NULL,
    payload_hash       BYTEA       NOT NULL,
    hmac_signature     BYTEA       NOT NULL,
    crypto_key_version SMALLINT    NOT NULL,
    hmac_valid         BOOLEAN     NOT NULL,
    status             VARCHAR     NOT NULL
        CHECK (status IN ('pending', 'processed', 'rejected')),
    received_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at       TIMESTAMPTZ,
    rejection_reason   TEXT,
    FOREIGN KEY (device_id, user_id)
        REFERENCES devices(id, user_id) ON DELETE CASCADE
);

CREATE INDEX sync_events_user_status_idx ON sync_events (user_id, status);
CREATE INDEX sync_events_device_id_idx   ON sync_events (device_id);

-- ═══════════════════════════════════════════════════════════════════════════
-- 4. CONTENIDO EDUCATIVO (retiro en dos pasos)
-- ═══════════════════════════════════════════════════════════════════════════
--
-- RETIRAR CONTENIDO ES UN PROCESO DE DOS PASOS, deliberadamente separados:
--
--   1. ARCHIVAR (reversible, no libera espacio). Es el borrado lógico que ya
--      existía: `deleted_at` en levels, `deleted_at`/`archived_at` en sections.
--      El nivel deja de jugarse y de listarse, pero todo sigue en disco y se
--      puede revertir.
--   2. PURGAR (irreversible, libera espacio). Un DELETE físico, acción de admin
--      separada y con confirmación explícita. Solo debe permitirse sobre filas
--      que ya estén archivadas (`deleted_at IS NOT NULL`) — validarlo en Go.
--      Dispara los CASCADE de §5 y libera el almacenamiento de verdad.
--
-- El paso 1 por sí solo NO libera un byte. Un sistema que solo hiciera borrado
-- lógico —como el heredado de ../usbi— crece para siempre, y con 20 GB fijos
-- eso es una cuenta regresiva, no una arquitectura.
--
-- levels.section_id sigue siendo RESTRICT a propósito: purgar una sección exige
-- purgar antes sus niveles, uno por uno y con sus contadores acumulados. Un
-- CASCADE aquí convertiría un clic en la sección en un borrado masivo silencioso.

CREATE TABLE sections (
    id                  UUID        PRIMARY KEY,
    title               VARCHAR     NOT NULL,
    color               VARCHAR     NOT NULL,
    description         TEXT        NOT NULL DEFAULT '',
    created_by_admin_id UUID        REFERENCES accounts(id) ON DELETE SET NULL,
    is_published        BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    archived_at         TIMESTAMPTZ
);

CREATE INDEX sections_created_by_idx ON sections (created_by_admin_id);

CREATE TABLE levels (
    id                  UUID        PRIMARY KEY,
    section_id          UUID        NOT NULL REFERENCES sections(id) ON DELETE RESTRICT,
    title               VARCHAR     NOT NULL,
    color               VARCHAR     NOT NULL,
    -- Vocabulario oficial de plantillas. En ../usbi este CHECK nació NOT VALID
    -- en 0002 y no se validó hasta 0011; aquí es válido desde el inicio.
    template_type       VARCHAR     NOT NULL
        CHECK (template_type IN ('trivia', 'puzzle', 'word_search', 'fake_news',
                                 'crossword', 'memory', 'snakes_ladders')),
    -- Solo estructura y metadatos. Sin binarios.
    content             JSONB       NOT NULL,
    difficulty          INTEGER     NOT NULL CHECK (difficulty BETWEEN 1 AND 10),
    is_published        BOOLEAN     NOT NULL DEFAULT FALSE,
    created_by_admin_id UUID        REFERENCES accounts(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    deleted_by          UUID        REFERENCES accounts(id) ON DELETE SET NULL
);

CREATE INDEX levels_section_id_idx ON levels (section_id);
CREATE INDEX levels_created_by_idx ON levels (created_by_admin_id);
CREATE INDEX levels_deleted_by_idx ON levels (deleted_by);

-- ═══════════════════════════════════════════════════════════════════════════
-- 5. PROGRESO Y RENDICIÓN DE CUENTAS
-- ═══════════════════════════════════════════════════════════════════════════
--
-- ROTACIÓN DE NIVELES POR TEMPORADAS — regla de negocio que gobierna esta
-- sección entera. El almacenamiento del servidor es finito (~20 GB) y no se va
-- a ampliar pagando más: el contenido educativo se rota entre temporadas,
-- retirando niveles viejos para meter otros nuevos.
--
-- La regla, en una frase: **retirar un nivel libera su almacenamiento pero
-- NUNCA le quita a un jugador la experiencia que ganó jugándolo.**
--
-- De ahí salen las tres direcciones de borrado de esta sección, que no son
-- caprichosas y no deben "uniformarse" en una futura limpieza:
--
--   · level_attempts   → CASCADE   (se va: es el grueso del volumen)
--   · player_progress  → CASCADE   (se va: sin nivel no hay progreso "de" él,
--                                   pero antes se acumula en la tabla de
--                                   totales retirados, ver más abajo)
--   · experience_history → SET NULL (SE QUEDA: es la fuente de verdad del XP
--                                   total del jugador; solo pierde el vínculo)
--
-- Antes de borrar un nivel, la purga DEBE acumular los contadores en
-- account_retired_progress dentro de la misma transacción, o el jugador vería
-- caer sus "niveles completados" e "intentos totales". La consulta exacta está
-- documentada en el comentario de esa tabla.

CREATE TABLE player_progress (
    user_id            UUID    NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    -- CASCADE, no RESTRICT: retirar un nivel debe poder borrar el progreso
    -- ligado a él. El XP no se pierde porque no vive aquí, vive en
    -- experience_history; los contadores se preservan en
    -- account_retired_progress.
    level_id           UUID    NOT NULL REFERENCES levels(id)   ON DELETE CASCADE,
    best_score         INTEGER NOT NULL DEFAULT 0,
    xp_total_for_level INTEGER NOT NULL DEFAULT 0,
    attempts_count     INTEGER NOT NULL DEFAULT 0,
    first_completed_at TIMESTAMPTZ,
    last_completed_at  TIMESTAMPTZ,
    PRIMARY KEY (user_id, level_id)
);

CREATE INDEX player_progress_level_id_idx ON player_progress (level_id);

-- ── account_retired_progress ─────────────────────────────────────────────────
-- Contadores acumulados de los niveles YA RETIRADOS de una cuenta.
--
-- POR QUÉ EXISTE. "Niveles completados" e "intentos totales" se calculan
-- contando filas de player_progress, y esas filas se van con el nivel al
-- purgarlo. Sin esta tabla, rotar la temporada le bajaría a un jugador el
-- contador de 37 niveles completados a 12 — exactamente el "perder progreso"
-- que la rotación no debe causar. El XP total no necesita nada de esto: lo
-- sostiene experience_history, cuyas filas sobreviven.
--
-- Es acumulativa: cada purga SUMA a lo que ya hubiera. La purga de un nivel
-- debe ejecutar esto ANTES del DELETE, en la misma transacción:
--
--   INSERT INTO account_retired_progress AS arp
--       (user_id, levels_completed, attempts_total)
--   SELECT user_id,
--          COUNT(*) FILTER (WHERE first_completed_at IS NOT NULL),
--          COALESCE(SUM(attempts_count), 0)
--   FROM player_progress
--   WHERE level_id = $1
--   GROUP BY user_id
--   ON CONFLICT (user_id) DO UPDATE SET
--       levels_completed = arp.levels_completed + EXCLUDED.levels_completed,
--       attempts_total   = arp.attempts_total   + EXCLUDED.attempts_total,
--       updated_at       = NOW();
--
-- Y GetUserProgressTotals pasa a sumar las dos fuentes (vivos + retirados).
CREATE TABLE account_retired_progress (
    user_id          UUID        PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    levels_completed INTEGER     NOT NULL DEFAULT 0 CHECK (levels_completed >= 0),
    attempts_total   INTEGER     NOT NULL DEFAULT 0 CHECK (attempts_total >= 0),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE account_retired_progress IS
    'Contadores de niveles ya purgados. Solo crece; nunca se decrementa. No '
    'guarda XP: el XP de un nivel retirado sigue en experience_history con '
    'level_id en NULL, y sumarlo aquí lo contaría dos veces.';

-- ── level_attempts ───────────────────────────────────────────────────────────
-- Particionada por rango desde el inicio. En ../usbi se creó sin particionar en
-- 0001 y se reconstruyó con DROP TABLE ... CASCADE en 0007; aquí nace ya
-- particionada y con partición DEFAULT (que 0008 tuvo que parchar después).
--
-- La partición DEFAULT es solo red de seguridad: internal/dbmaint crea las
-- particiones anuales por adelantado (año actual + 2), así que en operación
-- normal debe quedar vacía.
--
-- NOTA DE CONCURRENCIA (RF15): la resolución de attempt_number usa bloqueo
-- transaccional en Go, para evitar carreras entre el camino online y el sync.
CREATE TABLE level_attempts (
    id             UUID        NOT NULL,
    user_id        UUID        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    -- CASCADE, no RESTRICT: esta es la tabla que más crece de todo el sistema
    -- (una fila por intento, de cada jugador, de cada nivel) y por tanto la que
    -- de verdad libera espacio al rotar la temporada. El intento en sí no es
    -- progreso conservable: su aportación al jugador ya está liquidada en
    -- experience_history (el XP) y en account_retired_progress (el conteo).
    level_id       UUID        NOT NULL REFERENCES levels(id)   ON DELETE CASCADE,
    attempt_date   DATE        NOT NULL,
    attempt_number INTEGER     NOT NULL, -- 1 = 100 % XP, 2-3 = 50 %, >3 = 0 %
    xp_awarded     INTEGER     NOT NULL,
    completed      BOOLEAN     NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, attempt_date),
    UNIQUE (user_id, level_id, attempt_date, attempt_number)
) PARTITION BY RANGE (attempt_date);

CREATE TABLE level_attempts_2026 PARTITION OF level_attempts
    FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');
CREATE TABLE level_attempts_2027 PARTITION OF level_attempts
    FOR VALUES FROM ('2027-01-01') TO ('2028-01-01');
CREATE TABLE level_attempts_2028 PARTITION OF level_attempts
    FOR VALUES FROM ('2028-01-01') TO ('2029-01-01');
CREATE TABLE level_attempts_default PARTITION OF level_attempts DEFAULT;

CREATE INDEX level_attempts_user_level_date_idx
    ON level_attempts (user_id, level_id, attempt_date);
CREATE INDEX level_attempts_level_id_idx ON level_attempts (level_id);

-- ── daily_streak ─────────────────────────────────────────────────────────────
CREATE TABLE daily_streak (
    user_id       UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    activity_date DATE NOT NULL,
    PRIMARY KEY (user_id, activity_date)
) PARTITION BY RANGE (activity_date);

CREATE TABLE daily_streak_2026 PARTITION OF daily_streak
    FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');
CREATE TABLE daily_streak_2027 PARTITION OF daily_streak
    FOR VALUES FROM ('2027-01-01') TO ('2028-01-01');
CREATE TABLE daily_streak_2028 PARTITION OF daily_streak
    FOR VALUES FROM ('2028-01-01') TO ('2029-01-01');
CREATE TABLE daily_streak_default PARTITION OF daily_streak DEFAULT;

-- ── badges / user_badges ─────────────────────────────────────────────────────
CREATE TABLE badges (
    id           UUID    PRIMARY KEY,
    name         VARCHAR NOT NULL,
    xp_threshold INTEGER NOT NULL,
    icon_key     VARCHAR NOT NULL
);

-- Insignias base del MVP (origen: ../usbi/backend/migrations/0004).
INSERT INTO badges (id, name, xp_threshold, icon_key) VALUES
    ('018fd2b4-3f0d-7c00-8000-000000000101', 'Primeros pasos',  20,  'first_steps'),
    ('018fd2b4-3f0d-7c00-8000-000000000102', 'Explorador USBI', 60,  'explorer'),
    ('018fd2b4-3f0d-7c00-8000-000000000103', 'Constancia',      120, 'streak_builder'),
    ('018fd2b4-3f0d-7c00-8000-000000000104', 'Experto USBI',    240, 'expert');

CREATE TABLE user_badges (
    user_id   UUID        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    badge_id  UUID        NOT NULL REFERENCES badges(id)   ON DELETE RESTRICT,
    earned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, badge_id)
);

CREATE INDEX user_badges_badge_id_idx ON user_badges (badge_id);

-- ═══════════════════════════════════════════════════════════════════════════
-- 6. AUDITORÍA Y SEGURIDAD
-- ═══════════════════════════════════════════════════════════════════════════
--
-- Este proyecto NO opera bajo un régimen de derechos ARCO (acceso,
-- rectificación, cancelación, oposición) con trámite de aprobación admin: no
-- hay correo ni ningún otro dato personal directo que un titular deba
-- reclamar por esa vía, y la cancelación de cuenta es autoservicio inmediato
-- (DELETE /auth/me, internal/privacy.CancelAccount) — cualquier jugador puede
-- eliminar su cuenta cuando quiera, sin trámite ni aprobación de por medio.
-- La tabla `arco_requests` (heredada de ../usbi, donde sí regía ese marco por
-- el correo electrónico) se eliminó por completo: no aplica aquí.

-- ── experience_history ───────────────────────────────────────────────────────
-- APPEND-ONLY, y FUENTE DE VERDAD DEL XP TOTAL DEL JUGADOR: el total se calcula
-- con SUM(xp_gained) sobre esta tabla, no con un contador guardado en accounts.
--
-- Por eso sus dos referencias se anulan en vez de arrastrar la fila:
--   · user_id  → NULL al cancelar la cuenta (la fila del libro mayor
--                sobrevive para el no repudio, sin vínculo con la cuenta).
--   · level_id → NULL al retirar un nivel en la rotación de temporadas. La fila
--                sobrevive con su xp_gained intacto: el jugador conserva la
--                experiencia que ganó aunque el nivel ya no exista en el
--                servidor. Es la pieza central de la regla de rotación (§5).
--
-- Se conserva el detalle fila por fila —en vez de colapsarlo en un total— para
-- no romper el append-only ni la evidencia por evento (source /
-- verification_method) que sostiene la defensa antitrampas del sistema offline.
CREATE TABLE experience_history (
    id                  UUID        PRIMARY KEY,
    user_id             UUID        REFERENCES accounts(id) ON DELETE SET NULL,
    -- NULLABLE a propósito: NULL significa "el nivel que originó este XP fue
    -- retirado del servidor". Las consultas que muestren historial deben usar
    -- LEFT JOIN contra levels y tolerar el nivel ausente.
    level_id            UUID        REFERENCES levels(id) ON DELETE SET NULL,
    event_type          VARCHAR     NOT NULL,
    xp_gained           INTEGER     NOT NULL,
    source              VARCHAR     NOT NULL,
    verification_method VARCHAR     NOT NULL,
    sync_event_id       UUID        REFERENCES sync_events(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (
        (source = 'online'       AND verification_method = 'online_direct') OR
        (source = 'offline_sync' AND verification_method = 'hmac_offline')
    )
);

CREATE INDEX experience_history_user_id_idx    ON experience_history (user_id);
CREATE INDEX experience_history_level_id_idx   ON experience_history (level_id);
CREATE INDEX experience_history_sync_event_idx ON experience_history (sync_event_id);

-- ── audit_log ────────────────────────────────────────────────────────────────
-- Fusión de identity_audit_log + admin_audit_log. Se partieron en dos en F1 por
-- una razón concreta: el before_state de "se editó el correo de la cuenta X"
-- ES el correo, y dejarlo en la base principal habría filtrado PII a la base
-- que promete no tenerla. Sin correo en el sistema esa razón desaparece, y
-- mantener dos bitácoras idénticas solo complicaría la consulta.
--
-- APPEND-ONLY: sin UPDATE ni DELETE, salvo el SET NULL del actor cuando se
-- cancela su cuenta. Lo garantiza el trigger de más abajo.
--
-- before_state / after_state NO DEBEN contener datos identificables. La única
-- tabla de la que podrían copiarse respuestas de texto libre es
-- account_quiz_answers: al auditar su consulta, registrar el user_id y el
-- conteo, nunca el contenido.
CREATE TABLE audit_log (
    id               UUID        PRIMARY KEY,
    actor_account_id UUID        REFERENCES accounts(id) ON DELETE SET NULL,
    action           VARCHAR     NOT NULL,
    entity_type      VARCHAR     NOT NULL,
    entity_id        UUID,
    before_state     JSONB,
    after_state      JSONB,
    ip_address       INET        NOT NULL,
    user_agent       TEXT        NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX audit_log_actor_idx      ON audit_log (actor_account_id);
CREATE INDEX audit_log_created_at_idx ON audit_log (created_at);

-- ── security_incidents ───────────────────────────────────────────────────────
CREATE TABLE security_incidents (
    id                      UUID        PRIMARY KEY,
    detected_at             TIMESTAMPTZ NOT NULL,
    reported_at             TIMESTAMPTZ,
    severity                VARCHAR     NOT NULL
        CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    affected_scope          TEXT        NOT NULL,
    description             TEXT        NOT NULL,
    containment_actions     TEXT        NOT NULL,
    resolved_at             TIMESTAMPTZ,
    reported_to_cutai       BOOLEAN     NOT NULL DEFAULT FALSE,
    notified_to_cutai_at    TIMESTAMPTZ,
    notified_to_subjects_at TIMESTAMPTZ,
    evidence_hash           BYTEA       NOT NULL
);

-- Estas tres columnas las redacta un operador en pleno incidente — el momento
-- exacto en que alguien escribe "se filtró la cuenta de Juan Pérez". No se
-- puede impedir con un CHECK: queda como control documental, reforzado con
-- validación en internal/incidents. Referirse siempre a los titulares por UUID.
COMMENT ON COLUMN security_incidents.affected_scope IS
    'PROHIBIDO nombrar titulares. Referirse a las personas afectadas por UUID '
    'o por conteo agregado.';
COMMENT ON COLUMN security_incidents.description IS
    'PROHIBIDO nombrar titulares. Referirse a las personas afectadas por UUID '
    'o por conteo agregado.';
COMMENT ON COLUMN security_incidents.containment_actions IS
    'PROHIBIDO nombrar titulares. Referirse a las personas afectadas por UUID '
    'o por conteo agregado.';

-- ── Invariante append-only ───────────────────────────────────────────────────
-- Se conserva de ../usbi/backend/migrations/0006: es la única lógica que 0012
-- dejó viva a propósito, porque es un invariante que debe sostenerse sin
-- importar qué código escriba en estas tablas. Las otras cinco funciones
-- PL/pgSQL de 0006 estaban muertas y NO se recrean: Go es la única fuente de
-- verdad del cálculo de XP y de la resolución ARCO.
CREATE OR REPLACE FUNCTION enforce_append_only_ledgers()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION '% es append-only', TG_TABLE_NAME USING ERRCODE = '55000';
    END IF;

    -- Mutaciones permitidas: anular UNA de las dos referencias de la fila,
    -- dejando todo lo demás —y en particular xp_gained— byte a byte idéntico.
    --
    --   · user_id  → NULL: cancelación de cuenta (autoservicio).
    --   · level_id → NULL: retiro de un nivel en la rotación de temporadas.
    --
    -- Nótese que nunca se permite tocar xp_gained: ni la cancelación de una
    -- cuenta ni el retiro de un nivel pueden alterar la experiencia registrada.
    -- Esa imposibilidad, sostenida por la base y no por el código, es lo que
    -- hace que "rotar niveles no quita XP" sea una garantía y no una promesa.
    IF TG_TABLE_NAME = 'experience_history' THEN
        IF OLD.user_id IS NOT NULL
           AND NEW.user_id IS NULL
           AND NEW.id                  IS NOT DISTINCT FROM OLD.id
           AND NEW.level_id            IS NOT DISTINCT FROM OLD.level_id
           AND NEW.event_type          IS NOT DISTINCT FROM OLD.event_type
           AND NEW.xp_gained           IS NOT DISTINCT FROM OLD.xp_gained
           AND NEW.source              IS NOT DISTINCT FROM OLD.source
           AND NEW.verification_method IS NOT DISTINCT FROM OLD.verification_method
           AND NEW.sync_event_id       IS NOT DISTINCT FROM OLD.sync_event_id
           AND NEW.created_at          IS NOT DISTINCT FROM OLD.created_at THEN
            RETURN NEW;
        END IF;

        IF OLD.level_id IS NOT NULL
           AND NEW.level_id IS NULL
           AND NEW.id                  IS NOT DISTINCT FROM OLD.id
           AND NEW.user_id             IS NOT DISTINCT FROM OLD.user_id
           AND NEW.event_type          IS NOT DISTINCT FROM OLD.event_type
           AND NEW.xp_gained           IS NOT DISTINCT FROM OLD.xp_gained
           AND NEW.source              IS NOT DISTINCT FROM OLD.source
           AND NEW.verification_method IS NOT DISTINCT FROM OLD.verification_method
           AND NEW.sync_event_id       IS NOT DISTINCT FROM OLD.sync_event_id
           AND NEW.created_at          IS NOT DISTINCT FROM OLD.created_at THEN
            RETURN NEW;
        END IF;
    ELSIF TG_TABLE_NAME = 'audit_log' THEN
        IF OLD.actor_account_id IS NOT NULL
           AND NEW.actor_account_id IS NULL
           AND NEW.id           IS NOT DISTINCT FROM OLD.id
           AND NEW.action       IS NOT DISTINCT FROM OLD.action
           AND NEW.entity_type  IS NOT DISTINCT FROM OLD.entity_type
           AND NEW.entity_id    IS NOT DISTINCT FROM OLD.entity_id
           AND NEW.before_state IS NOT DISTINCT FROM OLD.before_state
           AND NEW.after_state  IS NOT DISTINCT FROM OLD.after_state
           AND NEW.ip_address   IS NOT DISTINCT FROM OLD.ip_address
           AND NEW.user_agent   IS NOT DISTINCT FROM OLD.user_agent
           AND NEW.created_at   IS NOT DISTINCT FROM OLD.created_at THEN
            RETURN NEW;
        END IF;
    END IF;

    RAISE EXCEPTION '% es append-only; solo se permite el SET NULL de seudonimización',
        TG_TABLE_NAME USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER experience_history_append_only_trg
BEFORE UPDATE OR DELETE ON experience_history
FOR EACH ROW EXECUTE FUNCTION enforce_append_only_ledgers();

CREATE TRIGGER audit_log_append_only_trg
BEFORE UPDATE OR DELETE ON audit_log
FOR EACH ROW EXECUTE FUNCTION enforce_append_only_ledgers();

-- ═══════════════════════════════════════════════════════════════════════════
-- 7. DOCUMENTACIÓN DE COLUMNAS
-- ═══════════════════════════════════════════════════════════════════════════
-- El nombre de columna `user_id` se conserva en las tablas de progreso,
-- contenido y auditoría heredadas de ../usbi, para no invalidar la copia
-- verbatim de la capa de repositorio en Go. Lo que cambió no es el nombre sino
-- el contenido: aquí un user_id es un UUID sin ningún dato personal asociado.
-- Solo audit_log (tabla nueva por fusión) usa actor_account_id — ahí sí tiene
-- sentido semántico distinto: identifica a quien ejecutó la acción, no de
-- quién es el dato, y puede quedar NULL si esa cuenta se cancela.
-- account_quiz_answers.user_id y account_retired_progress.user_id ya no son
-- excepción: se renombraron desde account_id el 2026-09-02 por la misma
-- razón que refresh_tokens — no había justificación semántica para diferir.
COMMENT ON COLUMN devices.user_id IS
    'UUID de accounts.id. Esta base no guarda ningún dato que permita resolverlo '
    'a una persona por sí sola.';
COMMENT ON COLUMN player_progress.user_id IS
    'UUID de accounts.id. Esta base no guarda ningún dato que permita resolverlo '
    'a una persona por sí sola.';
COMMENT ON COLUMN experience_history.user_id IS
    'UUID de accounts.id. NULL tras la cancelación de la cuenta: la fila del '
    'libro mayor sobrevive para el no repudio, sin vínculo con la cuenta.';
