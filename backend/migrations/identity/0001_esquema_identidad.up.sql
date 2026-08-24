-- Migration: 0001_esquema_identidad.up.sql
-- Proyecto:  USBI-Anon — BASE DE DATOS DE IDENTIDAD (usbi_ident_db)
--
-- Esta base contiene el ÚNICO dato personal directo del sistema: el correo
-- electrónico (cifrado) y el hash de la contraseña, junto al UUID que
-- representa a esa persona en el resto de la plataforma.
--
-- Lo que esta base NO contiene, por diseño: progreso, XP, insignias, rachas,
-- intentos de nivel, contenido educativo ni bitácora de contenido. Todo eso
-- vive en la base principal (usbi_anon_db) indexado por UUID.
--
-- REGLA ESTRUCTURAL: ninguna sentencia de este archivo puede nombrar una tabla
-- de la base principal. No hay claves foráneas, JOIN, dblink ni postgres_fdw
-- entre ambas bases. Esa independencia es lo que permite alojarlas en
-- instancias distintas sin cambiar una línea de SQL.
--
-- Convención: golang-migrate (pares .up.sql / .down.sql, SQL plano, sin
-- anotaciones de goose). Los UUID los provee la aplicación (Go, google/uuid
-- v7); gen_random_uuid() no se usa como valor por defecto en ninguna PK.
--
-- Origen: consolida ../usbi/backend/migrations/0001, 0003, 0009 y 0011,
-- eliminando full_name y phone (ver plan/00_Plan_maestro.md §4).

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ── identities ───────────────────────────────────────────────────────────────
-- Reemplaza a la tabla `users` de ../usbi. Diferencias deliberadas:
--   · SE ELIMINA full_name  → PII que no participa en la autenticación.
--   · SE ELIMINA phone y phone_lookup_hash → se almacenaban cifrados con
--     índice ciego, pero el login nunca los consultaba y el formulario de
--     registro nunca los enviaba. Superficie de PII sin ninguna función.
CREATE TABLE identities (
    id                         UUID          PRIMARY KEY,
    -- Correo cifrado con pgcrypto (PGP_SYM_ENCRYPT). Nunca en claro.
    email                      BYTEA         NOT NULL,
    -- Índice ciego (HMAC) — es la vía de búsqueda en el login.
    email_lookup_hash          BYTEA         NOT NULL,
    password_hash              VARCHAR       NOT NULL, -- Argon2id, irreversible
    token_version              INTEGER       NOT NULL DEFAULT 1,
    is_adult                   BOOLEAN       NOT NULL,
    role                       VARCHAR       NOT NULL
        CHECK (role IN ('player', 'admin', 'operator', 'director')),
    privacy_notice_version     VARCHAR       NOT NULL,
    privacy_notice_accepted_at TIMESTAMPTZ   NOT NULL,
    privacy_acceptance_hash    BYTEA         NOT NULL, -- sello HMAC (no repudio)
    crypto_key_version         SMALLINT      NOT NULL,
    status                     VARCHAR       NOT NULL
        CHECK (status IN ('active', 'suspended', 'pending_tutor_consent', 'deleted')),
    -- Contador de intentos de transición a mayoría de edad (Ley 251, máx. 3).
    age_up_attempts            SMALLINT      NOT NULL DEFAULT 0,
    created_at                 TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at                 TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    last_login_at              TIMESTAMPTZ,
    deleted_at                 TIMESTAMPTZ,
    deletion_reason            VARCHAR
);

COMMENT ON TABLE identities IS
    'Único registro de datos personales directos del sistema. El resto de la '
    'plataforma solo conoce identities.id (UUID).';
COMMENT ON COLUMN identities.id IS
    'UUID v7 generado por la aplicación. Es el subject del JWT y el único valor '
    'que viaja a la base principal. Se conserva intacto tras la seudonimización '
    'ARCO, para que la saga de cancelación siempre pueda reintentarse.';

-- Correo único por cuenta activa (las canceladas conservan la fila seudonimizada).
CREATE UNIQUE INDEX identities_email_lookup_active_idx
    ON identities (email_lookup_hash)
    WHERE deleted_at IS NULL;

CREATE INDEX identities_status_idx ON identities (status);

-- Índices parciales para las tres rutinas de retención automática
-- (internal/maintenance). Sin ellos, cada corrida del planificador hace
-- recorrido secuencial completo sobre identities.
CREATE INDEX identities_pending_tutor_idx
    ON identities (created_at)
    WHERE status = 'pending_tutor_consent' AND deleted_at IS NULL;

CREATE INDEX identities_inactive_players_idx
    ON identities (COALESCE(last_login_at, created_at))
    WHERE role = 'player' AND status = 'active' AND deleted_at IS NULL;

