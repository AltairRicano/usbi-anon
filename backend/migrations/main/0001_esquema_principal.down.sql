-- Migration: 0001_esquema_principal.down.sql
-- Reversa del esquema principal. Orden inverso de dependencias (hijas → padres).
--
-- Las particiones de level_attempts y daily_streak caen junto con su tabla
-- padre, así que no se listan una por una.

DROP TRIGGER IF EXISTS admin_audit_log_append_only_trg    ON admin_audit_log;
DROP TRIGGER IF EXISTS experience_history_append_only_trg ON experience_history;
DROP FUNCTION IF EXISTS enforce_append_only_ledgers();

DROP TABLE IF EXISTS security_incidents;
DROP TABLE IF EXISTS admin_audit_log;
DROP TABLE IF EXISTS experience_history;
DROP TABLE IF EXISTS user_badges;
DROP TABLE IF EXISTS badges;
DROP TABLE IF EXISTS daily_streak;
DROP TABLE IF EXISTS level_attempts;
DROP TABLE IF EXISTS player_progress;
DROP TABLE IF EXISTS levels;
DROP TABLE IF EXISTS sections;
DROP TABLE IF EXISTS sync_events;
DROP TABLE IF EXISTS devices;

DROP VIEW  IF EXISTS account_aliases;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS alias_nouns;
DROP TABLE IF EXISTS alias_adjectives;
