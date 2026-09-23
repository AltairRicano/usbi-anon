---
tipo: estado_proyecto
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-19
---

# Tecnologías — USBI-Anon

## go & Go (backend)

Go 1.22 (pineado — `golang-migrate/v4` se fijó en `v4.18.1` porque v4.20+
exige go1.25). Módulo único en `backend/`, sin frameworks web — router propio
en `internal/transport`.

## postgresql & PostgreSQL 15

Una sola base `usbi_anon_db`, esquema unificado de 30+ tablas, particionado
por año donde aplica (`ensure_yearly_partition`). Tres roles con permisos de
tabla reales: `usbi_app`, `usbi_moderador`, `usbi_dbmaint` (más `usbi_migrate`
para `usbictl migrate`). `scram-sha-256` desde M3 (antes `trust`). Solo
extensión `plpgsql` — sin `pgcrypto`.

## golang_migrate & golang-migrate/v4

Migraciones embebidas (`go:embed`) ejecutadas por `usbictl migrate` como rol
`usbi_migrate`, reemplazando la aplicación manual de migraciones incrementales
que existía hasta M3. Versión pineada `v4.18.1` por la restricción de Go.

## react_vite & React 18 + Vite 6 + TypeScript

Frontend en `frontend/`. Tailwind v4. Sin Tauri — proyecto 100% web, sin plan
de empaquetado de escritorio (rama Tauri retirada del exportador del maker en
F10.8). Servido por nginx desde M3 (antes `vite preview`).

## phaser & Phaser (motores de minijuego)

`@usbi/engine` y `@usbi/schema` como npm workspaces (no pnpm). 7 plantillas de
nivel con contrato `GameResult` unificado desde M1. `npm run build` necesita
600–700 MB de pico — por encima del límite de memoria del contenedor de
desarrollo, de ahí que `dist/` se compile en el host.

## docker_compose & Docker Compose (3 servicios)

Desde M3: `db`/`api`/`web`, cada uno con límite de 250 MB de RAM, sustituyendo
la instancia de un solo contenedor (`usbi-anon`, Postgres+backend+frontend
juntos, 500 MB) que fue el entorno de referencia hasta 2026-09-17. Ver
[[Arquitectura|Arquitectura]] para el detalle de despliegue.

## usbictl & usbictl (binario propio)

`cmd/usbictl`: subcomandos `migrate`/`secrets`/`admin`/`doctor`/`pgconf`.
Sustituye módulos descartables previos (`cmd/hash_password`, aplicación manual
de GRANT). `doctor` comprueba también lo prohibido (que un rol NO pueda hacer
algo). Decisión D-04: despliegue por script/binario idempotente, no CI/CD.

## cloudflare_tunnel & Cloudflare Tunnel

Expone el frontend en `https://usbi.heimdall-lab.com/` sin abrir puertos en
el VPS directamente. `preview.allowedHosts` (o el equivalente en nginx desde
M3) debe incluir cualquier dominio nuevo del túnel.
