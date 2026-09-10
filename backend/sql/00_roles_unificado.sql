-- 00_roles_unificado.sql — roles de la ÚNICA base del sistema (usbi_anon_db)
--
-- Reemplaza a 00_roles_identidad.sql y 00_roles_principal.sql, que existían
-- para el diseño de dos bases descartado en el rediseño de identidad.
--
-- NO forma parte de la cadena de migraciones: crear roles es una operación de
-- clúster y exige superusuario. Lo aplica una sola vez quien administra la
-- instancia, ANTES de correr golang-migrate.
--
-- POR QUÉ SIGUE EXISTIENDO ESTE ARCHIVO AUNQUE YA NO HAYA DOS BASES. Con una
-- sola base, la separación de roles deja de ser "aislar identidad de progreso"
-- y pasa a ser lo que siempre debió ser también: separar QUIÉN PUEDE CAMBIAR
-- EL ESQUEMA de quién solo puede leer y escribir datos, y quitarle al backend
-- permisos sobre lo que por diseño no debe poder tocar (el vocabulario del
-- alias, el catálogo de insignias, el borrado de las bitácoras). Esos permisos
-- ausentes son lo que convierte varias promesas del diseño en garantías
-- verificables en vez de convenciones de código.
--
-- usbi_app vs usbi_moderador (2026-09-02): dentro de "quién puede leer y
-- escribir datos" hay dos niveles de confianza, no uno. usbi_app es la
-- credencial de todo lo NO privilegiado — login, registro, autoservicio de
-- jugador — y usbi_moderador es una credencial nueva, exclusiva de las
-- operaciones que YA estaban detrás de un guard de rol admin en Go
-- (internal/levels.canManageContent, internal/quiz.canManageQuizBank, etc.).
-- La app sigue decidiendo con ese guard quién puede llamar a esas funciones;
-- lo que gana esta separación es que, si algún día un bug o una inyección
-- corre bajo la credencial de usbi_app, no puede escribir en el contenido ni
-- en la configuración del sistema aunque el chequeo de rol en Go falle —
-- la base de datos ya no se lo permite, no solo el código.
--
-- No cubre `accounts` ni `account_quiz_answers` (lectura): mover esas
-- operaciones de administración de cuentas a usbi_moderador no añadía
-- ninguna restricción real (accounts no tiene Row-Level Security, así que
-- ambas credenciales podrían tocar cualquier fila igual) — decisión
-- explícita, ver estado_proyecto.md 2026-09-02.
--
-- `audit_log` (lectura) SÍ es distinto: en 2026-09-02 no existía ningún
-- endpoint de lectura (hueco de producto conocido); B1 (2026-09-09) lo
-- construyó sobre usbi_moderador —no usbi_app, que sigue sin ningún SELECT
-- aquí— porque, a diferencia de `accounts`, aislar la lectura de la
-- bitácora en la credencial no-privilegiada sí es una restricción real: si
-- `usbi_app` se compromete, no puede releer la evidencia forense.
--
-- Sustituir las contraseñas por valores reales tomados del gestor de secretos.
-- Nunca dejarlas escritas en este archivo ni en el control de versiones.
--
-- Uso:
--   psql -U postgres -v app_password="'…'" -v migrate_password="'…'" \
--        -v moderador_password="'…'" -f 00_roles_unificado.sql

-- Rol de aplicación: lo usa el backend en tiempo de ejecución para todo lo
-- no privilegiado (login, registro, autoservicio de jugador).
CREATE ROLE usbi_app LOGIN PASSWORD :'app_password';

-- Rol de moderación: lo usa el backend SOLO para las operaciones ya
-- restringidas a rol admin en Go (gestión de contenido, banco de preguntas,
-- catálogo de insignias, incidentes de seguridad). Ver bloque de permisos
-- más abajo para el detalle tabla por tabla.
CREATE ROLE usbi_moderador LOGIN PASSWORD :'moderador_password';

-- Rol de migración: solo aplica migraciones. El backend NO lo usa.
CREATE ROLE usbi_migrate LOGIN PASSWORD :'migrate_password';

-- Rol de mantenimiento de particiones (F3, 2026-09-09): lo usa
-- internal/dbmaint, el único componente del backend que hace DDL en tiempo
-- de ejecución (CREATE TABLE ... PARTITION OF sobre level_attempts/
-- daily_streak, ver plan/01_Base_de_datos.md). Ni usbi_app ni usbi_moderador
-- pueden tener este permiso — DDL no es DML, y ninguno de los dos debe poder
-- alterar el esquema aunque el binario que los usa esté comprometido. Este
-- rol NO recibe ningún GRANT de tabla: su único permiso es EXECUTE sobre
-- ensure_yearly_partition (migración 0005), una función SECURITY DEFINER
-- propiedad de quien aplicó las migraciones (usbi_migrate en el diseño
-- documentado; en el contenedor de stress-test usbi-anon, hoy es postgres,
-- porque el entrypoint aplica migrations/*.up.sql como superusuario) — el
-- mismo patrón de usbi_moderador ejecutando purge_account_quiz_answers sin
-- tener DELETE directo (migración 0004).
CREATE ROLE usbi_dbmaint LOGIN PASSWORD :'dbmaint_password';

CREATE DATABASE usbi_anon_db OWNER usbi_migrate;

\connect usbi_anon_db

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

-- Incidentes de seguridad: solo inserta (POST /admin/security-incidents).
-- Sin SELECT todavía — mismo caso que audit_log, sin endpoint de lectura
-- (decisión "déjalo" registrada en estado_proyecto.md).
GRANT INSERT ON security_incidents TO usbi_moderador;

-- Enlaces de interés y sus categorías (2026-09-08): CRUD completo — es
-- contenido editorial igual que sections/levels/badges.
GRANT SELECT, INSERT, UPDATE, DELETE ON interest_link_categories, interest_links TO usbi_moderador;

-- Buzón de sugerencias (2026-09-08): SELECT y DELETE únicamente. Nunca
-- INSERT (solo un jugador manda sugerencias, vía usbi_app) ni UPDATE (una
-- sugerencia no se edita, se lee o se borra).
GRANT SELECT, DELETE ON suggestions TO usbi_moderador;
