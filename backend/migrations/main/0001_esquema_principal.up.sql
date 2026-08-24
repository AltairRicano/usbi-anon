-- Migration: 0001_esquema_principal.up.sql
-- Proyecto:  USBI-Anon — BASE DE DATOS PRINCIPAL (usbi_anon_db)
--
-- Esta base contiene el progreso, el contenido educativo, las insignias, las
-- rachas, los dispositivos y la bitácora de contenido. Todo indexado por UUID.
--
-- PROMESA CENTRAL DEL PROYECTO: ninguna columna de esta base contiene nombre,
-- correo, teléfono ni dato alguno de tutor. Ninguna columna admite texto libre
-- escrito por una persona usuaria. Si una migración futura necesita romper eso,
-- no es una migración: es un cambio de arquitectura y debe discutirse como tal.
--
-- Nótese lo que NO se crea aquí: la extensión pgcrypto. Al no quedar un solo
-- campo cifrado, la dependencia desaparece. Que vuelva a hacer falta es la
-- señal de que se coló un dato sensible.
--
-- REGLA ESTRUCTURAL: ninguna sentencia de este archivo puede nombrar una tabla
-- de la base de identidad. No hay claves foráneas, JOIN, dblink ni
-- postgres_fdw entre ambas bases.
--
-- Convención: golang-migrate (pares .up.sql / .down.sql, SQL plano). Los UUID
-- los provee la aplicación (Go, google/uuid v7).
--
-- Origen: consolida ../usbi/backend/migrations/0001 a 0012 en un baseline
-- único. Al consolidar desaparecen seis deudas heredadas: el CHECK NOT VALID
-- de 0002 validado hasta 0011, la reconstrucción de tablas de 0007 para
-- particionarlas, el parche de partición DEFAULT de 0008, los índices de clave
-- foránea olvidados hasta 0010, las cinco funciones PL/pgSQL que 0006 creó y
-- 0012 eliminó por muertas, y la convención de migración inconsistente.

-- ═══════════════════════════════════════════════════════════════════════════
-- 1. ANCLA DE IDENTIDAD ANÓNIMA
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
-- Ancla local de todas las claves foráneas de esta base.
--
-- En ../usbi había 17 claves foráneas apuntando a `users(id)`. PostgreSQL no
-- admite claves foráneas entre bases distintas, así que esas 17 tenían que
-- morir o cambiar de destino. Esta tabla es la que permite lo segundo: réplica
-- SIN NINGUNA COLUMNA IDENTIFICABLE de la identidad correspondiente.
--
-- accounts.id es igual a identities.id de la otra base, PERO SIN CLAVE FORÁNEA:
-- son bases distintas. La consistencia la sostiene la aplicación con un upsert
-- idempotente en login/refresh, no la base.
--
-- La fuente de verdad de role / status / is_adult es SIEMPRE la base de
-- identidad. Lo de aquí es una réplica de conveniencia: el panel admin filtra
-- por rol y el motor de progreso rechaza cuentas suspendidas sin tener que
-- consultar la otra base en cada request. La revocación inmediata no depende de
-- esta réplica: sigue funcionando por token_version contra la base de identidad.
CREATE TABLE accounts (
    id                 UUID        PRIMARY KEY,
    role               VARCHAR     NOT NULL
        CHECK (role IN ('player', 'admin', 'operator', 'director')),
    status             VARCHAR     NOT NULL
        CHECK (status IN ('active', 'suspended', 'pending_tutor_consent', 'deleted')),
    is_adult           BOOLEAN     NOT NULL,
    -- Alias visible, generado por el sistema. Ver alias_adjectives/alias_nouns.
    alias_adjective_id SMALLINT    NOT NULL REFERENCES alias_adjectives(id) ON DELETE RESTRICT,
    alias_noun_id      SMALLINT    NOT NULL REFERENCES alias_nouns(id)      ON DELETE RESTRICT,
    alias_number       SMALLINT    NOT NULL CHECK (alias_number BETWEEN 0 AND 999),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ
);

