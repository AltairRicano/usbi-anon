-- Revertir exige que ninguna fila haya quedado con
-- verification_method IN ('online_verified', 'online_reported'): el trigger
-- experience_history_append_only_trg (0001) prohíbe reescribir esa columna en
-- filas existentes, así que este down solo puede aplicarse limpiamente sobre
-- una base sin intentos registrados bajo la migración 0007 (p. ej. antes del
-- primer despliegue). Es la misma garantía de append-only que protege xp_gained
-- funcionando como se espera, no un descuido de esta migración.

ALTER TABLE experience_history DROP CONSTRAINT experience_history_check;

ALTER TABLE experience_history ADD CONSTRAINT experience_history_check CHECK (
    (source = 'online'       AND verification_method = 'online_direct') OR
    (source = 'offline_sync' AND verification_method = 'hmac_offline')
);