CREATE INDEX identities_pending_cancel_idx
    ON identities (updated_at)
    WHERE role = 'player'
      AND status = 'suspended'
      AND deletion_reason = 'inactive_suspension_pending_cancel'
      AND deleted_at IS NULL;

-- ── refresh_tokens ───────────────────────────────────────────────────────────
-- Tokens opacos de sesión (7 días) con revocación del lado servidor.
CREATE TABLE refresh_tokens (
    id          UUID        PRIMARY KEY,
    user_id UUID        NOT NULL REFERENCES identities(id) ON DELETE CASCADE,
    token_hash  BYTEA       NOT NULL UNIQUE,
    issued_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ
);

CREATE INDEX refresh_tokens_user_active_idx
    ON refresh_tokens (user_id)
    WHERE revoked_at IS NULL;

-- Soporta la purga periódica de tokens vencidos o revocados hace más de 7 días.
CREATE INDEX refresh_tokens_expires_at_idx ON refresh_tokens (expires_at);

-- ── tutor_consents ───────────────────────────────────────────────────────────
-- Consentimiento parental YA VERIFICADO. Contiene datos personales de un
-- tercero (el tutor), así que su lugar natural es esta base y no la principal.
CREATE TABLE tutor_consents (
    id                     UUID        PRIMARY KEY,
    user_id            UUID        NOT NULL REFERENCES identities(id) ON DELETE CASCADE,
    tutor_name             BYTEA       NOT NULL, -- cifrado con pgcrypto
    tutor_email            BYTEA       NOT NULL, -- cifrado con pgcrypto
    privacy_notice_version VARCHAR     NOT NULL,
    accepted_at            TIMESTAMPTZ NOT NULL,
    -- IP y user-agent DEL CLIC de verificación: son la evidencia legal.
    acceptance_ip          INET        NOT NULL,
    acceptance_user_agent  TEXT        NOT NULL,
    consent_signature      BYTEA       NOT NULL, -- sello HMAC (evidencia legal)
    crypto_key_version     SMALLINT    NOT NULL,
    revoked_at             TIMESTAMPTZ
);

CREATE INDEX tutor_consents_user_id_idx ON tutor_consents (user_id);

-- ── tutor_consent_tokens ─────────────────────────────────────────────────────
-- Doble opt-in real: los datos del tutor se guardan como solicitud PENDIENTE
-- junto a un token de un solo uso (HMAC, 24 h) que se envía por correo. La
-- cuenta del menor se activa únicamente cuando el tutor hace clic.
CREATE TABLE tutor_consent_tokens (
    id                      UUID        PRIMARY KEY,
    user_id             UUID        NOT NULL REFERENCES identities(id) ON DELETE CASCADE,
    tutor_name              BYTEA       NOT NULL, -- cifrado con pgcrypto
    tutor_email             BYTEA       NOT NULL, -- cifrado con pgcrypto
    privacy_notice_version  VARCHAR     NOT NULL,
    -- HMAC-SHA256 del token entregado al tutor. El token en claro no se guarda,
    -- así que una fuga de esta base no permite forjar un enlace válido.
    token_hash              BYTEA       NOT NULL,
    -- IP / user-agent de QUIEN ENVIÓ el formulario: auditoría del inicio del
    -- trámite, NO la evidencia de consentimiento (esa es el clic, ver arriba).
    requested_ip            INET        NOT NULL,
    requested_user_agent    TEXT        NOT NULL,
    crypto_key_version      SMALLINT    NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at              TIMESTAMPTZ NOT NULL,
    -- Se escribe exactamente una vez, en el clic. NULL = sigue pendiente.
    verified_at             TIMESTAMPTZ,
    verification_ip         INET,
    verification_user_agent TEXT
);

CREATE UNIQUE INDEX tutor_consent_tokens_token_hash_idx
    ON tutor_consent_tokens (token_hash);

CREATE INDEX tutor_consent_tokens_user_id_idx
    ON tutor_consent_tokens (user_id);

-- Sirve tanto a la purga de vencidos como a "reemplazar el token pendiente
-- anterior de este usuario" en un reenvío.
CREATE INDEX tutor_consent_tokens_pending_idx
    ON tutor_consent_tokens (user_id, expires_at)
    WHERE verified_at IS NULL;

