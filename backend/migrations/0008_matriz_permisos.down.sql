-- Revierte la matriz de 0008 a un estado neutro: ningún privilegio de tabla
-- para usbi_app/usbi_moderador/usbi_dbmaint más allá de USAGE en el esquema.
-- No intenta reconstruir el GRANT ALL histórico de la instancia de
-- stress-test (ese nunca pasó por la cadena de migraciones) — el objetivo de
-- este down es dejar la base sin la matriz de 0008 aplicada, no imitar un
-- estado previo que nunca existió como migración.

REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM usbi_app;
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM usbi_moderador;
ALTER DEFAULT PRIVILEGES FOR ROLE usbi_migrate IN SCHEMA public
    REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM usbi_app;

REVOKE USAGE ON SCHEMA public FROM usbi_app;
REVOKE USAGE ON SCHEMA public FROM usbi_moderador;
REVOKE USAGE ON SCHEMA public FROM usbi_dbmaint;
GRANT ALL ON SCHEMA public TO PUBLIC;
