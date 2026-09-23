---
tipo: estado_proyecto
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-19
---

# Plan — USBI-Anon

Plan vigente: `plan/00_Plan_de_maduracion.md` (bloques M1–M5, cerrado el
2026-09-14). La numeración anterior (F0–F11, B1–B4, C1–C4, D1–D11) quedó
retirada — no reintroducirla en documentación nueva. Detalle completo de cada
bloque en ese archivo; aquí solo la estructura de tareas y su estado.

## m4_documentacion_tecnica_y_legal & M4 — Documentación técnica y legal

Siguiente bloque a ejecutar. Nada de esto tiene código ni texto escrito
todavía.

### Artefactos técnicos (M4.1–M4.2)

- [ ] Diagrama entidad-relación en tres vistas (identidad/acceso; contenido/progreso; auditoría/comunidad) + mapa general
- [ ] Diccionario de datos, tabla por tabla y columna por columna, con la razón de negocio donde exista
- [ ] Diagrama de paquetes de Go (sustituto del diagrama de clases) — 20 paquetes de `internal/`, dependencias y pool al que pertenece cada uno
- [ ] Matriz de permisos rol × tabla × operación en forma legible (hoy solo existe como SQL en `00_roles_unificado.sql`/migración `0008`)
- [ ] Diagrama de despliegue (VPS, contenedores, red interna, puertos, frontera pública)
- [ ] 4 diagramas de secuencia: registro en 3 pasos; login con alta de dispositivo; completar un nivel (con verificación de M1); purgar un nivel preservando XP
- [ ] Contrato de API: OpenAPI 3.1 + tabla de endpoints por rol y pool
- [ ] Máquinas de estados: cuenta (`active`→`deleted`…) y nivel (borrador→publicado→archivado→purgado)
- [ ] Manual de operación (arranque, respaldo, restauración, rotación de secretos, rotación de temporada, qué hacer cuando `usbictl doctor` falla)
- [ ] Registro de decisiones de arquitectura (ADR) — ver [[Decisiones|Decisiones]] como fuente
- [ ] Ficha de una página por paquete Go (20) + 3 del frontend (`features/`, `shared/`, `packages/`), formato uniforme con sección explícita "qué NO hace"

### Documentación legal (M4.3)

Matiz obligatorio en todos: son datos personales **seudonimizados**, no
anónimos — la obligación normativa no desaparece, solo baja el riesgo/volumen.

- [ ] Aviso de privacidad integral (jugadores)
- [ ] Aviso de privacidad simplificado (jugadores, el de pantalla/cartel)
- [ ] Aviso de privacidad de staff (diferenciado)
- [ ] Documento de seguridad (inventario, riesgo, medidas técnicas/organizativas, gestión de bitácoras e incidentes)
- [ ] Inventario de datos personales / registro de tratamientos (actualizar `LG-01` con D-01)
- [ ] Convenio de encargo de tratamiento con el proveedor de alojamiento (cambia por completo con D-01; cláusula de transferencia internacional si el centro de datos está fuera de México)
- [ ] Evaluación de impacto en la protección de datos
- [ ] Procedimiento de ejercicio de derechos ARCO — presencial en la USBI; debe resolver cómo se acredita identidad sobre una cuenta anónima
- [ ] Política de conservación y borrado (bitácoras, intentos, IPs, qué queda tras borrar cuenta)
- [ ] Nota sobre menores y autorreporte de edad sin consentimiento de tutor

### Entrega y credenciales (M4.4–M4.5)

- [ ] Paquete A (UV): avisos + documento de seguridad + inventario + convenio + procedimiento ARCO + ER/diccionario + matriz de permisos + diagrama de despliegue + manual de operación + procedimiento de resguardo de credenciales
- [ ] Paquete B (operación/USBI): manual del jugador, manual del administrador, cartel del aviso simplificado con QR a `/privacidad`, procedimiento de resguardo de credenciales
- [ ] Procedimiento de resguardo de credenciales: secretos de máquina nunca en papel; contraseña del primer admin + credenciales VPS en sobre sellado con registro de entrega

## Completado

### M1 — Veracidad del resultado de un nivel

**Sección de origen:** m1_veracidad_resultado_nivel
**Cerrado:** 2026-09-17

- [x] Fase A — contrato `GameResult` en los 7 motores, callback unificado `onFinish(result)`, `OfficialLevelPage`/`LocalLevelPage` dejan de mandar `completed: true` fijo, pantalla de resultado distingue superado/no superado
- [x] Fase B — recálculo en servidor: `internal/levels/verify.go`, verificación por respuestas (trivia, fake_news) y por estado final (crossword, word_search, puzzle); memory y snakes_ladders documentados como no verificables sin reproducir la partida
- [x] Migración `0007` — `verification_method` distingue `online_verified`/`online_reported`

### M2 — Avisos de privacidad: dónde, cómo y con qué forma

**Sección de origen:** m2_avisos_privacidad
**Cerrado:** 2026-09-17

- [x] `GET /legal/privacy-notice` server-side (`go:embed`, texto `v0-provisional` hasta M4.3)
- [x] Sello de aceptación ligado al checksum del texto servido
- [x] `POST /legal/accept` para banner de cambio de versión (D-06, informativo, no bloqueante)
- [x] `/privacidad` pública, acordeón `<details>` accesible
- [x] `RegisterPage` consume el aviso real (deja de usar `PRIVACY_NOTICE_VERSION` hardcodeado)
- [x] `GET /auth/me` extendido — primer uso real desde que D7 lo dejó documentado como no usado

### m3_despliegue_hostinger & M3 — Despliegue automatizado en Hostinger

**Sección de origen:** m3_despliegue_hostinger
**Cerrado:** 2026-09-17/19 (commits `8e1b100`, `9f58842`, `0e25e31`, `6d77781`, `d2d96a8`)

- [x] `usbictl` único binario: `migrate`/`secrets`/`admin`/`doctor`/`pgconf`
- [x] Matriz de permisos movida de `sql/00_roles_unificado.sql` a migración `0008_matriz_permisos`
- [x] `docker-compose.yml` de 3 servicios (`db`/`api`/`web`), versionado, reemplaza la instancia de un solo contenedor
- [x] `scram-sha-256` en vez de `trust`
- [x] Bug de comillas duplicadas en contraseñas de rol (`00_init_cluster.sh` vs. `:'app_password'` de psql) — corregido, commit `0e25e31`
- [x] `DEPLOY.md` con la secuencia real de despliegue, escrita durante el ensayo
- [x] Límite de memoria de 250 MB por contenedor (`db`/`api`/`web`)

### M5 — Consolidación documental

**Sección de origen:** m5_consolidacion_documental
**Cerrado:** 2026-09-14

- [x] `plan/00_Plan_de_maduracion.md` como único plan vigente
- [x] Planes anteriores + `estado_proyecto.pre-purga-2026-09-09.md` movidos a `plan/_historico/`, excluidos de git
- [x] Retirada la numeración F0–F11/B1–B4/C1–C4/D1–D11 de la documentación activa
- [x] `CLAUDE.md`, `README.md`, `SKILLS.md` actualizados
- [x] Migración de `estado_proyecto.md` a `grafo_ia` (`.graph/Estado_Proyecto/`) — ejecutada 2026-09-19, ver [[Decisiones|Decisiones]]
