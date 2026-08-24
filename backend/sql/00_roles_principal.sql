-- 00_roles_principal.sql — roles de la BASE PRINCIPAL (usbi_anon_db)
--
-- NO forma parte de la cadena de migraciones. Ver la nota equivalente en
-- 00_roles_identidad.sql: dos bases con el mismo usuario de conexión no aíslan
-- nada, y el aislamiento real es justamente lo que este proyecto promete.
--
-- Si la base principal y la de identidad terminan en INSTANCIAS DISTINTAS
-- (recomendación de plan/00_Plan_maestro.md §4.1), este archivo se aplica en la
-- instancia de la base principal y el otro en la de identidad. Si terminan en
-- la misma instancia, se aplican ambos ahí: los roles siguen siendo distintos y
-- ninguno tiene permisos sobre la base del otro.

CREATE ROLE usbi_main_app     LOGIN PASSWORD :'main_app_password';
CREATE ROLE usbi_main_migrate LOGIN PASSWORD :'main_migrate_password';

CREATE DATABASE usbi_anon_db OWNER usbi_main_migrate;

\connect usbi_anon_db

REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT  USAGE ON SCHEMA public TO usbi_main_app;

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO usbi_main_app;
ALTER DEFAULT PRIVILEGES FOR ROLE usbi_main_migrate IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO usbi_main_app;

-- Vocabulario del alias: solo lectura. Que la aplicación no pueda escribir aquí
-- es lo que convierte el alias generado en una garantía estructural — no basta
-- con que el código "no lo haga", la base no se lo permite.
REVOKE INSERT, UPDATE, DELETE ON alias_adjectives, alias_nouns FROM usbi_main_app;

-- Bitácoras append-only: sin DELETE, misma lógica de defensa en profundidad.
REVOKE DELETE ON admin_audit_log, experience_history FROM usbi_main_app;

-- Catálogo de insignias: lo gestiona una migración, no la aplicación.
REVOKE INSERT, UPDATE, DELETE ON badges FROM usbi_main_app;

GRANT SELECT ON account_aliases TO usbi_main_app;
