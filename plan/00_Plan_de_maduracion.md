# Plan de maduración — USBI-Anon

> **Documento único de planeación vigente.** Sustituye y consolida los planes
> `00_Plan_maestro`, `01_Base_de_datos`, `02_Backend`, `03_Frontend`,
> `04_Rediseno_identidad_gustos` y `05_Contenido_maker_y_juego`, y retira de
> circulación la numeración vieja de fases (F0–F11, B1–B4, C1–C4, D1–D11). Los
> archivos anteriores quedan en `plan/_historico/` (fuera de git) como respaldo
> consultable, no como referencia activa. `plan/Convenciones_de_color_UV.md` se
> conserva vigente y sin cambios.
>
> **Fecha de redacción:** 2026-09-14
> **Numeración nueva:** M1 … M5. No reintroducir letras de planes anteriores.
> **Estado:** plan aprobado en sus decisiones de fondo; sin código escrito.

---

## 0. Decisiones de partida (tomadas con el usuario el 2026-09-14)

Estas seis decisiones condicionan todo lo que sigue. Cambiar cualquiera obliga a
releer el bloque correspondiente.

| # | Decisión | Consecuencia principal |
|---|---|---|
| D-01 | **El alojamiento pasa a un VPS de Hostinger.** Reemplaza al servidor de la Universidad Veracruzana como destino de producción. | Reaparece un **encargado del tratamiento tercero**, eliminado del proyecto en su premisa original. Obliga a convenio de encargo y, según la ubicación del datacenter, a cláusula de transferencia internacional. Ver M4.3. |
| D-02 | **Cada plantilla de juego define su propio criterio de victoria.** No hay umbral porcentual global. | Siete reglas distintas que documentar, probar y mantener. Ver M1.2. |
| D-03 | **La validación del resultado se cierra por fases:** primero el reporte honesto desde el cliente, después el recálculo en servidor, plantilla por plantilla. | M1 se parte en Fase A (corta, cierra el bug visible) y Fase B (larga, cierra el hueco de confianza). |
| D-04 | **El despliegue automático es un script/binario idempotente ejecutado en el servidor**, no un pipeline de CI/CD. | Quien clone el repo desde GitHub puede desplegar sin secretos en la nube ni runner con acceso SSH. Ver M3. |
| D-05 | **La app de escritorio futura es un cliente offline del mismo proyecto**, con JWT de jugador y sincronización posterior. | Aprovecha `devices` y `sync_events`, que ya existen en el esquema. No requiere abrir CORS. Ver M3.7. |
| D-06 | **Los cambios de versión del aviso de privacidad se notifican de forma informativa, no bloqueante.** El despliegue a producción no se hará con el aviso provisional. | La versión `v1.0-preliminar` desaparece antes del primer despliegue; no hay que migrar consentimientos de cuentas de prueba. Ver M2.5. |

### Advertencia que acompaña a D-01

La premisa fundacional del proyecto era no alojar en un tercero; el modelo de
identidad por UUID se diseñó, en parte, como compensación por la PII que sí se
guardaba en la propuesta original. Con Hostinger como host, el sistema vuelve a
tener un encargado externo, y la documentación legal debe declararlo en vez de
heredar el texto de "operación directa de infraestructura institucional". La
decisión está tomada y el plan la asume íntegramente; lo único que exige es que
el convenio de encargo y el aviso de privacidad digan la verdad sobre dónde
viven los datos.

---

## M1 — Veracidad del resultado de un nivel

### M1.1 Diagnóstico verificado

Tres defectos encadenados, confirmados leyendo el código:

1. **El frontend miente por construcción.** `frontend/src/features/content/OfficialLevelPage.tsx:100` envía
   `completed: true` como literal en el cuerpo de `POST /levels/:id/complete`, en
   todas las plantillas y en todos los desenlaces. El mismo patrón está en
   `LocalLevelPage.tsx`, que no persiste pero sí muestra resultado.

2. **El backend ya estaba bien y nadie lo usaba.** `backend/internal/levels/service.go:46`
   (`CalculateXP`) devuelve 0 cuando `completed` es falso, y
   `player_service.go:178` registra el evento como `level_failed` en
   `experience_history`; la racha diaria solo se marca si `completed` es
   verdadero (`player_service.go:169`). Toda la lógica de la derrota existe,
   compila y está cubierta por pruebas (`service_test.go:120`) — simplemente
   nunca se ejecuta, porque el único cliente que la invoca manda siempre
   `true`.

3. **Serpientes y Escaleras premia la derrota.** `SnakeLadderGame.tsx:78` sí
   calcula `score = 0` cuando gana la IA, pero llama a `onComplete` igual, y
   `CalculateXP` **ignora el score por completo** — solo mira dificultad,
   número de intento y el `completed` que llega hardcodeado. Perder contra la
   máquina otorga hoy exactamente la misma XP que ganarle.

**Efecto acumulado sobre los datos vivos:** `player_progress.completed`,
`experience_history.event_type`, la racha diaria, los contadores de
`account_retired_progress` y los umbrales de insignias están todos calculados
sobre una señal falsa. No es un defecto cosmético de la pantalla de resultados:
es la integridad del progreso de toda la plataforma.

**Segundo hueco, independiente del anterior:** el servidor no recalcula nada. El
campo `Answers json.RawMessage` existe en `CompleteLevelRequest`
(`dto.go:111`) y se descarta sin leerse. Un `curl` autenticado con
`{"score": 99999, "completed": true}` es indistinguible de una partida real.
El único freno actual es involuntario: `CountLevelAttemptsByDate` hace que a
partir del cuarto intento del día la XP sea 0 (`service.go:46`), lo que acota el
abuso pero no lo impide.

### M1.2 Criterio de victoria, plantilla por plantilla (D-02)

Ningún motor tiene hoy el concepto de "superado": `IGameState` expone
`status`, `score`, `maxScore` y `timeLeft`, y nada más. Cuatro de los siete
motores ya distinguen de hecho entre terminar y lograr; tres no pueden perderse.

