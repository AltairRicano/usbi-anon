-- 0007_verificacion_resultado_nivel: M1 Fase B (D-03) distingue, dentro de
-- source='online', entre XP que el servidor recalculó y verificó contra
-- level.content (verification_method = 'online_verified' — trivia y
-- fake_news por ahora, ver internal/levels/verify.go) y XP que el servidor
-- acepta sin poder verificarla (verification_method = 'online_reported' —
-- memory y snakes_ladders, decisión M1.4-B3: no son verificables sin
-- reproducir la partida completa). Sustituye al único valor anterior
-- 'online_direct', que no distinguía entre ambos casos.

ALTER TABLE experience_history DROP CONSTRAINT experience_history_check;

ALTER TABLE experience_history ADD CONSTRAINT experience_history_check CHECK (
    (source = 'online'       AND verification_method IN ('online_verified', 'online_reported')) OR
    (source = 'offline_sync' AND verification_method = 'hmac_offline')
);
