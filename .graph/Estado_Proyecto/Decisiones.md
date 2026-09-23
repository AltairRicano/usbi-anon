---
tipo: estado_proyecto
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-22
---

# Decisiones — USBI-Anon

Solo decisiones con peso real sobre arquitectura, producto o alcance legal.
Detalle narrativo completo en `estado_proyecto.md` (respaldo) y en
`plan/00_Plan_de_maduracion.md`.

## migracion_a_grafo_ia

**Fecha:** 2026-09-19

Se reemplazó el `estado_proyecto.md` suelto (bitácora cronológica de 2367
líneas) por `grafo_ia` (`.graph/Estado_Proyecto/`), siguiendo la regla 5 del
`CLAUDE.md` general del usuario, que declara esa directiva vieja reemplazada.
`estado_proyecto.md` se conserva en la raíz como respaldo consultable, no
como referencia activa. `.graph/` se agregó al `.gitignore` del proyecto.

**Carpetas que afecta:** ninguna (documento operativo, no código).

## base_de_datos_unica_sin_pii

**Fecha:** 2026-08-25

Se descartó el diseño inicial de dos bases de datos Postgres separadas
(identidad con correo + progreso), que había llegado a compilar y pasar
pruebas de integración. Decisión más radical: eliminar el correo electrónico
del sistema por completo. Registro por cuestionario de gustos no sensibles;
backend deriva nickname y password. Consentimiento de tutor por correo
eliminado — autorreporte de edad sin gate.

**Carpetas que afecta:** `backend/migrations`, `backend/internal/quiz`,
`backend/internal/auth`.

## alojamiento_hostinger_d01

**Fecha:** 2026-09-14

El alojamiento pasa a un VPS de Hostinger, revocando la premisa fundacional
de "no usar un tercero como host / servidor de la UV". Reaparece un
**encargado del tratamiento externo**: obliga a convenio de encargo y, según
el centro de datos, a cláusula de transferencia internacional. Ningún
documento debe seguir diciendo "operación directa de infraestructura
institucional". Ver [[Plan#m4_documentacion_tecnica_y_legal|Plan]] (M4.3·6).

**Carpetas que afecta:** `plan/`, documentación legal futura de M4.

## vps_hostinger_especificaciones_confirmadas

**Fecha:** 2026-09-22

