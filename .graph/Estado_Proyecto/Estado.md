---
tipo: estado_proyecto
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-28
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
- **Bug de UI corregido (2026-09-28): pantalla duplicada de fin de nivel en
  trivia.** Reporte de usuarios: al terminar un nivel aparecía por ~1 segundo
  una pantalla de "reintentar" y luego, encima, la pantalla final con
  XP/intentos/racha. Causa raíz: `TriviaGame.tsx` renderizaba su propia
  tarjeta "Trivia Completada" (con botón "Volver a jugar" y solo el score
  local) apenas `state.isFinished` se ponía en `true`, **antes** de que
  `OfficialLevelPage`/`LocalLevelPage` terminaran de mostrar su propia
  pantalla de resultado (que en `OfficialLevelPage` espera la respuesta del
  servidor vía `POST /levels/:id/complete` — la única fuente de verdad tras
  M1). Las otras 6 plantillas (`MemoryGame`, `FakeNewsGame`,
  `SnakeLadderGame`, `WordSearchGame`, `PuzzleGame`, `CrosswordGame`) ya
  seguían el patrón correcto: solo llaman a `onFinish`, sin pintar su propia
  pantalla de cierre; `TriviaGame` era la única plantilla que se
  desviaba — de ahí que el fenómeno solo se reportara en niveles de trivia.
  Corrección: `TriviaGame` ahora retorna `null` cuando `state.isFinished`,
  igual que las demás plantillas — una sola pantalla de cierre, la de la
  página contenedora. `PuzzleGame` y `CrosswordGame` muestran un banner
  "completado" inline (no reemplazan toda la pantalla ni tienen botón de
  reintentar) — mismo patrón de raíz pero no la pantalla reportada; se dejan
  igual por ahora. **Desplegado (2026-09-28)**: corrección de nota anterior
  — el entorno visible en `192.168.1.210:8092` / `usbi.heimdall-lab.com` ya
  no es el contenedor único de prueba de esfuerzo descrito en `CLAUDE.md`
  (ese sigue existiendo en `/home/altair/usbi-anon/`, aparte, sin tocar);
  es el compose de 3 servicios de M3 (`usbi-anon-db-1`/`api-1`/`web-1`), cuyo
  `web` se construye con un Dockerfile multi-etapa (`frontend/Dockerfile`)
  que compila `dist/` con Node dentro de la imagen — no depende de la
  memoria del contenedor en runtime. Se reconstruyó con
  `docker compose build web && docker compose up -d web`; el bundle server
  actualmente ya no contiene la tarjeta "Trivia Completada" retirada.

## En curso

- Ninguna fase de código abierta ahora mismo. El plan vigente
  ([[Plan#m4_documentacion_tecnica_y_legal|Plan]]) tiene M4 como siguiente
  bloque, sin arrancar.
- **Limpieza de contenido de prueba (2026-09-26)**: el contenido sembrado el
  2026-09-23 para revisión visual (las 2 secciones "Prueba de fotos", sus 4
  niveles, la categoría "Prueba de fotos — Redes" con sus 2 tarjetas) **se
  purgó por completo**, junto con residuos de pruebas `golden` anteriores que
  habían quedado archivados o en borrador sin publicar (8 secciones/niveles
  "Golden admin"/"Golden errores admin" + 1 sección/nivel borrador sin
  archivar que ninguno de los dos listados de admin mostraba, y 2 categorías
  de enlaces vacías del mismo origen). Motivo: el usuario va a mostrar este
  entorno como prototipo final y quería la base limpia para sembrar
  contenido nuevo. Estado resultante: `sections`, `levels`,
  `sections/archived`, `levels/archived`, `admin/interest-links` y
  `admin/interest-link-categories` devuelven todos `items: []`.
- **Contenido de demostración para el prototipo final (2026-09-26)**: sembrado
  vía API admin sobre la base ya limpia. 2 secciones publicadas —
  "¿Qué tanto sabes de psicología?" (`#18529D`) y "¿Qué tan bien conoces tu
  cuerpo?" (`#28AD56`) — con 5 niveles publicados cada una (10 en total),
  usando las 7 plantillas de `levels.template_type` con solo 3 repetidas
  (trivia, memory y fake_news aparecen 2 veces; puzzle, word_search,
  crossword y snakes_ladders 1 vez cada una): "Trivia: emociones y mente
  humana", "Memorama: conceptos de psicología", "Noticia real o falsa: mitos
  de la mente", "Sopa de letras: palabras de la mente", "Rompecabezas: frase
  sobre la salud mental" (sección psicología) y "Trivia: el cuerpo humano en
  movimiento", "Crucigrama: partes del cuerpo", "Serpientes y escaleras: reto
  de actividad física", "Memorama: deportes y beneficios", "Noticia real o
  falsa: mitos del cuerpo" (sección cuerpo). También se creó la categoría de
  enlaces de interés "Museo AMI" con 3 tarjetas de un proyecto hermano del
  mismo cliente (`msusbicoatza.com`): Audiocuentos (`#18529D`), Mapa sonoro
  (`#28AD56`) y Museo AMI (`#000000`). **Sigue siendo contenido de
  demostración, no material educativo definitivo** — igual que el sembrado
  del 2026-09-23, no quedó registrado en ninguna migración SQL, solo vive en
  la base de datos del contenedor `usbi-anon`.

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
- **Bug real, descubierto el 2026-09-26: `DELETE /api/v1/levels/{id}` (purga)
  falla siempre con 500 en el despliegue de 3 servicios.** Causa:
  `AccumulateRetiredProgressForLevel` (que preserva el XP/contador de
  niveles completados antes del `DELETE` físico, ver regla de rotación en
  `CLAUDE.md`) hace `SELECT` sobre `player_progress` usando el pool
  `usbi_moderador`, que no tiene ese GRANT — solo `usbi_app` (pool jugador)
  puede leer esa tabla. Ningún rol de aplicación existente puede correr la
  secuencia completa (leer `player_progress` + escribir
  `account_retired_progress` + `DELETE` en `levels`) sin superusuario. El
  patrón correcto ya existe en el repo
  (`backend/migrations/0004_procedimientos_purga_moderador.up.sql`): una
  función `SECURITY DEFINER` propiedad de `usbi_app` con `GRANT EXECUTE` a
  `usbi_moderador`, nunca ampliar el `GRANT` directo. **No se corrigió**: a
  petición explícita del usuario, la limpieza de contenido de este día se
  hizo con una purga manual por SQL como superusuario de Postgres
  (`docker exec` contra `usbi-anon-db-1`), sin tocar código ni migraciones.
  Sigue pendiente escribir esa migración — el próximo admin que purgue un
  nivel desde `/admin/content` en este entorno va a chocar con el mismo 500.

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