| Plantilla | Señal existente en el motor | Criterio propuesto | ¿Se puede perder hoy? |
|---|---|---|---|
| `trivia` | `isFinished` al agotar preguntas; `score` acumula `100 + timeLeft*2` por acierto | **Aciertos ≥ 60 % de las preguntas.** Exige contar aciertos, que el motor hoy no guarda (solo acumula puntos con bonus de tiempo, que no es proporcional a los aciertos). Añadir `correctCount`. | Sí — fallar todo termina el juego igual |
| `fake_news` | `getScore()` = ítems acertados, `getMaxScore()` = total | **`score / maxScore` ≥ 60 %.** Ya es directamente proporcional; no requiere cambios en el motor más allá de exponer el resultado. | Sí |
| `crossword` | `validate()` marca `isFinished` solo si **todas** las celdas son correctas (`CrosswordEngine.ts:167`) | **`isFinished` tal cual.** Rejilla parcial = no superado. El motor ya es estricto. | No hay derrota, hay abandono |
| `word_search` | `isFinished` solo cuando `foundWords.length === words.length` (`WordSearchEngine.ts:135`) | **`isFinished` tal cual.** | Igual |
| `puzzle` | `isFinished` solo si todas las piezas quedan en su orden original | **`isFinished` tal cual.** | Igual |
| `memory` | `gameOver` cuando `matchedPairs === totalPairs` | **Terminar = superar.** No existe forma de perder: el memorama solo acaba emparejando todo. Documentarlo como decisión explícita, no dejarlo implícito. | No |
| `snakes_ladders` | `winner: 'player' \| 'ai' \| null` — **el único motor con derrota real modelada** | **`winner === 'player'`.** Es el arreglo de mayor impacto de todo M1. | Sí, y hoy se premia |

**Caso transversal que el criterio por plantilla no cubre: el abandono.** Si el
jugador cierra la pestaña a mitad de nivel no se envía nada y no queda registro.
Propuesta: **mantener ese comportamiento** (no registrar intento al abandonar) y
reservar el intento fallido para las partidas que sí terminan en derrota. Razón:
registrar abandonos exigiría latidos periódicos o `beforeunload`, ambos frágiles,
y castigaría una desconexión de red como si fuera un fracaso.

**Consecuencia de diseño para las tres plantillas sin derrota** (`memory`,
`word_search`, `puzzle`, y `crossword` en la práctica): en ellas "completado" y
"terminado" son lo mismo, y eso es correcto — no hay que inventarles una
condición de fracaso artificial para uniformar la tabla. La uniformidad falsa
aquí sería el error.

### M1.3 Fase A — reporte honesto desde el cliente

Cierra el defecto visible. No toca el esquema ni la API.

| Paso | Trabajo | Archivos |
|---|---|---|
| A-1 | Definir el contrato `GameResult { completed: boolean; score: number; maxScore: number }` en `@usbi/engine`. Hoy conviven dos firmas incompatibles: `onFinish(score)` y `onComplete(score, maxScore)`. | `packages/engine/src/interfaces/` |
| A-2 | Añadir `getResult(): GameResult` a los siete motores, cada uno con el criterio de M1.2. `TriviaEngine` necesita además un contador de aciertos nuevo. | `packages/engine/src/games/*` |
| A-3 | Unificar la firma de los siete componentes de juego a `onFinish(result: GameResult)`. | `src/features/games/**` |
| A-4 | `OfficialLevelPage.finishLevel(result)` envía `completed: result.completed` y el score real. | `OfficialLevelPage.tsx:93-107` |
| A-5 | Mismo cambio en `LocalLevelPage` para que el maker pruebe lo mismo que verá el jugador. | `LocalLevelPage.tsx` |
| A-6 | Pantalla de resultado: distinguir visualmente superado / no superado, explicar por qué (p. ej. "acertaste 4 de 10, necesitas 6") y ofrecer reintentar. Hoy solo hay un bloque "Resultado oficial" que asume éxito. | `OfficialLevelPage.tsx:135-150` |
| A-7 | Pruebas unitarias del criterio en los siete motores. Cuatro ya tienen `.test.ts`; `MemoryEngine`, `SnakeLadderEngine` y `FakeNewsEngine` no. | `packages/engine/src/games/*.test.ts` |

**Sin migración.** `CalculateXP`, `experience_history` y la racha ya se comportan
correctamente en cuanto reciben la verdad.

**Datos heredados:** la XP y las insignias acumuladas en las cuentas actuales se
obtuvieron bajo la lógica falsa. Como esas cuentas son de prueba y se eliminan
antes de producción (ver M3.8), no hace falta un recálculo retroactivo. Si
alguna cuenta tuviera que sobrevivir, la decisión correcta sería dejar la XP
intacta — el mismo principio que protege la XP en la rotación de temporadas.

### M1.4 Fase B — recálculo en servidor (D-03)

El objetivo es que el backend deje de creerle al cliente. Se ataca por orden de
verificabilidad, no por orden de aparición.

**B-0 — Contrato de `answers`.** Definir, por plantilla, qué evidencia envía el
cliente. Es el prerequisito de todo lo demás: sin formato, no hay nada que
recalcular. El campo ya existe en el DTO y hoy se tira.

**B-1 — `trivia` y `fake_news` (directamente verificables).** El servidor ya
tiene `levels.content` con `correct_index` e `isFake`. Recibe los índices
elegidos, cuenta aciertos, calcula `completed` y **descarta** el `completed` y el
`score` que mandó el cliente. Estas dos plantillas pasan a ser inmunes.

**B-2 — `crossword`, `word_search`, `puzzle` (verificables por estado final).**
El cliente envía la rejilla resuelta, las palabras encontradas o el orden final
de piezas; el servidor compara contra el contenido del nivel. Más trabajo que
B-1 porque hay que serializar estructuras, pero la verificación sigue siendo
determinista.

**B-3 — `memory` y `snakes_ladders` (no verificables sin reproducir la
partida).** El memorama depende de la secuencia de clics y Serpientes de tiradas
de dado generadas en el cliente. Verificarlas de verdad exigiría que el servidor
generase la semilla y validara un registro completo de jugadas — trabajo
desproporcionado para dos plantillas.

**La salida ya está en el esquema:** `experience_history.verification_method`
existe y hoy siempre vale `"online_direct"`. Propuesta: distinguir
`online_verified` (el servidor recalculó) de `online_reported` (el servidor
aceptó el reporte del cliente). Así la diferencia queda registrada en los datos y
es auditable, en vez de esconderse. Esa misma columna sirve después para marcar
el XP que llegue por sincronización desde la app de escritorio, donde el problema
es idéntico por naturaleza (ver M3.7).

**B-4 — Límite de tasa.** Documentar el freno existente (cuarto intento diario
sin XP) y añadir límite de peticiones por cuenta e IP en `POST
/levels/:id/complete`, que hoy no existe y pasa a importar cuando la API quede
expuesta públicamente.

