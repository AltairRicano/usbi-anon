-- Migration: 0001_esquema_unificado.down.sql
-- Reversa del esquema unificado. Orden inverso de dependencias (hijas → padres).
--
-- Las particiones de level_attempts y daily_streak caen junto con su tabla
-- padre, así que no se listan una por una.
--
-- No hay DROP EXTENSION: este esquema nunca crea pgcrypto (no queda un solo
-- campo cifrado en el sistema).

DROP TRIGGER IF EXISTS audit_log_append_only_trg          ON audit_log;
DROP TRIGGER IF EXISTS experience_history_append_only_trg ON experience_history;
DROP FUNCTION IF EXISTS enforce_append_only_ledgers();

DROP TABLE IF EXISTS security_incidents;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS experience_history;
DROP TABLE IF EXISTS arco_requests;
DROP TABLE IF EXISTS user_badges;
DROP TABLE IF EXISTS badges;
DROP TABLE IF EXISTS daily_streak;
DROP TABLE IF EXISTS level_attempts;
DROP TABLE IF EXISTS account_retired_progress;
DROP TABLE IF EXISTS player_progress;
DROP TABLE IF EXISTS levels;
DROP TABLE IF EXISTS sections;
DROP TABLE IF EXISTS sync_events;
DROP TABLE IF EXISTS devices;

DROP TABLE IF EXISTS account_quiz_answers;
DROP TABLE IF EXISTS registration_settings;
DROP TABLE IF EXISTS registration_questions;

DROP TABLE IF EXISTS refresh_tokens;
DROP VIEW  IF EXISTS account_aliases;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS alias_nouns;
DROP TABLE IF EXISTS alias_adjectives;
