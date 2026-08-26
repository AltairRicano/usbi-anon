# 05 — Contenido, maker y vista de juego (F10.6 – F10.12)

Documento de diseño del bloque de fases que devuelve a USBI-Anon todo lo que
`../usbi` tenía y que no era tratamiento de datos personales: el maker de
niveles, el administrador de contenido, los minijuegos, la vista de jugador y
la navegación entre ambas vistas.

**Este documento manda sobre `plan/03_Frontend.md` y sobre la §5 de
`plan/04_Rediseno_identidad_gustos.md`** en todo lo relativo a contenido,
maker y juego. La §5 del 04 describe únicamente el frontend de identidad
(registro/login/admin de cuentas), que es lo que F10 entregó; nunca describió
el resto de la aplicación, y esa omisión es exactamente el hueco que este
documento cierra.

Va **antes de F11 (legal)**, que sigue siendo la última fase y no tiene
dependencia técnica con nada de lo de aquí.

---

## 1. Diagnóstico: qué falta de verdad y por qué

La percepción de "el backend está a medias" es entendible pero no exacta. El
reparto real, verificado archivo por archivo contra `../usbi`:

### 1.1 Lo que el backend YA tiene (paridad completa o superior)

`backend/internal/levels/` está portado íntegro y es **más estricto** que el de
`../usbi`:

- `GET/POST /sections`, `PATCH /sections/{id}`, publish, unpublish, archive.
- `GET/POST /levels`, `GET/PATCH /levels/{id}`, publish, unpublish, archive,
  `POST /levels/{id}/complete`.
- `GET /profile/progress`.
- `validateLevelInput` en `service.go:704` valida título, color, dificultad
  (1–10), tamaño del `content`, JSON bien formado, `template_type` contra
  `AllowedTemplateTypes`, y **despacha a un validador por plantilla**:
  `validateTriviaContent`, `validateMemoryContent`, `validateFakeNewsContent`,
  `validateWordSearchContent`, `validatePuzzleContent`,
  `validateCrosswordContent`, `validateSnakesContent`. Las siete plantillas
  tienen validación de estructura y de tipos en Go.
- El `CHECK` de `levels.template_type` en el esquema fija el mismo vocabulario
  de siete plantillas a nivel de base de datos.

O sea: **la capa de contenido del backend no está a medias, está terminada** —
y las validaciones de formato/tipado que se pidieron sí existen, del lado del
servidor, que es el lado que importa. Lo que falta del backend es una lista
corta y concreta (§1.3).

### 1.2 Lo que falta es el frontend

`frontend/src/features/` de USBI-Anon tiene hoy cinco carpetas: `auth`,
`admin-accounts`, `admin-quiz-bank`, `home`, `settings`. De `../usbi` faltan
por completo: `content/` (incluido `content/maker/` con sus 7 formularios y 7
previsualizadores), `maker/`, `games/`, `dashboard/`, `profile/`, y los dos
paquetes locales `frontend/packages/engine/` y `frontend/packages/schema/`.

Son ~2 200 líneas de TSX de maker/contenido más los motores de juego. Nada de
eso se borró por decisión: nunca se planeó, porque el árbol de archivos de la
§5 del plan 04 solo listaba el frontend de identidad. Es el mismo tipo de
hueco que ya produjo la ausencia de `settings/` (corregida en F10.5).

`home/HomePage.tsx` es hoy una landing mínima post-login. Eso es lo que se
percibe como "un cascarón para agregar usuarios", y es literalmente correcto:
la aplicación sabe registrar y administrar cuentas, y no sabe hacer nada más.

### 1.3 Lo que falta del backend (poco, pero real)