---

## M2 — Avisos de privacidad: dónde, cómo y con qué forma

> **Alcance:** este bloque planifica ubicación, formato e implementación. **No
> redacta el texto legal**, que no existe todavía. El insumo de selección de
> plantilla UV (`pruebas/04_legal/LG-03`) se conserva; la redacción se ejecuta en
> M4.3.

### M2.1 Estado verificado

- Una sola casilla en `RegisterPage.tsx:238`, con la constante
  `PRIVACY_NOTICE_VERSION = 'v1.0-preliminar'` hardcodeada en el cliente
  (`RegisterPage.tsx:21`).
- **No hay texto que leer en ninguna parte.** La casilla afirma aceptar un aviso
  que el sistema no muestra.
- No existe ruta `/privacidad` en `App.tsx`; ninguna pantalla de la app menciona
  el aviso fuera del registro. El login no lo menciona.
- La base sí está preparada: `accounts.privacy_notice_version`,
  `privacy_notice_accepted_at` y `privacy_acceptance_hash` existen desde la
  migración `0001` (líneas 116–120), y el backend ya sella el hash.
- El backend contempla un aviso de staff diferenciado: la variable
  `STAFF_PRIVACY_NOTICE_VERSION` ya se lee.

**Defecto de confianza a corregir:** la versión aceptada la declara el cliente.
Un cliente modificado puede sellar "acepté v9.0" sin haber visto nada. El sello
debe calcularse sobre el texto que el servidor sirvió, no sobre la etiqueta que
el cliente dice.

### M2.2 Los dos avisos y dónde va cada uno

Se producen **dos textos** para jugadores (más uno de staff), según la
recomendación de `LG-03`: **simplificado** (corto, el que se lee en pantalla) e
**integral** (completo, el que contiene finalidades, transferencias, derechos y
datos del responsable).

| Momento | Qué se muestra | Formato | ¿Bloquea? |
|---|---|---|---|
| **Registro**, último paso, antes de generar credenciales | Aviso **simplificado íntegro**, visible en la propia página dentro de un contenedor con desplazamiento | Casilla obligatoria, **no premarcada**, + enlace "Ver aviso completo" que abre `/privacidad` en pestaña nueva | **Sí.** Sin casilla no hay cuenta |
| **Login** | Enlace discreto al pie: "Aviso de privacidad" | Enlace a `/privacidad` | No |
| **Dentro de la app (jugador)** | Entrada "Privacidad y datos" en la sección "Más" y en `/perfil` | Página `/privacidad`: simplificado arriba, integral desplegable debajo, y al final **la versión que esta cuenta aceptó y cuándo** | No |
| **Cambio de versión del aviso** | Banner descartable al entrar (D-06) | Al descartarlo se registra la aceptación de la versión nueva | No |
| **Staff / administradores** | Aviso de privacidad de staff, distinto del de jugadores | En el primer acceso de una cuenta admin | Por definir al redactarlo |
| **Cartel físico en la USBI** | Aviso simplificado impreso, con código QR a `/privacidad` | Impreso (ver M4.4) | — |

**Por qué el enlace lleva al integral y no al simplificado:** el simplificado se
muestra completo en el propio formulario. Esconder tras un enlace el texto que se
está aceptando convierte la casilla en un trámite vacío; el enlace debe llevar a
*más* información, no a la única que hay.

**Por qué "pestaña nueva" y no navegación:** el registro es un flujo de tres
pasos con estado en memoria. Navegar fuera y volver perdería las respuestas del
cuestionario.

### M2.3 Formato del control de aceptación

- **Casilla sin premarcar**, siempre. El consentimiento tiene que ser un acto.
- Etiqueta de la casilla: una frase, no un párrafo. El texto explicativo va
  arriba, en el bloque del aviso simplificado.
- **"Saber más" como acordeón accesible en la misma página** (`<details>` o
  equivalente con `aria-expanded`), **no como modal**. Tres razones concretas:
  este proyecto ya arrastró un fallo real en el que el filtro de daltonismo rompía
  todos los overlays `position: fixed`; los modales se comportan mal con zoom al
  200 %, que es requisito de accesibilidad vigente; y un acordeón es navegable con
  teclado sin trampa de foco.
- Área de clic mínima 44×44 px y contraste según
  `plan/Convenciones_de_color_UV.md`, igual que el resto de la interfaz.
- El bloque de aviso debe ser legible con lector de pantalla: encabezados reales,
  no `<div>` con estilo de título.

### M2.4 Implementación: cómo queda mejor distribuido

**Principio rector: el texto legal es dato del servidor, no constante del
cliente.**

1. **El texto vive en el repositorio, versionado con git**, en
   `backend/legal/` (por ejemplo `aviso_simplificado_v1.json`,
   `aviso_integral_v1.json`, `aviso_staff_v1.json`), empotrado en el binario con
   `go:embed`. Ventajas: el texto legal queda bajo control de versiones con
   historial de cambios (que es exactamente lo que un auditor pide), viaja con el
   binario, y no depende de que alguien cargue un archivo al servidor.

2. **Endpoint público nuevo:** `GET /api/v1/legal/privacy-notice` devuelve
   `{ version, effective_date, simplified, full, checksum }`. Público porque el
   aviso debe poder leerse **antes** de tener cuenta. Es el mismo criterio con el
   que `/settings` ya es pública.

3. **Formato del texto: JSON estructurado, no Markdown ni HTML.** Un arreglo de
   secciones `{ heading, paragraphs[] }` renderizado con componentes React. Evita
   meter un intérprete de Markdown en el bundle, evita `dangerouslySetInnerHTML`
   por completo, y da control total sobre tipografía, escalado de texto y
   estructura de encabezados para el lector de pantalla — los tres requisitos de
   accesibilidad que ya cumple el resto de la app.

4. **El sello lo calcula el servidor.** En el registro, el cliente envía la
   versión que vio; el backend **verifica que coincida con la vigente** y calcula
   `privacy_acceptance_hash` sobre el texto que él mismo sirve. Si no coincide,
   rechaza con 409 y obliga a recargar. La constante `PRIVACY_NOTICE_VERSION` del
   frontend desaparece.

