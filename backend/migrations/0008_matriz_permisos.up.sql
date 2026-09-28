-- 0008_matriz_permisos: la matriz de GRANT/REVOKE tabla por tabla, movida
-- aquí desde backend/sql/00_roles_unificado.sql (M3.3). Antes vivía en un
-- script suelto que solo el entrypoint de la instancia de stress-test
-- reaplicaba a mano en cada arranque, corriendo como el superusuario
-- postgres; a partir de M3 la aplica usbictl migrate con el rol usbi_migrate
-- (dueño del esquema, ver 00_roles_unificado.sql), como cualquier otra
-- migración — una sola fuente de verdad, versionada y con historial de
-- cambios real en vez de un script que se reaplica por completo cada vez.
--
-- Contenido idéntico al que tenía 00_roles_unificado.sql — ver ese archivo
-- (histórico en git) para el razonamiento original de cada línea; los
-- comentarios se conservan aquí tal cual para no perder ese contexto.

REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT  USAGE ON SCHEMA public TO usbi_app;
GRANT  USAGE ON SCHEMA public TO usbi_moderador;
-- Sin GRANT SELECT/INSERT/UPDATE/DELETE en ninguna tabla: usbi_dbmaint solo
-- necesita USAGE para poder ejecutar ensure_yearly_partition, cuyo cuerpo
-- corre con los privilegios de su dueño, no los de quien la llama.
GRANT  USAGE ON SCHEMA public TO usbi_dbmaint;

-- ── usbi_app ─────────────────────────────────────────────────────────────

-- Permisos del rol de aplicación sobre lo que ya existe y lo que se cree luego.
-- Las tablas que abajo se le retiran (sections, levels, badges,
-- registration_questions, registration_settings) parten de aquí con el
-- mismo CRUD completo que el resto y se recortan explícitamente después —
-- así una tabla nueva creada por una migración futura queda accesible por
-- defecto y hay que decidir activamente si se le recorta, en vez de que el
-- olvido la deje sin abrir.
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO usbi_app;
ALTER DEFAULT PRIVILEGES FOR ROLE usbi_migrate IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO usbi_app;

-- Vocabulario del alias: solo lectura. Que la aplicación no pueda escribir aquí
-- es lo que convierte el alias generado en una garantía estructural — no basta
-- con que el código "no lo haga", la base no se lo permite.
REVOKE INSERT, UPDATE, DELETE ON alias_adjectives, alias_nouns FROM usbi_app;

-- Secciones y niveles: jugador es SOLO LECTURA. Crear/editar/publicar/
-- archivar/purgar contenido corre con usbi_moderador (internal/levels, los
-- métodos ya separados de CreateLevel/PublishLevel/ArchiveLevel/PurgeLevel/…
-- frente a ListLevels/GetLevel/CompleteLevel).
REVOKE INSERT, UPDATE, DELETE ON sections, levels FROM usbi_app;

-- Catálogo de insignias: gestionado por moderador (decisión de producto
-- 2026-09-02 — antes lo gestionaba solo una migración; se abrió para que el
-- equipo de USBI pueda agregar insignias sin depender de un ingeniero). El
-- CRUD en Go todavía no existe (no hay rutas /admin/badges) — este GRANT
-- deja el terreno listo para cuando se construya.
REVOKE INSERT, UPDATE, DELETE ON badges FROM usbi_app;

-- Banco de preguntas de registro y su configuración: jugador solo lee
-- (registro, muestreo aleatorio); administrar el banco corre con
-- usbi_moderador (internal/quiz, canManageQuizBank).
REVOKE INSERT, UPDATE, DELETE ON registration_questions FROM usbi_app;
REVOKE INSERT, UPDATE, DELETE ON registration_settings FROM usbi_app;

-- Respuestas del cuestionario de registro: jugador solo inserta las suyas
-- durante el registro, nunca las lee de vuelta — leerlas (para que un admin
-- compare y decida si resetea una cuenta olvidada) es exclusivo de
-- usbi_moderador.
REVOKE SELECT, UPDATE, DELETE ON account_quiz_answers FROM usbi_app;