El usuario confirmó las especificaciones reales del VPS de Hostinger
contratado: **1 GB de RAM, 1 vCPU, 20 GB de almacenamiento**. Cierra la
decisión abierta "tamaño del VPS contratado" listada en el `CLAUDE.md` del
proyecto — gobierna tanto el presupuesto de memoria de `docker-compose.yml`
(hoy 250 MB por contenedor, fijado antes de conocer esta cifra — **reconfirmar
que sigue siendo válido para 3 contenedores + Postgres bajo 1 GB total**,
dejando margen para el propio sistema operativo del VPS) como el límite real
de almacenamiento que gobierna la rotación de temporadas de niveles (ver
[[Decisiones#rotacion_niveles_tres_fk_distintas|rotación de niveles]]) — 20 GB
es sensiblemente menor a los ~20 GB de referencia histórica del servidor UV,
así que **no asumir que la cifra vieja de rotación por temporada sigue
sirviendo sin recalcularla contra este límite real**.

**Dominio confirmado:** `www.msusbicoatza.com`. Ese mismo dominio **ya aloja
otro sitio en producción** — "Museo AMI", HTML/PHP estático, en
`https://www.msusbicoatza.com/Museo%20AMI/index.php` — pero en **hosting
compartido separado del VPS**, no en el mismo VPS ni compitiendo por sus
recursos (confirmado por el usuario: "es compartido, pero solo es una página
estática, no usa demasiado"). Consecuencia para M3: el DNS raíz del dominio
ya apunta al hosting compartido, así que USBI-Anon en el VPS necesitará un
**subdominio propio** (p. ej. `usbi.msusbicoatza.com`) con su propio registro
DNS A hacia la IP del VPS — no puede vivir en la raíz del dominio sin
desplazar a Museo AMI. Falta decidir el subdominio exacto y actualizarlo en
`CORS_ALLOWED_ORIGIN` y `preview.allowedHosts`/config de nginx de M3.

**Carpetas que afecta:** `docker-compose.yml`, `DEPLOY.md`, `plan/` (M3, M4
legal — tamaño real de VPS también entra en la matriz de responsable/recursos
del encargado del tratamiento).

## rotacion_niveles_tres_fk_distintas

**Fecha:** 2026-08-25

El almacenamiento del servidor es finito y no se amplía pagando más; el
contenido se rota por temporadas. Retirar un nivel libera espacio pero nunca
debe quitarle a un jugador el XP ganado. Solución: tres direcciones de FK
deliberadamente distintas hacia `levels(id)` que **no deben uniformarse
nunca** — `level_attempts`/`player_progress` en `CASCADE`,
`experience_history` en `SET NULL`, `levels.section_id` en `RESTRICT`. Más
`account_retired_progress` para preservar contadores. Detalle completo en
[[Arquitectura#Rotación de niveles y XP|Arquitectura]].

**Carpetas que afecta:** `backend/migrations`, `backend/internal/levels`.

## roles_postgres_tres_pools

**Fecha:** 2026-09-02, completado en F3 (2026-09-09)

Separación de acceso a datos a nivel de Postgres, no solo en middleware Go.
Se descartó RLS por fila (sobre-ingeniería); elegido: permisos de tabla con
pools/DSN separados (`usbi_app`/`usbi_moderador`/`usbi_dbmaint`) — se
descartó `SET ROLE` sobre un pool compartido por riesgo de fuga de sesión.
Operaciones que cruzan la matriz de permisos se resuelven con funciones
`SECURITY DEFINER`, nunca ampliando GRANT (p. ej. `purge_account_quiz_answers`,
`null_user_in_pseudonymizable_ledgers`, `ensure_yearly_partition`).
`accounts`/`account_quiz_answers` se quedan deliberadamente en `usbi_app`
aunque sirvan rutas de admin — sin RLS, moverlas no da protección real, solo
trazabilidad forense.

**Carpetas que afecta:** `backend/internal/repository`, `backend/sql`,
`backend/main.go` (o su ubicación actual), `backend/internal/dbmaint`.

## eliminar_cuenta_borrado_duro_pendiente

**Fecha:** 2026-08-25/26 (F10.6)

Eliminar cuenta debía ser borrado duro real (no seudonimización) — sin PII,
seudonimizar no protege nada. **Sigue sin ejecutarse** en
`DELETE /api/v1/auth/me` (hoy seudonimización blanda): condicionado a fijar
antes la política de conservación (M4.3·9). Ver deuda conocida en
[[Estado|Estado]].

**Carpetas que afecta:** `backend/internal/auth`, `backend/internal/privacy`.

## incidentes_sin_delete_tres_capas

**Fecha:** 2026-09-10 (B2), decisión D1

Un incidente de seguridad puede editarse (`PATCH`, incluida la narrativa
sellada, resellando `evidence_hash`) pero **nunca borrarse** desde la
aplicación — decisión legal explícita (evidencia que la app no puede
destruir). Bloqueado en tres capas independientes: sin ruta `DELETE`, sin
privilegio de ningún rol, y trigger `BEFORE DELETE` en el esquema que
sobrevive incluso a un `DELETE` directo como superusuario.

**Carpetas que afecta:** `backend/internal/incidents`,
`backend/migrations`.

## suggestions_sin_auditoria

**Fecha:** 2026-09-09 (F4)

El buzón de sugerencias (`suggestions`) es deliberadamente anónimo: sin
`account_id` ni FK, y `Submit` **no llama a `audit.Log`** — una entrada de
auditoría con `actor_account_id` + timestamp casi idéntico reconstruiría el
vínculo cuenta↔sugerencia que la tabla busca impedir. Mismo espíritu que las
funciones `SECURITY DEFINER`: nunca dejar que el jugador "vea" una operación
que sí ejecuta por debajo. `AdminService` audita solo el ID al borrar, nunca
el contenido.

**Carpetas que afecta:** `backend/internal/suggestions`.

## docker_compose_nombre_proyecto_colision

**Fecha:** 2026-09-17

Incidente real: `docker compose down --remove-orphans` corrido desde
`/mnt/wolf/codigo/usbi-anon` (este repo) eliminó el contenedor de desarrollo
real en `/home/altair/usbi-anon/` porque Docker Compose infiere el nombre de
proyecto del nombre del directorio y ambos comparten `usbi-anon`. No se
perdieron datos (bind mount), pero sí un parche de `entrypoint.sh` que solo
existía en la capa de escritura del contenedor viejo — recuperado desde
backup. **Regla operativa:** cualquier `docker compose` corrido desde la raíz
de este repo debe usar `-p <nombre-distinto>` explícito y nunca
`--remove-orphans` mientras coexistan ambos directorios. Deja de ser riesgo
una vez M3 reemplazó por completo la instancia vieja (ya ejecutado, ver
[[Plan#m3_despliegue_hostinger|Plan]]).

**Carpetas que afecta:** ninguna (infraestructura fuera del repo).

## decisiones_abiertas_bloqueantes

**Fecha:** 2026-09-14, sin resolver a 2026-09-19

Cuatro decisiones bloquean trabajo concreto de M3/M4:

| Decisión pendiente | Qué bloquea |
|---|---|
| Nombre y dominio definitivos del producto | TLS, `CORS_ALLOWED_ORIGIN`, nombre del responsable en los avisos |
| Qué área de la UV figura como responsable del tratamiento | Redacción de los tres avisos de privacidad (M4.3) |
| Tamaño del VPS contratado en Hostinger | Valores de memoria de Postgres y límite real de almacenamiento que gobierna la rotación de temporadas |
| Ubicación del centro de datos del proveedor | Cláusula de transferencia internacional en el convenio de encargo (M4.3·6) |

**Carpetas que afecta:** `plan/`, M3/M4 en general.