5. **Estructura en el frontend** — feature nueva `src/features/legal/`:
   - `PrivacyPage.tsx` — ruta pública `/privacidad`, ambos avisos.
   - `PrivacyNoticeInline.tsx` — bloque reutilizable que usa el registro.
   - `PrivacyVersionBanner.tsx` — el aviso informativo de cambio de versión.
   - `usePrivacyNotice.ts` — carga y cachea el aviso vigente.
   - Ruta declarada junto a `/login`, `/register` y `/settings` en `App.tsx`,
     fuera del guardián de sesión.

6. **Enlaces de entrada:** pie del login, pie del registro, sección "Más" y
   `/perfil`. Cuatro puntos, un solo componente de enlace.

### M2.5 Cambio de versión (D-06)

No se desplegará a producción con el aviso provisional, así que no hay
consentimientos viejos que migrar. El mecanismo se diseña igualmente para cambios
futuros:

- Al iniciar la app con sesión activa se compara la versión aceptada por la
  cuenta contra la vigente.
- Si difieren: **banner informativo descartable**, no bloqueante.
- Al descartarlo, `POST /api/v1/legal/accept` registra la versión nueva, su fecha
  y el hash del texto vigente.
- **Aquí encuentra uso `GET /auth/me`**, documentado hasta ahora como endpoint sin
  consumidor: es la fuente natural de la versión aceptada por la cuenta.

### M2.6 Lo que este bloque deja pendiente a M4

- Redacción de los tres textos (jugador simplificado, jugador integral, staff).
- Retirada de `v1.0-preliminar` y numeración del aviso definitivo como `v1.0`.
- Decisión institucional sobre qué área de la UV figura como responsable del
  tratamiento y qué domicilio se publica, que no es una decisión técnica.

---

## M3 — Despliegue automatizado en Hostinger

> El bloque más largo. Se ejecuta de arriba abajo: cada paso asume el anterior
> hecho.

### M3.1 Estado verificado de la infraestructura actual

Lo que hoy existe es una **instancia de prueba de esfuerzo**, no un embrión de
producción, y conviene decirlo sin rodeos antes de planear sobre ella:

- Un contenedor único (`usbi-anon`, 500 MB de RAM) con Postgres 15, el backend Go
  y el frontend. `wait -n` sobre dos procesos: si cae uno, cae todo.
- **`pg_hba.conf` está en `trust`** para `local`, `127.0.0.1/32` y `::1/128`. Las
  contraseñas de `usbi_app`, `usbi_moderador`, `usbi_migrate` y `usbi_dbmaint`
  existen en `backend/.env.roles` pero **Postgres nunca las valida**. Nadie sabe
  hoy si son correctas.
- El frontend se sirve con `vite preview`, un servidor de desarrollo: sin
  compresión, sin caché, sin cabeceras de seguridad, y con una lista blanca de
  `Host` (`allowedHosts`) que hay que editar a mano cada vez que cambia el
  dominio.
- **Las migraciones solo se aplican si el esquema está vacío.** `0004`, `0005` y
  `0006` se aplicaron a mano. No hay tabla de control de versiones del esquema.
- Las migraciones corren como `postgres` (superusuario), y por eso el entrypoint
  necesita un bloque de `GRANT` correctivo: las tablas nacen sin permisos para los
  roles de la aplicación.
- `JWT_SECRET` y `HMAC_SECRET` viven en los argumentos `-e` de `docker run`,
  visibles con `docker inspect`, sin archivo ni rotación.
- Tres puertos publicados al host: frontend, backend y **Postgres**.
- El `entrypoint.sh` vive dentro de la imagen, no en el repositorio: editarlo
  exige `docker cp` y no queda registrado en git.

Nada de esto es un defecto del trabajo hecho — era una instancia para medir
comportamiento bajo presión de memoria y cumplió su función. Pero **ninguna de
esas siete piezas puede viajar a producción tal cual.**

### M3.2 Arquitectura de destino

**Un VPS Hostinger con tres contenedores orquestados por `docker compose`**, en
lugar del monolito:

| Servicio | Imagen | Publica | Notas |
|---|---|---|---|
| `db` | `postgres:15` oficial | **nada** — solo red interna | Volumen nombrado para `PGDATA`; overrides por `conf.d` |
| `api` | Multi-etapa: `golang` compila → imagen mínima | **nada** — solo red interna | Binario estático + migraciones y avisos legales empotrados |
| `web` | Multi-etapa: `node` compila `dist/` → `nginx` | **80 / 443** | Sirve el estático y hace proxy de `/api` a `api` |

Razones de la forma:

- **Postgres y el compilador de Go dejan de compartir contenedor.** Un fallo del
  backend deja de tumbar la base.
- **Solo 80 y 443 salen al mundo.** Hoy Postgres está publicado; en un VPS con IP
  pública eso es inaceptable.
- **nginx sustituye a `vite preview`**: compresión, caché por huella de archivo,
  cabeceras de seguridad (CSP, HSTS), terminación TLS, y el fin de la lista
  `allowedHosts`.
- **El artefacto que viaja es la imagen, no el código.** El pico de 600–700 MB de
  RAM de `vite build` ocurre en la etapa de compilación, no en el servidor en
  ejecución.

Alternativa considerada y descartada: instalación nativa con systemd. Es
perfectamente válida y consume menos recursos, pero reproducir a mano la
secuencia de roles, permisos y migraciones en cada servidor es exactamente el
problema que este bloque busca eliminar, y no le sirve a un tercero que clone el
repositorio.

### M3.3 Migraciones integradas al backend

**Decisión: `golang-migrate` como biblioteca, con los `.sql` empotrados por
`go:embed`, invocada por un subcomando explícito — nunca automática al
arrancar.**

- `//go:embed migrations/*.sql` hace que las doce migraciones viajen dentro del
  binario. No hay forma de desplegar un binario sin sus migraciones ni de aplicar
  migraciones de otra versión.
- Subcomando explícito (`usbictl migrate up`), no ejecución en el arranque:
  permite `down` para revertir, permite ver qué va a pasar antes de que pase, y
  no compite consigo mismo si algún día hay más de una instancia.
- **Se ejecuta con el rol `usbi_migrate`**, que ya existe y ya tiene contraseña
  propia, no con `postgres`. Consecuencia directa y valiosa: las tablas pasan a
  pertenecer a un rol de aplicación y **el bloque correctivo de `GRANT` del
  entrypoint deja de ser necesario**.

**Dos detalles que no pueden olvidarse:**

1. `golang-migrate` mantiene su estado en una tabla `schema_migrations`. La base
   actual tiene el esquema aplicado a mano hasta `0006` sin esa tabla: hace falta
   un bautizo (`migrate force 6`) documentado paso a paso, o partir de una base
   nueva. En producción se parte de base nueva, así que el bautizo solo afecta al
   entorno de desarrollo actual.