-- ── arco_requests ────────────────────────────────────────────────────────────
-- Solicitudes ARCO. Viven aquí porque pertenecen al titular del dato.
--
-- CAMBIO FRENTE A ../usbi: el vocabulario de `status` se amplía con dos estados
-- intermedios. Con dos bases separadas, resolver una cancelación deja de ser
-- una transacción única y pasa a ser una saga reanudable de tres pasos:
--
--   pending
--     └─► purging_main            (base principal: purga de progreso,
--     │                            SET NULL en bitácoras, wipe de dispositivos)
--     └─► identity_pseudonymized  (esta base: seudonimiza identidad y tutores,
--     │                            revoca refresh tokens, token_version + 1)
--     └─► resolved
--
-- La base principal va PRIMERO a propósito: si el proceso muere a mitad, queda
-- una cuenta viva con progreso ya purgado (recuperable y reintentable) en vez
-- de una cuenta bloqueada que ya no puede completar su propia cancelación.
-- Ver plan/02_Backend.md §5.
CREATE TABLE arco_requests (
    id               UUID        PRIMARY KEY,
    -- SET NULL: la solicitud sobrevive a la cancelación del titular, para
    -- conservar trazabilidad legal.
    user_id      UUID        REFERENCES identities(id) ON DELETE SET NULL,
    requester_type   VARCHAR     NOT NULL,
    request_type     VARCHAR     NOT NULL
        CHECK (request_type IN ('acceso', 'rectificacion', 'cancelacion', 'oposicion')),
    received_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at      TIMESTAMPTZ,
    status           VARCHAR     NOT NULL
        CHECK (status IN ('pending', 'purging_main', 'identity_pseudonymized',
                          'resolved', 'rejected')),
    handled_by       UUID        REFERENCES identities(id) ON DELETE SET NULL,
    response_summary TEXT,
    evidence_hash    BYTEA       NOT NULL -- sello criptográfico (no repudio)
);

CREATE INDEX arco_requests_user_id_idx ON arco_requests (user_id);
CREATE INDEX arco_requests_handled_by_idx  ON arco_requests (handled_by);

-- Cola de trabajo del panel ARCO y del job de reconciliación que rebarre
-- solicitudes atascadas en un estado intermedio.
CREATE INDEX arco_requests_open_idx
    ON arco_requests (received_at)
    WHERE status IN ('pending', 'purging_main', 'identity_pseudonymized');

-- ── identity_audit_log ───────────────────────────────────────────────────────
-- TABLA NUEVA, no existe en ../usbi.
--
-- Motivo: en ../usbi había una sola bitácora (`admin_audit_log`) que guardaba
-- before_state / after_state en JSONB para CUALQUIER acción administrativa,
-- incluidas las de identidad. Con dos bases, dejar esa bitácora entera en la
-- base principal filtraría PII a la base que promete no tenerla: el
-- before_state de "se editó el correo de la cuenta X" ES el correo.
--
-- Por eso la bitácora se parte en dos, con idéntico esquema y garantías:
--   · identity_audit_log (aquí)          → auth, ARCO, consentimiento de tutor,
--                                          suspensión, cancelación.
--   · admin_audit_log (base principal)   → contenido, secciones, niveles,
--                                          incidentes.
--
-- APPEND-ONLY: sin UPDATE ni DELETE, salvo el SET NULL del actor durante la
-- seudonimización ARCO. Lo garantiza el trigger de más abajo.
CREATE TABLE identity_audit_log (
    id              UUID        PRIMARY KEY,
    actor_user_id UUID      REFERENCES identities(id) ON DELETE SET NULL,
    action          VARCHAR     NOT NULL,
    entity_type     VARCHAR     NOT NULL,
    entity_id       UUID,
    before_state    JSONB,
    after_state     JSONB,
    ip_address      INET        NOT NULL,
    user_agent      TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX identity_audit_log_actor_idx ON identity_audit_log (actor_user_id);
CREATE INDEX identity_audit_log_created_at_idx ON identity_audit_log (created_at);

-- ── Invariante append-only ───────────────────────────────────────────────────
-- Se conserva de ../usbi/backend/migrations/0006. Es la única pieza de lógica
-- que 0012 dejó viva a propósito y con razón: es un invariante que debe
-- sostenerse sin importar qué código escriba en la tabla. El resto de las
-- funciones PL/pgSQL de 0006 estaban muertas y NO se recrean — Go es la única
-- fuente de verdad de XP y de la resolución ARCO.
CREATE OR REPLACE FUNCTION enforce_append_only_ledger()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION '% es append-only', TG_TABLE_NAME USING ERRCODE = '55000';
    END IF;

    -- Única mutación permitida: anular el actor durante la seudonimización
    -- ARCO, dejando el resto de la fila byte a byte idéntica.
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

    RAISE EXCEPTION '% es append-only; solo se permite el SET NULL de seudonimización',
        TG_TABLE_NAME USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER identity_audit_log_append_only_trg
BEFORE UPDATE OR DELETE ON identity_audit_log
FOR EACH ROW EXECUTE FUNCTION enforce_append_only_ledger();
