-- 00_roles_unificado.sql — roles de la ÚNICA base del sistema (usbi_anon_db)
--
-- Reemplaza a 00_roles_identidad.sql y 00_roles_principal.sql, que existían
-- para el diseño de dos bases descartado en el rediseño de identidad.
--
-- NO forma parte de la cadena de migraciones: crear roles es una operación de
-- clúster y exige superusuario. Lo aplica una sola vez quien administra la
-- instancia, ANTES de correr golang-migrate.
--
-- POR QUÉ SIGUE EXISTIENDO ESTE ARCHIVO AUNQUE YA NO HAYA DOS BASES. Con una
-- sola base, la separación de roles deja de ser "aislar identidad de progreso"
-- y pasa a ser lo que siempre debió ser también: separar QUIÉN PUEDE CAMBIAR
-- EL ESQUEMA de quién solo puede leer y escribir datos, y quitarle al backend
-- permisos sobre lo que por diseño no debe poder tocar (el vocabulario del
-- alias, el catálogo de insignias, el borrado de las bitácoras). Esos permisos
-- ausentes son lo que convierte varias promesas del diseño en garantías
-- verificables en vez de convenciones de código.
--
-- Sustituir las contraseñas por valores reales tomados del gestor de secretos.
-- Nunca dejarlas escritas en este archivo ni en el control de versiones.
--
-- Uso:
--   psql -U postgres -v app_password="'…'" -v migrate_password="'…'" \
--        -f 00_roles_unificado.sql

-- Rol de aplicación: lo usa el backend en tiempo de ejecución.
CREATE ROLE usbi_app LOGIN PASSWORD :'app_password';

-- Rol de migración: solo aplica migraciones. El backend NO lo usa.
CREATE ROLE usbi_migrate LOGIN PASSWORD :'migrate_password';

CREATE DATABASE usbi_anon_db OWNER usbi_migrate;

\connect usbi_anon_db

REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT  USAGE ON SCHEMA public TO usbi_app;

-- Permisos del rol de aplicación sobre lo que ya existe y lo que se cree luego.
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO usbi_app;
ALTER DEFAULT PRIVILEGES FOR ROLE usbi_migrate IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO usbi_app;

-- Vocabulario del alias: solo lectura. Que la aplicación no pueda escribir aquí
-- es lo que convierte el alias generado en una garantía estructural — no basta
-- con que el código "no lo haga", la base no se lo permite.
REVOKE INSERT, UPDATE, DELETE ON alias_adjectives, alias_nouns FROM usbi_app;

-- Catálogo de insignias: lo gestiona una migración, no la aplicación.
REVOKE INSERT, UPDATE, DELETE ON badges FROM usbi_app;

-- Bitácoras append-only: sin DELETE. El trigger enforce_append_only_ledgers()
-- ya lo impide, pero esto lo bloquea una capa antes, y un trigger puede
-- desactivarse mientras que un permiso ausente deja rastro en la bitácora del
-- servidor.
REVOKE DELETE ON audit_log, experience_history FROM usbi_app;

-- Fila única de configuración: se actualiza, nunca se borra ni se duplica.
REVOKE INSERT, DELETE ON registration_settings FROM usbi_app;

GRANT SELECT ON account_aliases TO usbi_app;