2. `backend/sql/00_roles_unificado.sql` **sí debe convertirse en migración**
   (`0007`): la matriz de `GRANT`/`REVOKE` es estructura de la base y hoy se
   reaplica en cada arranque desde un script suelto. `01_seed_primer_admin.sql`
   **no**: es un dato, lleva una contraseña, y su lugar es el subcomando
   `usbictl admin create` (ver M3.6).

### M3.4 ¿Módulos descartables de Go? — `usbictl`

**Recomendación: un único binario `cmd/usbictl` con subcomandos, no varios
módulos descartables.** Tres razones:

1. Todos necesitarían lo mismo — `internal/config`, `internal/crypto`,
   `internal/repository`. Repartirlos en módulos separados multiplica el código
   de arranque sin ganar aislamiento real.
2. Un módulo "descartable" que toca la base de producción **no se descarta
   nunca**: acaba siendo la herramienta de operación del sistema, y conviene que
   lo sea a propósito, documentado y con pruebas, en vez de por accidente.
3. Un solo binario es lo que se puede escribir en un manual de operación impreso
   sin que ocupe una página de rutas.

Subcomandos previstos:

| Subcomando | Qué hace |
|---|---|
| `migrate up \| down \| force \| version` | Migraciones con el rol `usbi_migrate` |
| `secrets init [--rotate <clave>]` | Genera y guarda los secretos (M3.6) |
| `admin create` | Siembra el primer administrador; sustituye al SQL a mano |
| `doctor` | Verifica los tres pools, la versión del esquema y los permisos efectivos (M3.5) |
| `pgconf render` | Rellena las plantillas de configuración de Postgres (M3.7) |

Lo que sí es genuinamente descartable y ya existe: `cmd/hash_password`, que se
absorbe dentro de `admin create` y desaparece como binario suelto.

### M3.5 Quitar el `trust` y comprobar que las credenciales sirven

**El problema real no es que `trust` sea inseguro** — que lo es —, **sino que
hace imposible saber si el sistema funciona con credenciales.** Las contraseñas
llevan meses configuradas y jamás se han validado; podrían estar todas mal y el
proyecto arrancaría igual.

Plan:

1. **`scram-sha-256`, no `md5`.** La única regla no-`trust` de hoy
   (`host all all 10.10.14.0/24 md5`) usa un método obsoleto. Fijar
   `password_encryption = scram-sha-256` **antes** de crear o rotar los roles: si
   se hace después, los hashes se quedan en md5 y hay que rotar de nuevo.
2. **Reglas mínimas**: acceso local por socket para tareas de mantenimiento, y
   `host` restringido a la red interna de compose para los tres roles de la
   aplicación. Ninguna regla que acepte tráfico externo.
3. **Rotar las cuatro contraseñas** con `usbictl secrets init` una vez activado
   scram, para partir de hashes válidos y conocidos.
4. **Comprobación explícita — `usbictl doctor`**, que es la respuesta concreta a
   "testear que el proyecto realmente sirva con las credenciales". Por cada uno de
   los tres pools:
   - conecta y ejecuta `SELECT current_user, session_user` (confirma que entra y
     con qué identidad),
   - ejecuta **una operación representativa de su matriz de permisos**: un
     `SELECT` sobre `accounts` con el pool de jugador; un `INSERT` sobre
     `suggestions` (que es el caso límite conocido: `usbi_app` no tiene `SELECT`
     ahí, y por eso el `INSERT` no puede llevar `RETURNING`); un `EXECUTE` de
     `ensure_yearly_partition` con `usbi_dbmaint`, que no tiene ni un `GRANT` de
     tabla,
   - y comprueba **que lo prohibido siga prohibido**: que `usbi_app` no pueda
     leer `security_incidents` ni actualizar `audit_log`.

   Un `doctor` que solo comprueba lo que debe funcionar no detecta un `GRANT` de
   más. Esa segunda mitad es la que evita que una migración futura afloje la
   matriz sin que nadie lo note.
5. **Ejecutarlo en cada despliegue**, no solo la primera vez.

### M3.6 Credenciales con el repositorio público (D-04)

El repositorio se publicará en GitHub. **Verificado: el historial está limpio** —
40 commits, sin remoto configurado, ningún archivo `.env` real jamás versionado
(solo `.env.example`) y ningún secreto literal en ningún commit. No hace falta
reescribir historia ni crear un repositorio nuevo.

De las tres opciones planteadas, la recomendación es la primera **con matices
importantes**: un subcomando `usbictl secrets init` que

- **genera** `JWT_SECRET` y `HMAC_SECRET` con `crypto/rand` — **no los pide por
  consola**. Un secreto tecleado por una persona es un secreto débil, y no hay
  ninguna razón para que un humano elija estos dos: nadie los teclea jamás
  después.
- **genera** igualmente las cuatro contraseñas de rol de Postgres.
- **pide por consola una sola cosa: la contraseña del primer administrador**,
  porque es la única credencial que una persona tiene que recordar y escribir. La
  valida contra las restricciones reales (el `CHECK` de `accounts.nickname` es
  `^[a-z0-9]{6,20}$` — sin mayúsculas) y contra una política de fuerza.
- **escribe un único archivo** de entorno con `chmod 600`, propiedad del usuario
  de servicio, **fuera del árbol del repositorio** (`/etc/usbi-anon/env`), para
  que un `git add -A` distraído no pueda alcanzarlo ni por accidente.
- **es idempotente**: si el archivo ya existe no lo pisa; `--rotate <clave>` rota
  un secreto concreto y avisa de qué hay que reiniciar.

Las otras dos opciones se descartan por razones concretas:

- **"Contraseñas temporales y cambio en el primer acceso"** obliga a añadir al
  modelo de cuentas un estado de "debe cambiar contraseña" que hoy no existe
  (columna, endpoint, pantalla, y su interacción con el flujo de login), y **no
  resuelve el problema principal**: `JWT_SECRET` y `HMAC_SECRET` no son
  credenciales de nadie y no pueden "cambiarse en el primer acceso".
- **"Contraseñas a mano en variables según la ruta"** es, literalmente, lo que
  hay hoy — secretos en los argumentos de `docker run`, invisibles para git,
  imposibles de auditar y de rotar.

