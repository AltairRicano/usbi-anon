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
-- Sustituir las contraseñas por valores reales tomados del gestor de secretos
-- (usbictl secrets init las genera). Nunca dejarlas escritas en este archivo
-- ni en el control de versiones.
--
-- Uso: el valor de cada -v va SIN comillas propias — este archivo usa la
-- sintaxis de psql :'variable' (quote-literal) más abajo, que ya envuelve el
-- valor en comillas SQL y escapa las internas. Agregar comillas aquí además
-- las deja como caracteres literales dentro de la contraseña real (bug
-- real, encontrado al ensayar backend/sql/00_init_cluster.sh, M3.10).
--   psql -U postgres -v app_password=… -v migrate_password=… \
--        -v moderador_password=… -v dbmaint_password=… \
--        -f 00_roles_unificado.sql
--
-- M3.3: la matriz de GRANT/REVOKE tabla por tabla que antes vivía en este
-- archivo se movió íntegra a la migración 0008_matriz_permisos.up.sql — este
-- archivo ya solo crea roles y la base, la única parte que de verdad exige
-- superusuario y no puede vivir en la cadena de golang-migrate (usbi_migrate
-- todavía no existe cuando esto corre). Mantener las dos cosas en el mismo
-- archivo invitaba a que quedaran desincronizadas; ahora hay una sola fuente
-- de verdad para los permisos.

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

-- usbi_migrate necesita ser miembro de usbi_moderador para poder ejecutar
-- ALTER FUNCTION ... OWNER TO usbi_moderador (migración 0004): Postgres exige
-- que el rol que reasigna la propiedad de un objeto sea miembro del nuevo
-- dueño, no solo que sea dueño del esquema. Descubierto al aplicar la cadena
-- de migraciones por primera vez con el rol real en vez del superusuario
-- postgres (M3.3) — sin este GRANT, `usbictl migrate up` falla en 0004 con
-- "must be member of role usbi_moderador".
GRANT usbi_moderador TO usbi_migrate;

CREATE DATABASE usbi_anon_db OWNER usbi_migrate;

-- La matriz de GRANT/REVOKE tabla por tabla (quién puede leer/escribir qué)
-- vive ahora en la migración 0008_matriz_permisos.up.sql, aplicada por
-- usbictl migrate con el rol usbi_migrate recién creado arriba — no en este
-- archivo. Ver esa migración para el detalle completo, tabla por tabla, con
-- la razón de cada GRANT y cada REVOKE.
