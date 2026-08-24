-- 00_roles_identidad.sql — roles de la BASE DE IDENTIDAD (usbi_ident_db)
--
-- NO forma parte de la cadena de migraciones: crear roles es una operación de
-- clúster y exige superusuario. Lo aplica una sola vez quien administra la
-- instancia, ANTES de correr golang-migrate.
--
-- POR QUÉ EXISTE ESTE ARCHIVO: dos bases de datos con el mismo usuario de
-- conexión no aíslan nada. Si el backend habla con ambas usando la misma
-- credencial, la separación es nominal y una inyección o un bug de cableado
-- puede leer identidad desde el camino de progreso. Que el proceso Go tenga
-- las dos credenciales es inevitable; que la base las trate como la misma, no.
--
-- Sustituir las contraseñas por valores reales tomados del gestor de secretos.
-- Nunca dejarlas escritas en este archivo ni en el control de versiones.

-- Rol de aplicación: lo usa el backend en tiempo de ejecución.
CREATE ROLE usbi_ident_app LOGIN PASSWORD :'ident_app_password';

-- Rol de migración: solo aplica migraciones. El backend NO lo usa.
CREATE ROLE usbi_ident_migrate LOGIN PASSWORD :'ident_migrate_password';

CREATE DATABASE usbi_ident_db OWNER usbi_ident_migrate;

\connect usbi_ident_db

REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT  USAGE ON SCHEMA public TO usbi_ident_app;

-- Permisos del rol de aplicación sobre lo que ya existe y lo que se cree luego.
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO usbi_ident_app;
ALTER DEFAULT PRIVILEGES FOR ROLE usbi_ident_migrate IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO usbi_ident_app;

-- La bitácora es append-only: el rol de aplicación no puede borrarla. El
-- trigger enforce_append_only_ledger() ya lo impide, pero esto lo bloquea una
-- capa antes, y un trigger puede desactivarse mientras que un permiso ausente
-- deja rastro en la bitácora del servidor.
REVOKE DELETE ON identity_audit_log FROM usbi_ident_app;
ALTER DEFAULT PRIVILEGES FOR ROLE usbi_ident_migrate IN SCHEMA public
    REVOKE DELETE ON TABLES FROM usbi_ident_app;
GRANT DELETE ON identities, refresh_tokens, tutor_consents, tutor_consent_tokens
    TO usbi_ident_app;

-- El rol de aplicación NO tiene ningún permiso sobre la base principal:
-- simplemente no existe como rol ahí.