Lo que sí viaja en el repositorio público: `.env.example` con todas las claves
documentadas y vacías, `docker-compose.yml`, `usbictl`, las plantillas de
configuración de Postgres y un `DEPLOY.md` con la secuencia exacta.

### M3.7 Los dos archivos de configuración de Postgres

La pregunta era si copiarlos al servidor o delegar en un módulo Go. **La
respuesta es: ninguna de las dos en su forma literal.** Copiar los archivos
completos los ata a una versión concreta de Postgres y los desincroniza en
silencio en la siguiente actualización.

- **`postgresql.conf`: no sobrescribir nunca.** Montar solo un archivo de
  *overrides* en `conf.d/usbi.conf` (`shared_buffers`, `work_mem`,
  `max_connections`, `ssl`, parámetros de bitácora) e incluirlo con
  `include_dir`. Así las novedades y valores por defecto de cada versión siguen
  llegando y solo se versiona lo que realmente se decidió cambiar.
- **`pg_hba.conf`: plantilla versionada y parametrizada**, no copia literal. Con
  la imagen oficial, `POSTGRES_INITDB_ARGS` fija el método de autenticación en la
  inicialización; las reglas propias se montan como archivo de solo lectura con
  las redes como variables, no como literales.
- **Dónde encaja el módulo Go:** `usbictl pgconf render` toma esas plantillas del
  repositorio y las rellena con los datos del servidor de destino — redes de la
  red interna de compose, y los valores de memoria calculados a partir de la RAM
  contratada. Es la pieza que hace que el mismo repositorio sirva en un VPS de
  4 GB y en uno de 16 GB sin editar nada a mano. **Es conveniente, no
  imprescindible**: con overrides versionados el despliegue ya funciona.
- **Añadir al plan una tabla de valores por tamaño de VPS.** Los parámetros de
  memoria actuales están calibrados para el límite artificial de 500 MB de la
  prueba de esfuerzo y no tienen ninguna relación con el servidor real.

### M3.8 Frontend precompilado, rutas y puertos

Las tres preguntas planteadas, respondidas:

**¿Cómo se manda el frontend?** Como imagen. El `Dockerfile` multi-etapa compila
`dist/` en una etapa con Node y copia únicamente el resultado a una imagen nginx.
Al servidor no llega ni el código fuente ni `node_modules`, y el pico de memoria
de `vite build` ocurre en la máquina que construye, no en el VPS.

**¿Hay que ajustar rutas y puertos?** **No, y es importante no hacerlo.**
`apiClient.ts:5` ya resuelve `apiBaseURL` como `'/api/v1'` relativo. Esa ruta
relativa **es** la detección automática correcta: el navegador pega contra el
mismo origen que sirvió la página, nginx hace proxy de `/api` al contenedor del
backend, y el mismo artefacto funciona en local, en el túnel de pruebas y en
Hostinger sin recompilar. Cablear un dominio en el bundle sería un retroceso.

**¿Detección automática?** Ya la hay, por ruta relativa, que es la forma más
robusta: no depende de variables en tiempo de compilación ni de adivinar el
entorno en tiempo de ejecución.

Lo que **sí** cambia:

- `vite preview` desaparece, y con él `preview.allowedHosts` — el pie de página
  de "403 Blocked request al añadir un dominio nuevo al túnel" deja de existir.
- Solo 80 y 443 quedan publicados. El backend y Postgres salen de la lista de
  puertos expuestos.
- `CORS_ALLOWED_ORIGIN` se fija al dominio real de producción. Hoy tiene como
  valor por defecto `https://usbi.edu.mx` cuando la variable está vacía, que ya
  no corresponde a dónde vive el sistema.

### M3.9 Exposición de la API para la app de escritorio (D-05)

La API ya son **121 rutas bajo `/api/v1`**. La app de escritorio será un cliente
offline del mismo proyecto, con JWT de jugador.

**Forma de exposición:** el mismo dominio y el mismo puerto 443 que el frontend
(`https://dominio/api/v1/…`), servido por nginx. **No un subdominio `api.`**: un
subdominio obliga a un segundo certificado, convierte todas las peticiones del
frontend en peticiones con CORS real, y expone el backend por separado sin ganar
nada a este tamaño.

**CORS:** una aplicación de escritorio nativa **no envía cabecera `Origin`**, así
que CORS no la afecta y **no hay que aflojarlo por ella**. Si la app acabara
siendo Electron o Tauri con un webview, sí enviaría un `Origin` propio
(`tauri://localhost`, `file://`) y habría que contemplarlo entonces,
explícitamente. Mientras tanto, no abrir nada preventivamente.

**Autenticación:** la app usa `POST /auth/login` y obtiene access + refresh, igual
que la web. El alta automática de dispositivo al iniciar sesión ya existe y cubre
el registro del cliente. La diferencia: la web guarda en `sessionStorage` (nunca
`localStorage`, decisión vigente); una app de escritorio debe guardar el refresh
en el almacén de credenciales del sistema operativo, no en un archivo plano.

**Lo que falta planear y no existe todavía:**

1. **Contrato de sincronización.** `sync_events` existe y hay pantalla de
   procesos offline, pero un cliente offline **reenvía**: el contrato tiene que
   ser idempotente por identificador de evento, o la XP se duplicará en la primera
   reconexión con red inestable.
2. **Procedencia del XP offline.** El XP ganado sin conexión tiene exactamente el
   mismo problema de confianza que las plantillas no verificables de M1.4-B3, y la
   misma solución: `experience_history.verification_method` distingue el origen.
   Los dos bloques deben usar el mismo vocabulario de valores; decidirlo una vez,
   en M1.4-B0.
3. **Compromiso de versión.** Una vez haya clientes de escritorio instalados,
   `/api/v1` no se puede romper: no hay forma de actualizar a todos a la vez.
   Escribir la política de compatibilidad antes de que exista el primer cliente,
   no después.
4. **Límite de tasa**, que hoy no existe en ninguna ruta y pasa a ser necesario
   cuando la API queda accesible desde internet.
5. **Contrato publicado en OpenAPI 3.1.** El insumo está en
   `pruebas/05_documentacion_tecnica/DT-03_contrato_api_v1.md`.

### M3.10 Orden de ejecución

1. `docker-compose.yml` + los dos `Dockerfile` multi-etapa; arrancar en local.
2. `cmd/usbictl` con `migrate` y `doctor`; migración `0007` con la matriz de
   permisos; bautizo de `schema_migrations`.
