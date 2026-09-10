DROP INDEX IF EXISTS security_incidents_detected_at_idx;
DROP TRIGGER IF EXISTS security_incidents_no_delete_trg ON security_incidents;
DROP FUNCTION IF EXISTS forbid_security_incident_delete();
DROP INDEX IF EXISTS sync_events_user_received_idx;
ALTER TABLE badges DROP CONSTRAINT IF EXISTS badges_name_key;
