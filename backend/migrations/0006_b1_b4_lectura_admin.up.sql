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