3. Cambio a `scram-sha-256` y plantillas de configuración de Postgres; `doctor`
   en verde, incluida la mitad que comprueba lo prohibido.
4. `usbictl secrets init` y `admin create`; secretos fuera de `docker run -e`.
5. nginx con TLS, cabeceras de seguridad y proxy de `/api`; cierre de los puertos
   de backend y base.
6. Ensayo completo en el VPS **partiendo de base vacía**, sin arrastrar datos de
   prueba.
7. `DEPLOY.md` escrito mientras se ejecuta el ensayo, no después: lo que no se
   documenta en el momento, se documenta mal.

### M3.11 Fuera del alcance de M3, pero necesario antes de abrir al público

- **Respaldos.** `pg_dump` programado, política de retención y —lo que de verdad
  importa— **una restauración de prueba ejecutada**. No hay respaldo hasta que se
  ha restaurado uno.
- **Renovación automática de TLS** y qué pasa si falla.
- **Bitácoras y monitoreo**: espacio en disco, y un aviso antes de que se llene.
- **Reconfirmar el límite de almacenamiento.** La regla de rotación de temporadas
  nació de un límite de ~20 GB en el servidor de la UV. En Hostinger el número
  depende del plan contratado y hay que reemplazarlo por el real: toda la
  arquitectura de archivado y purga está calibrada sobre esa cifra.

---

## M4 — Documentación técnica y legal

### M4.1 Qué artefactos, si no hay clases

La observación de partida es correcta: el sistema no usa programación orientada a
objetos, y un diagrama de clases sería una ficción dibujada para rellenar un
requisito. Lo que sí describe este sistema:

| # | Artefacto | Por qué aquí |
|---|---|---|
| 1 | **Diagrama entidad-relación, en tres vistas** | 30+ tablas en una sola hoja son ilegibles. Tres vistas por dominio: identidad y acceso; contenido y progreso; auditoría y comunidad. Más un mapa general de cómo se conectan las tres. |
| 2 | **Diccionario de datos** | Tabla por tabla y columna por columna, con tipo, nulabilidad y **la razón de negocio donde exista**. Sin ella, las tres direcciones deliberadamente distintas de las claves foráneas hacia `levels(id)` parecen una inconsistencia y el siguiente que las vea las "uniformará", rompiendo la conservación de XP al rotar temporadas. |
| 3 | **Diagrama de paquetes de Go** | **Este es el sustituto real del diagrama de clases.** Los veinte paquetes de `internal/`, sus dependencias, y a qué pool de base de datos pertenece cada uno. |
| 4 | **Matriz de permisos: rol × tabla × operación** | Es el contenido de `00_roles_unificado.sql` en forma legible. Probablemente el artefacto de seguridad más valioso del proyecto, y hoy no existe como documento — solo como SQL. |
| 5 | **Diagrama de despliegue** | VPS, contenedores, red interna, puertos, y qué cruza exactamente la frontera pública. |
| 6 | **Diagramas de secuencia de los cuatro flujos no obvios** | Registro en tres pasos; inicio de sesión con alta de dispositivo; completar un nivel (con la validación de M1); purgar un nivel preservando la XP. |
| 7 | **Contrato de API** | OpenAPI 3.1 más una tabla de los 121 endpoints por rol y pool. |
| 8 | **Máquinas de estados** | De una cuenta (`active` → `deleted`…) y de un nivel (borrador → publicado → archivado → purgado). La segunda es donde vive la regla de negocio de la rotación. |
| 9 | **Manual de operación** | Arranque, respaldo, restauración, rotación de secretos, rotación de temporada, qué hacer cuando `doctor` falla. |
| 10 | **Registro de decisiones de arquitectura (ADR)** | Las decisiones que un tercero **no puede deducir del código**: por qué se descartaron dos bases de datos; por qué se usan funciones `SECURITY DEFINER` en vez de ampliar permisos; por qué el `INSERT` en `suggestions` no lleva `RETURNING`; por qué el buzón de sugerencias deliberadamente **no** se audita. Todas están narradas en la bitácora; necesitan un formato estable. |

### M4.2 Cómo representar cada módulo

**Una ficha por paquete, de una página, con formato idéntico:**

- Nombre y ruta.
- Responsabilidad, en una frase.
- Pool de base de datos al que pertenece.
- Tablas que toca, con la operación concreta (leer / insertar / actualizar / borrar).
- Endpoints que expone.
- Dependencias sobre otros paquetes internos.
- Invariantes y reglas de negocio que sostiene.
- **Qué NO hace**, explícitamente.

El último punto no es relleno: es lo que impide que el siguiente que pase le
cuelgue responsabilidades ajenas a un paquete. `internal/suggestions` es el
ejemplo — "no registra en la bitácora de auditoría" es una decisión de privacidad
deliberada que, sin escribirla, parece un olvido y alguien la "arreglará".

El diagrama de paquetes (artefacto 3) hace de índice de las fichas. Veinte fichas
para `internal/`, más tres para el frontend por capa (`features/`, `shared/`,
`packages/`).

### M4.3 Documentación legal, sin datos personales

**Matiz necesario antes de la lista.** "No guardamos datos personales" no es
exacto en términos legales. UUID, nickname derivado de respuestas reales,
respuestas del cuestionario de registro, edad autorreportada y dirección IP en
las bitácoras son **datos personales seudonimizados**, no datos anónimos: mientras
exista posibilidad razonable de reidentificación, la normativa aplica igual. Lo
que cambia respecto a un sistema convencional es que el riesgo y el volumen son
mucho menores — no que la obligación desaparezca. Y con D-01, además, hay un
encargado del tratamiento externo que antes no existía. La documentación debe
partir de ahí; un aviso que afirme "no tratamos datos personales" sería a la vez
falso y frágil ante una revisión.

