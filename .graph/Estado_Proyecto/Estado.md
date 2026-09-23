---
tipo: estado_proyecto
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-22
---

# Estado del proyecto — USBI-Anon

Migrado desde el `estado_proyecto.md` suelto (bitácora cronológica, 2367 líneas,
2026-08-24 → 2026-09-17) a `grafo_ia` el 2026-09-19, según la regla 5 del
`CLAUDE.md` general. El archivo original se conserva en la raíz del repo como
respaldo consultable; este documento es la fotografía del ahora, no la
cronología — para el detalle sesión por sesión, leer `estado_proyecto.md` o
`plan/_historico/`.

## Hecho

- **Identidad sin PII**: cuestionario de gustos → nickname/password derivados,
  `display_alias` aleatorio independiente, sin correo ni teléfono en el sistema.
- **Esquema unificado** (`0001`–`0008`) en una sola base `usbi_anon_db`, con
  rotación de niveles sin pérdida de XP (tres direcciones de FK deliberadamente
  distintas hacia `levels(id)`, ver [[Arquitectura#Rotación de niveles y XP|Arquitectura]]).
  `0008_matriz_permisos` mueve el GRANT/REVOKE de `sql/00_roles_unificado.sql`
  a una migración real.
- **Tres pools Postgres reales** (`usbi_app`/`usbi_moderador`/`usbi_dbmaint`)
  con permisos efectivos, funciones `SECURITY DEFINER` para cruzar la matriz
  sin ampliar GRANT.
- **Backend completo**: auth (registro en 3 pasos, JWT por UUID), quiz,
  niveles/secciones (jugador+admin), insignias, incidentes de seguridad,
  bitácora de auditoría con auto-auditoría, enlaces de interés + buzón de
  sugerencias anónimo, dispositivos (alta automática, revocación lógica) y
  sincronización offline, `usbictl` (migrate/secrets/admin/doctor/pgconf).
- **Frontend completo**: registro/login, `/settings` pública (temas,
  accesibilidad, daltonismo — un solo mecanismo de filtro, sin dependencia
  SVG externa desde el fix de Firefox), maker local, admin de contenido con
  purga real, vista de jugador (7 plantillas + Phaser), perfil, procesos
  offline, comunidad (enlaces + sugerencias), navegación estandarizada
  (`HomeButton`/`LinkButton`).
- **M1 — Veracidad del resultado de nivel**: completo (2026-09-17). Contrato
  `GameResult` en los 7 motores, callback unificado `onFinish(result)`,
  verificación en servidor para trivia/fake_news (por respuestas) y
  crossword/word_search/puzzle (por estado final); memory y snakes_ladders
  documentados como permanentemente no verificables sin reproducir la partida.
- **M2 — Avisos de privacidad**: mecanismo completo desde 2026-09-17
  (`GET /legal/privacy-notice` server-side, sello ligado al checksum del
  texto servido, `POST /legal/accept`, `/privacidad` pública). **Texto
  actualizado el 2026-09-23** de `v0-provisional` a `v0.9-borrador`
  (`backend/legal/aviso_simplificado_v1.json` / `aviso_integral_v1.json`,
  fuente completa en `pruebas/04_legal/LG-04_aviso_de_privacidad_v1.md`):
  corrige los hallazgos de la revisión del borrador previo (fuentes UV
  alineadas con `LG-03`, quita la mención errónea a `arco_requests` —esa
  tabla no existe en este proyecto, solo en `../usbi`—, sección ARCO
  distingue cancelación autoservicio de acceso/rectificación/oposición,
  fundamento legal tomado de `LG-02`, sección de menores redactada desde
  cero). **Sigue sin ser el aviso legal final**: "responsable del
  tratamiento = la USBI" y "centro de datos de Hostinger en México" son
  supuestos de trabajo confirmados verbalmente por el usuario, no una
  determinación formal del área jurídica de la UV — por eso la versión no
  subió a `v1.0` (ver TODO en `backend/legal/embed.go`). Contenedor `api`
  de `usbi-anon` reconstruido y reiniciado con este texto el 2026-09-23
  (`db`/`web` no se tocaron).
- **M3 — Despliegue reproducible**: completo (commits `8e1b100`…`d2d96a8`).
  `usbictl` único binario; `docker-compose.yml` de 3 servicios (`db`/`api`/`web`)
  versionado, reemplazando la instancia de un solo contenedor en
  `/home/altair/usbi-anon/`; `scram-sha-256` en vez de `trust`; `DEPLOY.md` con
  la secuencia real de despliegue; límite de 250 MB de RAM por contenedor.
- **Limpieza de comentarios** (commit `c3c3290`, 2026-09-21): quita marcadores
  `(Útil)`/`(Relleno)` y comentarios obsoletos que referenciaban fases F0–F11
  ya retiradas, en 63 archivos de `backend/internal/*`. Solo comentarios —
  ningún cambio funcional. Trabajo de rama (`limpieza-comentarios-y-normalizacion`)
  coordinado entre varios usuarios sobre el mismo checkout del repositorio.

## En curso

- Ninguna fase de código abierta ahora mismo. El plan vigente
  ([[Plan#m4_documentacion_tecnica_y_legal|Plan]]) tiene M4 como siguiente
  bloque, sin arrancar.
- **Contenido de prueba para revisión visual (2026-09-23)**: se sembraron vía
  API admin, en el entorno `usbi-anon` corriendo (`http://192.168.1.210:8092/`),
  2 secciones ("Prueba de fotos — Naturaleza", "Prueba de fotos — Cultura
  general"), 4 niveles publicados de 4 plantillas distintas (trivia, memory,
  fake_news, puzzle — 2 por sección) y 1 categoría de enlaces de interés
  ("Prueba de fotos — Redes") con 2 tarjetas externas genéricas (YouTube,
  Instagram). **Es contenido de prueba, no contenido real del curso** — no
  confundirlo con material educativo definitivo al planear M4 o rotación de
  temporadas; no quedó registrado en ninguna migración SQL, solo vive en la
  base de datos del contenedor `usbi-anon`.

## Falta

- **M4 — Documentación técnica y legal**: 10 artefactos técnicos (ER en 3
  vistas, diccionario de datos, diagrama de paquetes Go, matriz de permisos
  legible, diagrama de despliegue, 4 diagramas de secuencia, contrato OpenAPI,
  máquinas de estados, manual de operación, ADR) + 10 documentos legales, en
  dos paquetes impresos (UV / operación en la USBI). Ver
  [[Plan#m4_documentacion_tecnica_y_legal|Plan]].
- **Verificación manual en navegador de las 7 plantillas de nivel** (ganar y
  perder donde aplique) tras M1 — no se hizo por el límite de memoria del
  contenedor de desarrollo para `npm run build`.
- Deuda conocida fuera de M1–M5: paginación en catálogos extensos
  (`DashboardPage`, `SectionLevelsPage`, `AdminContentPage`), pulido de
  `AdminCommunityPage.tsx`, borrado duro de cuenta propia en
  `DELETE /api/v1/auth/me` (hoy seudonimización blanda, condicionado a fijar
  antes la política de conservación en M4.3), `internal/maintenance` sin
  evaluar para partición de pools.
- **Auditoría de Arquitectura (QA)**: Se detectaron ciclos de dependencia en el frontend (`registry.ts` con 17 ciclos, el más corto `registry.ts` ➔ `LevelMakerForm.tsx` ➔ `registry.ts`), lo cual representa un riesgo de empaquetado (evaluación `undefined` en runtime). En el backend, existen ciclos intra-paquete (ej. `auth/service.go` ➔ `auth/handler.go` con 41 ciclos, y `audit/audit.go` con 9), los cuales compilan en Go pero evidencian un alto acoplamiento semántico. Pendiente de prueba de integración (E2E) para verificar recuperación del cliente ante fallos.

## Siguientes pasos

1. Arrancar M4 cuando el usuario dé la instrucción — de las 4 decisiones
   abiertas del `CLAUDE.md`, **2 ya se confirmaron el 2026-09-22**: dominio
   (`www.msusbicoatza.com`) y tamaño del VPS (1 GB RAM / 1 vCPU / 20 GB), ver
   [[Decisiones#vps_hostinger_especificaciones_confirmadas|Decisiones]].
   Siguen abiertas: responsable del tratamiento en la UV (bloquea qué aviso de
   privacidad de la UV adoptar — candidatos preliminares revisados, ninguno
   calza exacto, pendiente de confirmar con Transparencia UV) y ubicación del
   centro de datos de Hostinger.
2. Verificación manual en navegador de M1 (7 plantillas) si se quiere cerrar
   del todo esa validación antes de M4.
3. ~~Reconfirmar con el usuario si `admin01`/`Admin12345678!` debe rotarse~~ —
   rotada el 2026-09-23 a petición del usuario: `admin01` / `Password123!`.
   Se aplicó con hash Argon2id vía `UPDATE accounts SET password_hash = …,
   token_version = token_version + 1` (mismo mecanismo que
   `repository.SetAccountPassword`, hoy sin ruta HTTP/CLI que lo invoque —
   no hay `usbictl admin reset-password` ni endpoint de autoservicio; solo
   `usbictl admin create` para cuentas nuevas). Pendiente real: agregar ese
   comando a `usbictl` para no depender de un `UPDATE` manual la próxima vez.