| Falta | Por qué | Fase |
|---|---|---|
| Purga irreversible de secciones/niveles archivados | No existe en `../usbi` tampoco. Archivar es hoy el último estado posible: **no libera un solo byte**, y la regla de rotación por temporadas depende de que sí se libere | F10.9 |
| Poda de los roles `operator` y `director` | Heredados verbatim de `../usbi` en F3, nunca pedidos, sin ningún permiso asociado en el código | F10.6 |
| Borrado duro de la propia cuenta | `DELETE /auth/me` existe pero hace cancelación blanda (seudonimización heredada de un modelo con PII que ya no aplica) | F10.11 |

---

## 2. Decisiones cerradas (2026-08-25)

Tres decisiones tomadas con el usuario antes de redactar este documento. No
volver a abrirlas sin instrucción explícita.

### 2.1 Roles: se reducen a `player` + `admin`

`operator` y `director` se eliminan del `CHECK` de `accounts.role`, de
`domain.UserRole` y de cualquier lista de roles del frontend. **Nunca fueron
pedidos**: entraron copiados de `../usbi` cuando F3 replicó su `domain`, y en
todo el código no hay una sola comprobación de autorización que los use — solo
`admin` y `player` deciden algo.

### 2.2 Eliminar cuenta: borrado duro real

Sin PII en el sistema, seudonimizar no protege nada: solo deja UUIDs huérfanos
ocupando espacio. Una cancelación borra la fila de `accounts`, sus respuestas
del cuestionario de registro y su progreso. Solo sobrevive la bitácora de
auditoría con el UUID, que es lo que sostiene el no-repudio.

### 2.3 Purga: irreversible, con doble confirmación, y solo sobre archivados