COMMENT ON TABLE accounts IS
    'Réplica no autoritativa y sin PII de la base de identidad. Ancla de todas '
    'las claves foráneas de esta base. Añadir aquí una columna que pueda '
    'contener texto escrito por una persona es incompatible con el diseño.';
COMMENT ON COLUMN accounts.id IS
    'UUID anónimo. Coincide con identities.id de la base de identidad, sin FK '
    '(bases distintas). No lleva ningún dato personal asociado en esta base.';

CREATE INDEX accounts_role_status_idx ON accounts (role, status);


-- Vista de conveniencia: compone el alias legible sin que ninguna tabla
-- almacene la cadena. El alias NO es un identificador: no es único y jamás debe
-- usarse como clave de búsqueda. La llave siempre es accounts.id.
CREATE VIEW account_aliases AS
SELECT a.id,
       n.word || ' ' || adj.word || ' ' || a.alias_number::text AS display_alias
FROM accounts a
JOIN alias_adjectives adj ON adj.id = a.alias_adjective_id
JOIN alias_nouns      n   ON n.id   = a.alias_noun_id;

-- ═══════════════════════════════════════════════════════════════════════════
-- 2. OPERACIÓN Y ARQUITECTURA OFFLINE
-- ═══════════════════════════════════════════════════════════════════════════

-- ── devices ──────────────────────────────────────────────────────────────────
-- CAMBIO FRENTE A ../usbi: allí existía `device_label VARCHAR NOT NULL`, texto
-- libre escrito por la persona usuaria. En la práctica eso se llena con "iPad
-- de Sofía". Aquí se sustituye por un vocabulario cerrado, que basta para
-- distinguir dispositivos en la pantalla de sesiones y no admite PII.
CREATE TABLE devices (
    id              UUID        PRIMARY KEY,
    user_id      UUID        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    device_kind     VARCHAR     NOT NULL
        CHECK (device_kind IN ('movil', 'tablet', 'laptop', 'escritorio', 'otro')),
    platform        VARCHAR     NOT NULL CHECK (platform IN ('web', 'tauri')),
    registered_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Bandera ARCO: cuando es TRUE, la siguiente respuesta de sync DEBE inyectar
    -- wipe_local_data = true para que el dispositivo borre su SQLite local.
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
    user_id         UUID        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
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
CREATE INDEX sync_events_device_id_idx      ON sync_events (device_id);

-- ═══════════════════════════════════════════════════════════════════════════
-- 3. CONTENIDO EDUCATIVO (borrado lógico)
-- ═══════════════════════════════════════════════════════════════════════════

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

CREATE INDEX levels_section_id_idx    ON levels (section_id);
CREATE INDEX levels_created_by_idx    ON levels (created_by_admin_id);
CREATE INDEX levels_deleted_by_idx    ON levels (deleted_by);

-- ═══════════════════════════════════════════════════════════════════════════
-- 4. PROGRESO Y RENDICIÓN DE CUENTAS
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE player_progress (
    user_id         UUID    NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    level_id           UUID    NOT NULL REFERENCES levels(id)   ON DELETE RESTRICT,
    best_score         INTEGER NOT NULL DEFAULT 0,
    xp_total_for_level INTEGER NOT NULL DEFAULT 0,
    attempts_count     INTEGER NOT NULL DEFAULT 0,
    first_completed_at TIMESTAMPTZ,
    last_completed_at  TIMESTAMPTZ,
    PRIMARY KEY (user_id, level_id)
);

CREATE INDEX player_progress_level_id_idx ON player_progress (level_id);

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
    user_id     UUID        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    level_id       UUID        NOT NULL REFERENCES levels(id)   ON DELETE RESTRICT,
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
    user_id    UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
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
    user_id UUID        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    badge_id   UUID        NOT NULL REFERENCES badges(id)   ON DELETE RESTRICT,
    earned_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, badge_id)
);

CREATE INDEX user_badges_badge_id_idx ON user_badges (badge_id);

