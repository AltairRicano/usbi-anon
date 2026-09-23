---
tipo: estado_proyecto
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-19
---

# Arquitectura — USBI-Anon

## Hardware

### VPS Hostinger (destino de producción)

**2026-09-14 →** Reemplaza al servidor de la UV (decisión D-01, ver
[[Decisiones#alojamiento_hostinger_d01|Decisiones]]). Tamaño del plan
contratado **sin confirmar todavía** — bloquea los valores de memoria de
Postgres y el límite real de almacenamiento que gobierna la rotación de
temporadas.

### Entorno de desarrollo (`mimir`, `192.168.1.210`)

**Hasta 2026-09-17:** contenedor único `usbi-anon` (Postgres 15 + backend Go +
frontend juntos), límite duro de 500 MB de RAM. Puertos: `8092` (frontend),
`8093` (backend), `5435` (Postgres).

**Desde 2026-09-17:** compose de 3 servicios (`db`/`api`/`web`) en
`/home/altair/usbi-anon/`, cada contenedor con 250 MB de RAM, código y datos
en bind mounts hacia `/mnt/wolf/codigo/usbi-anon/`. Ver el incidente de
colisión de nombre de proyecto en
[[Decisiones#docker_compose_nombre_proyecto_colision|Decisiones]].

## Tecnología

Ver [[Tecnologias|Tecnologías]] para el detalle de cada pieza. Resumen de la
pila: Go 1.22 + Postgres 15 (backend), React 18 + Vite 6 + TS + Tailwind v4 +
Phaser (frontend), `golang-migrate` + `usbictl` (operación), Docker Compose de
3 servicios (despliegue), Cloudflare Tunnel (exposición pública).

## Componentes

### Tres pools de Postgres con permisos reales

**2026-09-02/09-09.** `usbi_app` (self-service de jugador + gestión de cuentas
de admin), `usbi_moderador` (CRUD de contenido/administración:
sections/levels/badges/registration_*, lectura de audit_log/security_incidents),
`usbi_dbmaint` (solo partición, cero GRANT de tabla). Cruces de la matriz de
permisos resueltos con funciones `SECURITY DEFINER`
(`purge_account_quiz_answers`, `null_user_in_pseudonymizable_ledgers`,
`ensure_yearly_partition`), nunca ampliando GRANT directo. Detalle de la
decisión en [[Decisiones#roles_postgres_tres_pools|Decisiones]].

### Rotación de niveles y XP

**2026-08-25, corregido en la 10ª sesión.** El almacenamiento es finito; el
contenido se rota por temporadas. Retirar un nivel libera espacio pero nunca
debe quitarle a un jugador el XP ganado.

| Referencia a `levels(id)` | Dirección | Motivo |
|---|---|---|
| `level_attempts.level_id` | `CASCADE` | Se van con el nivel — ahorro de espacio |
| `player_progress.level_id` | `CASCADE` | Idem |
| `experience_history.level_id` | `SET NULL` (nullable) | El XP debe sobrevivir; trigger append-only nunca permite modificar `xp_gained` |
| `levels.section_id` | `RESTRICT` | Purgar una sección exige purgar antes sus niveles |

Más `account_retired_progress`, que acumula contadores de niveles
completados/intentos ya purgados. Retirar contenido son dos pasos separados:
archivar (reversible) y purgar (irreversible). **Estas tres direcciones no
deben uniformarse nunca.**

### Backend — paquetes `internal/` (Go)

Router propio en `internal/transport`. Paquetes principales: `auth` (registro
3 pasos, login, JWT), `quiz` (banco de preguntas, credenciales), `levels`
(`PlayerService`/`AdminService` separados), `badges`, `incidents`
(admin-only, sin `DELETE`), `auditlog` (admin-only, con auto-auditoría D4),
`interestlinks` (`PlayerService`/`AdminService`), `suggestions` (buzón
anónimo, sin auditoría por diseño), `devices` (alta automática + revocación
lógica), `sync` (historial offline por jugador), `dbmaint` (partición,
`SECURITY DEFINER`), `privacy` (cancelación de cuenta en una sola tx), `legal`
(avisos server-side desde M2). Patrón repetido: split
`PlayerService`/`AdminService` como dos `Service` separados (no uno con dos
`repo`), cada uno con su propio pool.

### Frontend — `features/` (React)

Registro/login (`features/auth`), maker local (`/maker`, `localStorage`),
administración de contenido (`/admin/content`), vista de jugador (`/`,
`/sections/:id`, `/levels/:id/play` con 7 plantillas Phaser,
`/local-levels/:id/play`, `/perfil`), `/settings` pública, `/admin/insignias`,
`/admin/seguridad` (bitácora + incidentes, 2 pestañas), `/admin/comunidad`
(enlaces + buzón, 2 pestañas), `/procesos-offline` (dispositivos +
sincronización), pestaña "Más" del dashboard (enlaces de interés como
carrusel + buzón de sugerencias), `/privacidad` pública. Componentes
compartidos de navegación: `HomeButton`, `LinkButton` (reemplazan el patrón
roto `<Button><Link>` anidado).

### Despliegue — compose de 3 servicios (M3)

`docker-compose.yml` versionado en el repo: `db` (Postgres 15,
`scram-sha-256`), `api` (backend Go, migra con `usbi_migrate` vía `usbictl
migrate`), `web` (frontend precompilado servido por nginx). Secretos
generados con `usbictl secrets init`; `usbictl pgconf render` genera
`pg_hba.conf`/overrides de Postgres. `DEPLOY.md` documenta la secuencia real
de despliegue, ensayada primero en `/home/altair/usbi-anon/` como entorno de
desarrollo real antes del VPS de Hostinger.
