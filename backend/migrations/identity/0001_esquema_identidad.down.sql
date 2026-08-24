-- Migration: 0001_esquema_identidad.down.sql
-- Reversa del esquema de identidad. Orden inverso de dependencias (hijas → padres).

DROP TRIGGER IF EXISTS identity_audit_log_append_only_trg ON identity_audit_log;
DROP FUNCTION IF EXISTS enforce_append_only_ledger();

DROP TABLE IF EXISTS identity_audit_log;
DROP TABLE IF EXISTS arco_requests;
DROP TABLE IF EXISTS tutor_consent_tokens;
DROP TABLE IF EXISTS tutor_consents;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS identities;

-- pgcrypto NO se elimina: puede estar en uso por otras bases de la misma
-- instancia, y CREATE EXTENSION es idempotente al reaplicar la migración.