-- ═══════════════════════════════════════════════════════════════════════════
-- 5. AUDITORÍA DE CONTENIDO Y SEGURIDAD
-- ═══════════════════════════════════════════════════════════════════════════

-- ── experience_history ───────────────────────────────────────────────────────
-- APPEND-ONLY. user_id pasa a NULL en la seudonimización ARCO: la fila del
-- libro mayor sobrevive (no repudio) pero deja de estar vinculada.
CREATE TABLE experience_history (
    id                  UUID        PRIMARY KEY,
    user_id          UUID        REFERENCES accounts(id) ON DELETE SET NULL,
    level_id            UUID        NOT NULL REFERENCES levels(id) ON DELETE RESTRICT,
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

CREATE INDEX experience_history_user_id_idx ON experience_history (user_id);
CREATE INDEX experience_history_level_id_idx   ON experience_history (level_id);
CREATE INDEX experience_history_sync_event_idx ON experience_history (sync_event_id);

-- ── admin_audit_log ──────────────────────────────────────────────────────────
-- APPEND-ONLY. En ../usbi esta tabla registraba TODA acción administrativa.
-- Aquí solo registra acciones de CONTENIDO (secciones, niveles, incidentes):
-- las de identidad (auth, ARCO, tutor, suspensión) van a identity_audit_log en
-- la base de identidad, porque su before_state / after_state contendría PII.
--
-- before_state / after_state NO DEBEN contener datos identificables. En esta
-- base eso se sostiene solo: no hay ninguna tabla de la que copiarlos.
CREATE TABLE admin_audit_log (
    id               UUID        PRIMARY KEY,
    actor_user_id UUID        REFERENCES accounts(id) ON DELETE SET NULL,
    action           VARCHAR     NOT NULL,
    entity_type      VARCHAR     NOT NULL,
    entity_id        UUID,
    before_state     JSONB,
    after_state      JSONB,
    ip_address       INET        NOT NULL,
    user_agent       TEXT        NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX admin_audit_log_actor_idx      ON admin_audit_log (actor_user_id);
CREATE INDEX admin_audit_log_created_at_idx ON admin_audit_log (created_at);

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

-- Estas tres columnas son el ÚNICO texto libre que queda en esta base, y lo
-- redacta un operador en pleno incidente — el momento exacto en que alguien
-- escribe "se filtró la cuenta de Juan Pérez". No se puede impedir con un
-- CHECK: queda como control documental, reforzado con validación en
-- internal/incidents. Referirse siempre a los titulares por su UUID.
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

    -- Única mutación permitida: anular la referencia a la cuenta durante la
    -- seudonimización ARCO, dejando el resto de la fila byte a byte idéntica.
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
    ELSIF TG_TABLE_NAME = 'admin_audit_log' THEN
        IF OLD.actor_user_id IS NOT NULL
           AND NEW.actor_user_id IS NULL
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

CREATE TRIGGER admin_audit_log_append_only_trg
BEFORE UPDATE OR DELETE ON admin_audit_log
FOR EACH ROW EXECUTE FUNCTION enforce_append_only_ledgers();

-- ═══════════════════════════════════════════════════════════════════════════
-- 6. DOCUMENTACIÓN DE COLUMNAS
-- ═══════════════════════════════════════════════════════════════════════════
-- El nombre de columna `user_id` se conserva en toda esta base, igual que en
-- ../usbi, para no invalidar la copia verbatim de la capa de repositorio en Go.
-- Lo que cambió no es el nombre sino el contenido: aquí un user_id es un UUID
-- anónimo, sin ningún dato personal asociado en esta base.
COMMENT ON COLUMN devices.user_id IS
    'UUID anónimo (accounts.id). Esta base no puede resolverlo a una persona.';
COMMENT ON COLUMN player_progress.user_id IS
    'UUID anónimo (accounts.id). Esta base no puede resolverlo a una persona.';
COMMENT ON COLUMN experience_history.user_id IS
    'UUID anónimo (accounts.id). NULL tras la seudonimización ARCO: la fila del '
    'libro mayor sobrevive para el no repudio, sin vínculo con la cuenta.';