| # | Documento | Notas |
|---|---|---|
| 1 | **Aviso de privacidad integral** (jugadores) | El texto completo. Estructura recomendada en `LG-03`. |
| 2 | **Aviso de privacidad simplificado** (jugadores) | El que se muestra en pantalla (M2) y se imprime en cartel. |
| 3 | **Aviso de privacidad de staff** | Diferenciado; el backend ya prevé su versión por separado. |
| 4 | **Documento de seguridad** | Inventario, análisis de riesgo, medidas técnicas y organizativas, gestión de bitácoras e incidentes. Existe modelo previo reutilizable. |
| 5 | **Inventario de datos personales / registro de tratamientos** | `LG-01` ya lo tiene muy avanzado; hay que actualizarlo con D-01. |
| 6 | **Convenio de encargo de tratamiento con el proveedor de alojamiento** | **Cambia por completo con D-01.** Deja de ser "la UV opera su propia infraestructura" y pasa a ser un encargado tercero, con cláusula de transferencia internacional si el centro de datos está fuera de México. |
| 7 | **Evaluación de impacto en la protección de datos** | Recomendable aunque el volumen sea bajo: hay menores de edad involucrados. Existe modelo previo. |
| 8 | **Procedimiento de ejercicio de derechos ARCO** | **Hueco real que hay que resolver, no rodear.** El código de ARCO se eliminó y sin correo electrónico no hay canal de contacto. El procedimiento tendrá que ser presencial en la USBI, y debe explicar **cómo se acredita la identidad de quien ejerce un derecho sobre una cuenta anónima** — que es precisamente lo difícil cuando el sistema, por diseño, no sabe quién es nadie. |
| 9 | **Política de conservación y borrado** | Cuánto viven las bitácoras, los intentos, las direcciones IP, y qué queda tras el borrado de cuenta. |
| 10 | **Nota sobre menores y autorreporte de edad** | Cómo se justifica que un menor autorreportado juegue sin consentimiento de tutor, dado que no hay correo con el cual pedirlo. |

### M4.4 Qué se entrega impreso y a quién

Dos paquetes, con cuerpo técnico compartido.

**Paquete A — Responsables de la Universidad Veracruzana**

| Documento | Origen |
|---|---|
| Avisos de privacidad: integral, simplificado y de staff | M4.3 · 1, 2, 3 |
| Documento de seguridad | M4.3 · 4 |
| Inventario de datos personales | M4.3 · 5 |
| Convenio de encargo con el proveedor de alojamiento | M4.3 · 6 |
| Procedimiento de ejercicio de derechos (presencial) | M4.3 · 8 |
| Diagrama entidad-relación (tres vistas) y diccionario de datos | M4.1 · 1, 2 |
| Matriz de permisos rol × tabla × operación | M4.1 · 4 |
| Diagrama de despliegue | M4.1 · 5 |
| Manual de operación | M4.1 · 9 |
| Procedimiento de resguardo de credenciales | M4.5 |

**Paquete B — Operación y usuario final en la USBI**

| Documento | Notas |
|---|---|
| Manual del jugador | Registro, credenciales generadas, cómo se juega, qué pasa si se pierde la contraseña — **la respuesta honesta aquí importa: sin correo, la recuperación exige a un administrador** |
| Manual del administrador | Contenido, rotación de temporadas, insignias, auditoría, incidentes, cuentas |
| Cartel del aviso simplificado, con código QR a `/privacidad` | Visible en la sala |
| Procedimiento de resguardo de credenciales | M4.5 |

### M4.5 Credenciales en papel

**Las contraseñas de máquina, no. El procedimiento y la custodia, sí.**

- **Nunca en papel circulante:** `JWT_SECRET`, `HMAC_SECRET` y las contraseñas de
  los cuatro roles de Postgres. Nadie las teclea jamás: las genera
  `usbictl secrets init` y viven en un archivo con permisos 600 en el servidor.
  Imprimirlas solo multiplica los lugares desde los que pueden filtrarse.
- **Sí, en sobre sellado y firmado, con registro de entrega y bajo llave:**
  - la contraseña del **primer administrador**, y
  - las credenciales de acceso al VPS (SSH y panel de Hostinger).

  Son las dos únicas credenciales que una persona necesita para recuperar el
  sistema si todo lo demás se pierde.
- **Sí, como documento que circula:** el **procedimiento** — quién custodia el
  sobre, cómo se abre, cómo se registra la apertura, cada cuánto se rotan los
  secretos y qué se hace si el sobre apareció abierto.

**La razón operativa es dura y conviene escribirla en el propio documento:** al no
haber correo electrónico en el sistema, **no existe recuperación automatizada de
ninguna cuenta, incluida la del administrador**. Si se pierde la contraseña del
único administrador, la única vía es el acceso directo a la base de datos. El
sobre sellado no es burocracia: es el plan de continuidad del sistema.

---

## M5 — Consolidación documental

Se ejecuta al cerrar la redacción de este plan (D-04 de la segunda ronda:
consolidar conservando respaldo).

1. Este documento pasa a ser el **único plan vigente**;
   `plan/Convenciones_de_color_UV.md` se conserva intacto y sigue vigente.
2. Los seis planes anteriores y `estado_proyecto.pre-purga-2026-09-09.md` se
   mueven a `plan/_historico/`, excluido de git — conservados, no borrados.
3. Desaparece de la documentación activa la numeración F0–F11, B1–B4, C1–C4 y
   D1–D11. Lo que seguía pendiente de esas fases queda absorbido aquí:

| Pendiente heredado | Dónde vive ahora |
|---|---|
| Reescritura legal completa y `PrivacyPage` | M2 (ubicación y forma) + M4.3 (textos) |
| Borrado duro de la cuenta propia | M4.3 · 9 lo condiciona; ejecución tras fijar la política de conservación |
| Paginación en catálogos extensos | Pendiente de producto, fuera de M1–M5; se registra como deuda conocida |
| Pulido del formulario de comunidad | Igual |
| `internal/maintenance` sin evaluar para partición de pools | Se evalúa en M3.5, al comprobar permisos efectivos por pool |
| Herramienta formal de migraciones | M3.3 |
| Nombre y dominio final del producto | Bloqueante para M3.8 (CORS y TLS) y para M4.3 (avisos) |

4. Se actualizan `CLAUDE.md` (premisa de alojamiento según D-01, fases vigentes),
   `README.md` (estado y árbol de carpetas) y `SKILLS.md`.
5. Se añade la sección correspondiente a `estado_proyecto.md`.

### Decisiones que siguen abiertas y bloquean trabajo

| Decisión pendiente | Qué bloquea |
|---|---|
| Nombre y dominio definitivos del producto | TLS, `CORS_ALLOWED_ORIGIN`, y el nombre del responsable en los avisos |
| Qué área de la UV figura como responsable del tratamiento | Redacción de los tres avisos (M4.3) |
| Tamaño del VPS contratado en Hostinger | Valores de memoria de Postgres (M3.7) y el límite real de almacenamiento que gobierna la rotación de temporadas (M3.11) |
| Ubicación del centro de datos del proveedor | Cláusula de transferencia internacional en el convenio de encargo (M4.3 · 6) |
