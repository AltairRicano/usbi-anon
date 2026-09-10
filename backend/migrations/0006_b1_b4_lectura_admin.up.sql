-- 0006_b1_b4_lectura_admin: soporte de esquema para los 4 huecos de
-- backend que F1 dejó listados como "tocados solo a medias"
-- (estado_proyecto.md 2026-09-09, sección "B1–B4"). Cada bloque se aplica
-- en la fase correspondiente del plan (B3 → B4 → B1 → B2), documentado en
-- el propio bloque.

-- ── B3: CRUD de insignias ────────────────────────────────────────────────
-- badges.name no tenía restricción de unicidad; con el CRUD de admin recién
-- construido (internal/badges), dos insignias con el mismo nombre en el
-- panel serían un descuido de datos, no una situación válida (decisión D2,
-- estado_proyecto.md 2026-09-09).
ALTER TABLE badges ADD CONSTRAINT badges_name_key UNIQUE (name);

-- ── B4: historial de sincronización del jugador ─────────────────────────
-- Los índices existentes son (user_id, status) y (device_id) — ninguno
-- sirve para "mis últimas sincronizaciones" ordenado por received_at.
-- Sin este índice, la tabla que más crece del sistema se ordenaría con un
-- scan completo por usuario en GET /sync/events.
CREATE INDEX sync_events_user_received_idx
    ON sync_events (user_id, received_at DESC, id DESC);

-- ── B2: leer y editar incidentes de seguridad, nunca borrar ─────────────
-- El "jamás borrar desde la aplicación" que pidió el usuario ya estaba
-- sostenido por dos capas (ninguna ruta lo expone; ningún rol tiene el
-- privilegio DELETE). Esta es la tercera, a nivel de esquema: aunque un rol
-- futuro recibiera DELETE por error, o alguien ejecutara SQL directo con un
-- rol equivocado, la base misma lo rechaza. No reutiliza
-- enforce_append_only_ledgers() (esa función asume las columnas de
-- experience_history/audit_log, y security_incidents sí necesita permitir
-- UPDATE completo por decisión D1) — es una función dedicada, más simple:
-- prohíbe cualquier DELETE, sin excepción.
CREATE OR REPLACE FUNCTION forbid_security_incident_delete()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'security_incidents no admite DELETE; requiere acceso directo a Postgres'
        USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER security_incidents_no_delete_trg
BEFORE DELETE ON security_incidents
FOR EACH ROW EXECUTE FUNCTION forbid_security_incident_delete();

-- Índice de apoyo para la paginación de GET /admin/security-incidents
-- (detected_at DESC, id DESC) — la tabla no tenía ningún índice más allá de
-- la llave primaria.
CREATE INDEX security_incidents_detected_at_idx
    ON security_incidents (detected_at DESC, id DESC);