-- Bitácora unificada: jugador solo inserta (account.register, account.age_up
-- en internal/auth) — nunca lee. Sin DELETE tampoco: el trigger
-- enforce_append_only_ledgers() ya lo impide, pero esto lo bloquea una capa
-- antes, y un trigger puede desactivarse mientras que un permiso ausente deja
-- rastro en la bitácora del servidor.
--
-- Sin UPDATE tampoco (F3, 2026-09-09): el único UPDATE que usbi_app llegó a
-- necesitar aquí (poner a NULL actor_account_id al cancelar cuenta) ahora
-- pasa por la función SECURITY DEFINER null_user_in_pseudonymizable_ledgers
-- (migración 0004), propiedad de usbi_moderador — usbi_app solo la ejecuta,
-- nunca hace el UPDATE por su cuenta. Antes de esta revocación, usbi_app
-- conservaba el UPDATE heredado del GRANT ALL inicial sin que ningún REVOKE
-- lo cerrara; era un permiso de sobra, no usado por ningún camino de código,
-- pero abierto igual.
REVOKE SELECT, UPDATE, DELETE ON audit_log FROM usbi_app;

-- Libro mayor de XP: sin DELETE, mismo razonamiento que audit_log.
REVOKE DELETE ON experience_history FROM usbi_app;

-- Contadores de niveles ya retirados: jugador solo lee su propio total
-- (GetProfileProgress combina progreso vivo + retirado). Escribir ahí ocurre
-- solo al purgar un nivel (AccumulateRetiredProgressForLevel), exclusivo de
-- usbi_moderador. Sin DELETE para nadie: los contadores solo crecen: un bug
-- en la purga no puede borrarle a un jugador el progreso acumulado de
-- temporadas anteriores. Las filas se van solas al cancelar la cuenta, por
-- el CASCADE desde accounts, que no necesita este permiso.
REVOKE INSERT, UPDATE, DELETE ON account_retired_progress FROM usbi_app;

-- security_incidents nunca lo toca usbi_app: ningún flujo de jugador lo usa,
-- solo POST /admin/security-incidents (usbi_moderador, ver abajo).
REVOKE ALL ON security_incidents FROM usbi_app;

-- Enlaces de interés y sus categorías (sección "Más" del frontend, 2026-09-08):
-- jugador es SOLO LECTURA. CRUD completo corre con usbi_moderador.
REVOKE INSERT, UPDATE, DELETE ON interest_link_categories, interest_links FROM usbi_app;

-- Buzón de sugerencias (2026-09-08): jugador SOLO INSERTA — es anónimo por
-- diseño (sin account_id), así que ni siquiera tiene sentido que el propio
-- autor pueda releer su fila después de mandarla. Leer y borrar es exclusivo
-- de usbi_moderador; nadie tiene UPDATE, una sugerencia no se edita.
REVOKE SELECT, UPDATE, DELETE ON suggestions FROM usbi_app;

GRANT SELECT ON account_aliases TO usbi_app;

-- ── usbi_moderador ───────────────────────────────────────────────────────
-- Sin ALTER DEFAULT PRIVILEGES: a diferencia de usbi_app (acceso amplio por
-- defecto, recortado tabla por tabla), el acceso de usbi_moderador es
-- selectivo por diseño — una tabla nueva NO le llega automáticamente, hay
-- que concedérsela explícitamente aquí si corresponde.

-- Contenido: CRUD completo (internal/levels — todos los métodos detrás de
-- canManageContent/canArchiveContent).
GRANT SELECT, INSERT, UPDATE, DELETE ON sections, levels TO usbi_moderador;

-- Catálogo de insignias: CRUD completo (ver nota junto al REVOKE de usbi_app
-- arriba — endpoint Go construido en B3, estado_proyecto.md 2026-09-09).
GRANT SELECT, INSERT, UPDATE, DELETE ON badges TO usbi_moderador;

-- Solo lectura sobre las filas de titulares (B3): DeleteBadge pregunta
-- primero cuántos jugadores tienen una insignia antes de borrarla —mismo
-- guard que CountInterestLinksByCategory— y sin este GRANT ni siquiera
-- puede hacer esa pregunta. Sin INSERT/UPDATE/DELETE: conceder o revocar una
-- insignia ganada sigue siendo exclusivo de usbi_app
-- (repository.AwardEligibleBadges), nunca del panel de administración.
GRANT SELECT ON user_badges TO usbi_moderador;