Purgar libera almacenamiento de verdad y **no le quita a nadie la experiencia
ganada** — es justo el caso para el que se diseñaron las tres direcciones de
borrado distintas en las FK a `levels(id)` (ver `CLAUDE.md`, "rotación de
niveles por temporadas"). Solo se puede purgar lo que ya está archivado.

---

## 3. F10.6 — Poda de roles a `player` + `admin`

**Objetivo:** que el sistema tenga los dos roles que realmente existen.

| Archivo | Cambio |
|---|---|
| `backend/migrations/0002_roles_player_admin.up.sql` | `UPDATE accounts SET role='admin' WHERE role IN ('operator','director');` seguido de `ALTER TABLE accounts DROP CONSTRAINT ...; ADD CONSTRAINT ... CHECK (role IN ('player','admin'))` |
| `backend/migrations/0002_roles_player_admin.down.sql` | Restaura el CHECK de cuatro valores (no puede revertir el `UPDATE`; documentarlo en un comentario del propio archivo) |
| `backend/internal/domain/models.go` | Borrar `RoleOperator` y `RoleDirector` |
| `backend/internal/auth/`, `internal/transport/` | Compilar y corregir cualquier referencia; revisar los `switch` sobre rol |
| `frontend/src/features/auth/ProtectedRoute.tsx` | `allowedRoles` acepta solo `'admin' \| 'player'` |
| `frontend/src/features/admin-accounts/AdminAccountsPage.tsx` | El selector de rol al crear staff pierde las opciones que ya no existen |

**Verificación:** `go build ./... && go vet ./... && go test ./...` en verde;
migración aplicada contra base desechable; `INSERT` con `role='director'`
rechazado por el CHECK.

**Riesgo:** ninguno en producción — la única cuenta sembrada (`admin01`) es
`admin`. Aun así, el `UPDATE` va primero para que la migración no falle si
alguien creó una cuenta `operator` a mano.

---

## 4. F10.7 — Motor de juego y esquemas compartidos

**Objetivo:** tener en USBI-Anon los mismos motores y esquemas Zod que
`../usbi`, porque de ellos dependen tanto el maker como los minijuegos.

Se copian **verbatim** desde `../usbi/frontend/packages/`:

```text
frontend/packages/
  engine/          # lógica pura de los 7 juegos, sin React ni DOM
    src/games/     # Trivia, Memory, FakeNews, WordSearch, Puzzle, Crossword, Snakes
    src/interfaces/
  schema/          # esquemas Zod del content de cada plantilla
    index.ts
    snakes.ts
```

Y se replica el cableado de workspace: `pnpm-workspace.yaml` con
`frontend/packages/*`, `@usbi/engine` y `@usbi/schema` como
`workspace:*` en `frontend/package.json`, y el `build` del frontend
compilando el engine antes de `tsc --noEmit && vite build`.

**Punto de atención — nombre de los paquetes.** Se conservan como `@usbi/*`
aunque el proyecto sea `usbi-anon`: renombrarlos obligaría a tocar cada
`import` de los ~2 200 líneas que se van a copiar en F10.8–F10.10, a cambio de
cero beneficio funcional. Si el nombre final del producto se decide algún día
(decisión abierta n.º 2 del plan maestro), el renombre es un `sed` de un solo
paso y se hace entonces.

**Verificación:** las pruebas que ya vienen en `packages/engine/src`
(`TriviaEngine.test.ts`, `PuzzleEngine.test.ts`, `CrosswordEngine.test.ts`,
`WordSearchEngine.test.ts`) pasan; `tsc --noEmit` del frontend en verde.

---

## 5. F10.8 — Maker local (crear niveles sin ser admin)

**Objetivo:** recuperar `features/maker/MakerPage.tsx` — el maker *local*, que
es una cosa distinta del formulario de "agregar nivel" del panel de admin, y
por eso vive en su propia carpeta y tiene su propia ruta.

La diferencia, que conviene dejar escrita porque no es obvia:

| | Maker local (`/maker`) | Maker de admin (dentro de `/admin/content`) |
|---|---|---|
| Quién entra | Cualquier persona con sesión | Solo `admin` |
| Dónde guarda | `localStorage`, clave `usbi_local_levels` | Base de datos, vía `POST /levels` |
| Necesita sección | No | Sí, `section_id` obligatorio |
| Sirve para | Probar una idea, exportarla como JSON y pasársela a alguien | Publicar contenido oficial que da XP |

Archivos a portar desde `../usbi/frontend/src/`:

```text
features/maker/
  MakerPage.tsx            # 298 líneas: metadatos + formulario + preview + guardar/exportar
  index.ts
  MakerPage.test.tsx
features/content/
  types.ts                 # TemplateType, TEMPLATE_TYPE_LABELS, normalizadores de content
  maker/
    registry.ts            # tabla plantilla → {schema Zod, getDefaults, Form, Preview}
    LevelMetadataForm.tsx
    LevelActions.tsx
    forms/                 # 7 formularios
    previews/              # 7 previsualizadores
    snakesLayout.ts
```

Todo lleva ruta nueva `/maker` en `App.tsx`, dentro de `ProtectedRoute` sin
restricción de rol.

**Ajustes obligatorios respecto al original:**

1. **Quitar la rama Tauri de `onExport`.** El original intenta
   `window.__TAURI__` antes del fallback de navegador. USBI-Anon es web, no
   hay empaquetado de escritorio en ninguna fase: se conserva solo el camino
   de `Blob` + `<a download>`, y con eso desaparece la dependencia de
   `@tauri-apps/plugin-dialog` y `@tauri-apps/plugin-fs`.
2. **Los imports de `../../components/ui/` pasan a `../../shared/components/ui/`**,
   que es donde viven `Button`, `Input`, `Card` y `Brand` en este proyecto.
3. El campo `author` del maker local se queda como texto libre **del nivel, no
   de la persona**: es el nombre que se le pone a la autoría del contenido, y
   como el maker local no toca la base de datos, no introduce PII en ninguna
   tabla. En la importación al panel de admin se descarta (§7.2).

---

## 6. F10.9 — Panel de administración de contenido

**Objetivo:** `AdminContentPage`, la pantalla donde un admin gestiona el ciclo
de vida completo de secciones y niveles.

### 6.1 Frontend

Portar `features/content/AdminContentPage.tsx` (366 líneas) y
`content/maker/LevelMakerForm.tsx` (197). Ruta `/admin/content`, protegida con
`allowedRoles={['admin']}`. Lo que trae de fábrica:

- Crear sección (título, descripción, color), editarla, publicar/despublicar,
  archivar.
- Acordeón de niveles agrupados por sección.
- Crear nivel con el formulario por plantilla + previsualización.
- Editar nivel existente (carga el `content` en el formulario).
- Publicar/despublicar/archivar nivel.
- **Importar nivel desde JSON** (`handleImportCommunityLevel` +
  `<input type="file" accept=".json">`), que ya existe en el original y solo
  hay que traerlo.
- Enlace "Previsualizar" a `/levels/{id}/play`.

### 6.2 Backend: purga

Nuevos endpoints, ambos solo para `admin`:

```
DELETE /api/v1/levels/{level_id}      → 204
DELETE /api/v1/sections/{section_id}  → 204
```

Reglas, en el servicio, no en el handler:

1. **Solo purga lo archivado.** Si `archived_at IS NULL` → `409 Conflict` con
   `problem.detail` explicando que hay que archivar primero.
2. **Una sección con niveles vivos no se purga.** `levels.section_id` es
   `ON DELETE RESTRICT` a propósito: el servicio comprueba primero y devuelve
   `409` listando cuántos niveles quedan, en vez de dejar que Postgres tire un
   error de FK genérico.
3. **La XP sobrevive.** Antes de borrar, el servicio consolida los contadores
   en `account_retired_progress` y confía en las tres direcciones de FK ya
   definidas: `level_attempts` y `player_progress` en CASCADE se van,
   `experience_history` en SET NULL se queda con su `xp_gained` intacto.
   **No uniformar esas FK.**
4. Toda purga escribe en `admin_audit_log` con el UUID del admin, el UUID del
   contenido y el conteo de filas afectadas.

En la interfaz, doble confirmación: primero un aviso de que es irreversible y
de cuántos niveles/intentos se van, después escribir el título exacto del
nivel o sección para habilitar el botón. Nada de un `confirm()` del navegador.

### 6.3 Restaurar archivados

Como efecto secundario necesario de la purga: `POST /sections/{id}/unarchive` y
`POST /levels/{id}/unarchive`. Sin ellos, archivar por error solo se arregla
purgando, que es exactamente lo contrario de lo que se quiere.

---

## 7. Contrato JSON: exportar e importar deben ser la misma forma

Esta sección responde a una duda concreta y la respuesta es que **hoy no
coinciden**. Conviene arreglarlo al portar, no después.

### 7.1 Lo que exporta el maker

```jsonc
{
  "metadata": {
    "id": "uuid",
    "title": "string",
    "author": "string",
    "color": "#RRGGBB",
    "difficulty": 1,
    "template_type": "trivia",
    "creation_date": "2026-08-25T00:00:00.000Z"
  },
  "content": { /* forma dictada por el esquema Zod de la plantilla */ }
}
```

### 7.2 Lo que acepta la API

`POST /levels` espera un objeto plano: `section_id`, `title`, `color`,
`template_type`, `content`, `difficulty`. No tiene dónde poner `author`,
`creation_date` ni `id` — `levels` no tiene esas columnas, y `id` lo asigna el
servidor.

### 7.3 Regla que debe cumplir F10.9

- **Round-trip cerrado del maker:** exportar un nivel y volver a importarlo en
  el maker local produce un objeto idéntico campo por campo. Se prueba con un
  test por cada una de las siete plantillas.
- **Importación al panel de admin:** el archivo del maker se traduce al cuerpo
  de `POST /levels` descartando `author`, `creation_date` e `id`, y pidiendo la
  sección destino en la interfaz. La pérdida de esos tres campos es
  intencional y debe estar avisada en pantalla, no silenciosa.
- **Validación en tres capas, todas la misma:** el Zod de `@usbi/schema` en el
  navegador, `validateLevelInput` en Go, y el `CHECK` de `template_type` en
  Postgres. Si un JSON entra por importación, pasa por Zod **antes** de
  enviarse, para que el error se vea con el campo señalado y no como un `422`
  opaco.
- **Exportar desde el panel de admin:** un nivel ya guardado en la base se
  puede exportar al mismo formato del maker (rellenando `author` vacío y
  `creation_date` con `created_at`), para que el ciclo
  admin → archivo → maker → admin funcione en las dos direcciones.

---

## 8. F10.10 — Vista de jugador

**Objetivo:** que un jugador tenga a dónde entrar después del login.

Portar desde `../usbi/frontend/src/`:

```text
features/dashboard/DashboardPage.tsx     # catálogo de secciones, XP, racha
features/content/SectionLevelsPage.tsx   # niveles de una sección
features/content/OfficialLevelPage.tsx   # jugar un nivel oficial (da XP)
features/content/LocalLevelPage.tsx      # jugar un nivel del maker local (no da XP)
features/profile/ProfilePage.tsx         # progreso, insignias, historial de XP
features/games/                          # 7 componentes de juego + escenas Phaser
```

Rutas: `/dashboard`, `/sections/{id}`, `/levels/{id}/play`,
`/local-levels/{id}/play`, `/perfil`. Todas dentro de `ProtectedRoute`.

`HomePage` deja de ser una landing mínima: o redirige a `/dashboard`, o se
sustituye por él. Recomendación: `/` → `DashboardPage` y retirar `HomePage`,
para no tener dos pantallas peleándose por ser la primera.

**Sonido.** Aquí es donde deja de ser hipotético el campo `muteGameSounds` que
F10.5 dejó en `useSettingsStore` sin interfaz (decisión "opción a"). Al portar
los juegos: añadir la sección "Sonido" a `SettingsPage.tsx` y hacer que los
componentes de juego respeten el valor del store.

**Dependencia de Phaser.** `games/phaser/SnakeLadderScene.ts`,
`CrosswordScene.ts` y `WordSearchScene.ts` arrastran Phaser, que no es una
dependencia pequeña. Verificar el tamaño del bundle resultante y mantener el
`lazy()` por juego que ya usa `LocalLevelPage` — y comprobar que el contenedor
`usbi-anon`, con su límite duro de 500 MB de RAM, todavía puede correr
`npm run build` después de sumar Phaser. Si no puede, es un hallazgo legítimo
de esta fase y hay que reportarlo, no subir el límite en silencio.

---

## 9. F10.11 — Eliminar mi cuenta

**Backend.** `DELETE /api/v1/auth/me` deja de seudonimizar y pasa a borrado
duro, en una sola transacción, en `internal/privacy`:

1. Borrar respuestas del cuestionario (`PurgeAccountQuizAnswers`, ya existe).
2. Borrar progreso (`PurgeUserProgressData`, ya existe).
3. Revocar e invalidar los refresh tokens de la cuenta.
4. `DELETE FROM accounts WHERE id = $1`.
5. Escribir en la bitácora de auditoría el UUID, la marca de tiempo y el
   motivo `self_deletion`, **antes** del paso 4 si alguna FK lo exige.

Revisar cada FK que apunte a `accounts(id)` y decidir explícitamente su
dirección — `created_by_admin_id` ya es `ON DELETE SET NULL`, que es correcto:
borrar a un admin no debe borrar el contenido que creó.

**Frontend.** Sección "Eliminar mi cuenta" en `SettingsPage` o en `ProfilePage`
(preferible: perfil, para no mezclarla con ajustes de accesibilidad), con:
aviso de irreversibilidad, lista de lo que se pierde, confirmación escribiendo
el propio nickname, y cierre de sesión con redirección a `/login` al terminar.

**Lo que NO se trae:** las pantallas de derechos ARCO (`features/arco/`) ni
sus pestañas de administración, por instrucción explícita.

---

## 10. F10.12 — Navegación: volver y dashboard en todas partes

**Objetivo:** que ninguna pantalla sea un callejón sin salida. Dos fallos
concretos que arrastra `../usbi` y que no se deben replicar:

- Varias pantallas no tienen forma de volver.
- Desde la vista de jugador no se puede regresar a la vista de administrador:
  un admin que entra a previsualizar un nivel queda atrapado ahí.

Solución: un componente compartido, `shared/components/NavBack.tsx`, con dos
piezas siempre presentes en el encabezado de cada página:

1. **Volver** — `navigate(-1)`, con etiqueta que diga a dónde vuelve cuando se
   sepa (`← Volver a la sección`), no un genérico.
2. **Ir al inicio** — enlace directo a `/dashboard` para jugador, y a
   `/admin/content` cuando la sesión es `admin` y se llegó desde el panel.

Reglas de aceptación de la fase:

- Toda ruta que no sea `/login`, `/register` ni `/dashboard` tiene visible al
  menos un camino de salida sin usar el botón del navegador.
- Desde `/levels/{id}/play`, un `admin` ve además "Volver al panel de
  contenido".
- Los dos controles cumplen el área mínima de 44×44 px y tienen `aria-label`
  propio cuando son solo icono.
- Se pasa por `/maker`, `/settings`, `/perfil`, `/admin/content`,
  `/admin/accounts` y `/admin/registration-questions` verificando uno por uno.

---

## 11. Orden y dependencias

```
F10.6 (roles) ──────────────────────► independiente, primera por ser la más corta
F10.7 (engine + schema) ──┬──► F10.8 (maker local)
                          └──► F10.9 (admin contenido) ──► F10.10 (vista jugador)
F10.11 (borrar cuenta) ─────────────► independiente
F10.12 (navegación) ────────────────► al final: necesita que existan las pantallas
F11 (legal) ────────────────────────► sigue siendo la última, sin dependencias
```

F10.6 y F10.11 tocan backend y no dependen de nada del frontend: se pueden
adelantar o paralelizar. F10.12 tiene que ir al final por construcción.

Cada fase cierra en su propio commit con mensaje en español y con su entrada
en `estado_proyecto.md`, igual que F5–F10.5.

---

## 12. Verificación por fase

| Fase | Cómo se comprueba |
|---|---|
| F10.6 | `go build/vet/test ./...`; migración contra base desechable; `INSERT` con `role='director'` rechazado |
| F10.7 | Pruebas del engine en verde; `tsc --noEmit` del frontend limpio |
| F10.8 | Crear un nivel de cada plantilla, exportarlo, reimportarlo y comparar el JSON campo por campo (7 casos) |
| F10.9 | Crear sección → nivel → publicar → archivar → purgar contra backend real; verificar en SQL que `experience_history.xp_gained` sobrevive y que `account_retired_progress` conserva los contadores |
| F10.10 | Jugar y completar un nivel de cada plantilla; `POST /levels/{id}/complete` otorga la XP esperada; `npm run build` completa dentro del límite de 500 MB del contenedor |
| F10.11 | Borrar una cuenta de prueba; confirmar en SQL que no quedan filas en `accounts`, `account_quiz_answers` ni progreso, y que sí queda el registro de auditoría |
| F10.12 | Recorrido manual por las nueve rutas; `eslint` sin avisos de accesibilidad |

Todo lo de backend, contra base desechable — nunca contra `usbi-database`
persistente.

---

## 13. Lo que este bloque NO incluye

- **Derechos ARCO y sus pestañas** — excluidos por instrucción explícita.
- **Empaquetado de escritorio (Tauri)** — se retira al portar el maker.
- **Renombrar `@usbi/*`** — depende de la decisión abierta del nombre final.
- **Insignias y rachas nuevas** — se porta lo que `../usbi` ya tenía, no se
  diseñan mecánicas nuevas.
- **La reescritura legal (F11)** — sigue siendo la fase siguiente y última.
