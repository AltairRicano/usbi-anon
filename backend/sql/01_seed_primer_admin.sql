-- 01_seed_primer_admin.sql — INSERT manual del primer admin.
--
-- Bootstrap del primer admin (decisión 6 del rediseño de identidad, ver
-- plan/04_Rediseno_identidad_gustos.md §2): se inserta directamente en la
-- base vía este script, con el password ya hasheado por el binario
-- standalone cmd/hash_password. Desde la app, ese primer admin puede crear
-- a los siguientes con POST /api/v1/admin/accounts — este script solo hace
-- falta UNA VEZ por despliegue, para tener a alguien que pueda llamar a ese
-- endpoint.
--
-- PLANTILLA, no un script listo para correr tal cual: sustituir los
-- placeholders (marcados con «») antes de ejecutar. NO forma parte de la
-- cadena de golang-migrate — se aplica a mano, después de correr las
-- migraciones y 00_roles_unificado.sql, contra la base ya creada.
--
-- Paso 1 — generar id + password_hash + sello de privacidad. Corre esto en
-- una máquina con el módulo Go del backend (nunca pegues el password ni el
-- HMAC_SECRET en un shell compartido o en el historial de una terminal que
-- no controlas):
--
--   go run ./cmd/hash_password -seal -version "v1-staff" -secret "$HMAC_SECRET" "<password-elegido>"
--
-- Esto imprime cinco líneas: password_hash, id, privacy_notice_version,
-- privacy_notice_accepted_at y privacy_acceptance_hash. Cópialas en los
-- placeholders de abajo tal cual las imprimió — privacy_acceptance_hash ya
-- sale en formato \x... listo para pegar como literal bytea de Postgres.
--
-- Paso 2 — elegir un nickname que cumpla el CHECK de accounts
-- (^[a-z0-9]{6,20}$, ver 0001_esquema_unificado.up.sql) y sustituir «nickname».
--
-- Paso 3 — alias_adjective_id/alias_noun_id/alias_number son NOT NULL para
-- CUALQUIER cuenta, admin incluido, aunque a un admin no le importe ver un
-- alias amistoso. 1/1/0 son válidos (los vocabularios tienen ids 1..24,
-- alias_number acepta 0..999) — no hace falta sortearlos.
--
-- Paso 4 — ejecutar contra la base ya migrada:
--
--   psql "$DATABASE_URL" -f 01_seed_primer_admin.sql

INSERT INTO accounts (
    id,
    nickname,
    password_hash,
    is_adult,
    role,
    status,
    alias_adjective_id,
    alias_noun_id,
    alias_number,
    privacy_notice_version,
    privacy_notice_accepted_at,
    privacy_acceptance_hash,
    crypto_key_version
) VALUES (
    '«id-impreso-por-hash_password»',
    '«nickname»',
    '«password_hash-impreso-por-hash_password»',
    TRUE,
    'admin',
    'active',
    1,
    1,
    0,
    '«privacy_notice_version-impreso-por-hash_password»',
    '«privacy_notice_accepted_at-impreso-por-hash_password»',
    '«privacy_acceptance_hash-impreso-por-hash_password»',
    1
);