-- Banco de preguntas de registro y su configuración: CRUD completo
-- (internal/quiz, canManageQuizBank).
GRANT SELECT, INSERT, UPDATE, DELETE ON registration_questions TO usbi_moderador;
GRANT SELECT, INSERT, UPDATE, DELETE ON registration_settings TO usbi_moderador;

-- Respuestas del cuestionario: solo lectura (GetAccountQuizAnswers). Nunca
-- inserta — eso solo pasa durante el registro, bajo usbi_app.
--
-- DELETE (F3, 2026-09-09): agregado para purge_account_quiz_answers
-- (migración 0004), la función SECURITY DEFINER que usbi_app ejecuta al
-- cancelar una cuenta — la función corre con los privilegios de su dueño
-- (usbi_moderador), así que el dueño necesita el permiso subyacente aunque
-- solo se ejerza a través de la función, nunca por una sentencia DELETE
-- suelta de usbi_moderador.
GRANT SELECT, DELETE ON account_quiz_answers TO usbi_moderador;

-- Contadores de niveles retirados: inserta/actualiza durante la purga
-- (AccumulateRetiredProgressForLevel, upsert). Sin SELECT ni DELETE — leer
-- el total combinado sigue siendo cosa del jugador vía usbi_app.
GRANT INSERT, UPDATE ON account_retired_progress TO usbi_moderador;

-- Bitácora unificada: inserta las auditorías de contenido/incidentes
-- (internal/levels.logAdminAudit, internal/incidents).
GRANT INSERT ON audit_log TO usbi_moderador;

-- SELECT de tabla completa (B1, estado_proyecto.md 2026-09-09): endpoint de
-- lectura construido — GET /admin/audit-log (internal/auditlog). Cierra el
-- hallazgo de producto abierto desde 2026-09-02 ("un admin puede escribir
-- pero no releer sin acceso directo a Postgres"). Cada lectura exitosa
-- vuelve a insertar en audit_log (audit_log.read, con los filtros usados,
-- nunca los resultados — decisión D4), así que este mismo GRANT es también
-- lo que le permite auditarse a sí misma.
GRANT SELECT ON audit_log TO usbi_moderador;

-- Column-level (F3, 2026-09-09), no tabla completa: null_user_in_pseudonymizable_ledgers
-- (migración 0004) hace `UPDATE audit_log SET actor_account_id = NULL WHERE
-- actor_account_id = $1` — necesita leer y escribir esa sola columna para
-- filtrar y anular, sin que eso abra lectura del resto de la bitácora
-- (before_state/after_state pueden llevar datos sensibles de auditoría).
GRANT SELECT (actor_account_id), UPDATE (actor_account_id) ON audit_log TO usbi_moderador;

-- Libro mayor de XP (F3, 2026-09-09): mismo caso que audit_log arriba —
-- null_user_in_pseudonymizable_ledgers también anula experience_history.user_id.
-- Column-level, sin abrir xp_gained/source/verification_method.
GRANT SELECT (user_id), UPDATE (user_id) ON experience_history TO usbi_moderador;

-- Incidentes de seguridad (B2, estado_proyecto.md 2026-09-09): INSERT
-- (POST /admin/security-incidents) + SELECT/UPDATE de tabla completa
-- (GET/PATCH) — decisión D1: el PATCH puede corregir también la narrativa
-- sellada, resellando evidence_hash, así que el UPDATE es de tabla completa
-- y no column-level como en audit_log/experience_history. NUNCA DELETE:
-- ningún rol lo tiene, y la migración 0006 además lo prohíbe con un trigger
-- BEFORE DELETE — tres capas independientes.
GRANT SELECT, INSERT, UPDATE ON security_incidents TO usbi_moderador;

-- Enlaces de interés y sus categorías (2026-09-08): CRUD completo — es
-- contenido editorial igual que sections/levels/badges.
GRANT SELECT, INSERT, UPDATE, DELETE ON interest_link_categories, interest_links TO usbi_moderador;

-- Buzón de sugerencias (2026-09-08): SELECT y DELETE únicamente. Nunca
-- INSERT (solo un jugador manda sugerencias, vía usbi_app) ni UPDATE (una
-- sugerencia no se edita, se lee o se borra).
GRANT SELECT, DELETE ON suggestions TO usbi_moderador;
