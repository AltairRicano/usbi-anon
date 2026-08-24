# 03 — Frontend: los 6 archivos que cambian

Documento de detalle de [`00_Plan_maestro.md`](00_Plan_maestro.md). Cubre la
fase **F5**. Es, con diferencia, la parte que menos cambia del proyecto: el
frontend nunca supo que existía una sola base de datos, así que partirla en dos
le es invisible. Lo único que le llega es que **el DTO de usuario pierde
`full_name`**.

## 1. Alcance verificado

`../usbi/frontend/src` tiene **67 archivos** TypeScript/TSX. Un `grep` de
`full_name|fullName|email|phone|tutor` sobre todo `src/` toca **7 archivos**,
de los cuales uno es puramente legal. El resto del frontend —motores de juego,
Phaser, el Maker de niveles, componentes UI, Tauri, el store de sync offline—
**no se toca**.

## 2. Los archivos que cambian

### 2.1 `src/lib/schema.ts` — contrato Zod

- `RegisterSchema`: eliminar `full_name` (línea 4) y `phone` (línea 6).
  `phone` **ya estaba muerto**: `RegisterPage.tsx` nunca lo envía (verificado
  con `grep -n 'phone' features/auth/RegisterPage.tsx` → sin resultados).
- `UserSchema`: eliminar `full_name` (línea 63).
- `TutorConsentSchema`, `LoginSchema`, `ArcoSchema`, y todos los esquemas de
  progreso/sync: **sin cambios**. El correo del tutor y el de login siguen
  existiendo; lo que cambia es a qué base van a parar, y eso el cliente no lo
  ve.

### 2.2 `src/stores/useAuthStore.ts` — estado de sesión

Eliminar `full_name: string` de la interfaz `User` (línea 16). El comentario de
esa línea (*"el backend retorna full_name, no name"*) también se va.

El comentario del bloque (*"NUNCA incluye password_hash, email cifrado, ni
material criptográfico"*) conviene ampliarlo: en USBI-Anon el DTO **tampoco
incluye ningún dato identificable**, solo el UUID.

La persistencia en `sessionStorage` (no `localStorage`, por XSS) **se conserva
tal cual**: sigue siendo la decisión correcta.

### 2.3 `src/features/auth/RegisterPage.tsx` — formulario de alta

- Eliminar el campo *"Nombre completo"* (estado `fullName`, líneas 40 y 150) y
  su envío en el `POST /auth/register` (línea 74).
- **Conservar** el bloque de tutor (`tutorName`, `tutorEmail`, líneas 47–48 y
  218–229) y el `POST /auth/tutor-consent`. Esos datos siguen siendo
  necesarios y legítimos: van a la base de identidad, no a la principal.
- Revisar el texto de aceptación del aviso de privacidad, que hoy asume que se
  recaba nombre.

### 2.4 `src/features/dashboard/DashboardPage.tsx` — el saludo

Línea 418: `Bienvenido, {user?.full_name}`. Este es el **único punto del
frontend que rompe visiblemente** al quitar `full_name`.

**Decisión cerrada (§4.2 del plan maestro): alias generado por el sistema.**
La línea queda `Bienvenido, {user?.display_alias}` y el backend devuelve la
cadena ya compuesta (p. ej. *"Jaguar Azul 42"*) desde la vista
`account_aliases` de la base principal.

En `UserSchema` y en el store, `full_name: z.string()` se sustituye por
`display_alias: z.string()`. **El alias no es un identificador**: no es único y
jamás debe usarse como clave de búsqueda ni mostrarse como si lo fuera — la
llave siempre es el UUID.

### 2.5 `src/features/profile/ProfilePage.tsx` y `src/features/arco/ArcoPage.tsx`

Solo textos. `ProfilePage.tsx:42` y `:71` hablan de *"minimizar los datos del
tutor"*; sigue siendo cierto, pero conviene precisar que esos datos viven en un
sistema separado. `ArcoPage.tsx` debe reflejar que una cancelación ahora
coordina dos sistemas (ver [`02_Backend.md` §5](02_Backend.md)) — sin exponer
detalle técnico a la persona usuaria.

### 2.6 `src/features/legal/PrivacyPage.tsx` — **bloqueado por F6**

Es el archivo con el cambio más grande y el único que **no debe tocarse en
F5**. Hoy afirma literalmente que se recaban *"nombre, correo electrónico,
teléfono y, en caso de ser menor de edad, el nombre y correo de su tutor
legal"* y describe alojamiento *On-Premise* en una sola base institucional.

En USBI-Anon eso deja de ser exacto en tres puntos: ya no se recaba nombre ni
teléfono, los datos identificables viven en un sistema separado del progreso, y
el modelo de alojamiento cambia (la UV pasa de custodiar respaldos a operar
infraestructura de producción — ver el resumen de cambios legales en
`CLAUDE.md`).

**Un aviso de privacidad que describe mal el tratamiento es un incumplimiento,
no un texto desactualizado.** Este archivo se reescribe en **F6**, junto con
Convenio/EIPDP/Documento de seguridad/Condiciones/Diccionario de datos, y con
validación del área jurídica competente. Skill: agente
**`legal-compliance-checker`**.

## 3. Qué NO se toca (y por qué)

| Área | Motivo |
|---|---|
| `features/games/`, `features/content/maker/`, `packages/engine` | Lógica de minijuegos. Nunca vio identidad |
| `lib/localSyncQueue.ts` (SQLite local vía `@tauri-apps/plugin-sql`) | Encola progreso por `user_id` UUID. Ya era anónimo. **Verificar en F5** que ningún campo del payload local guarde nombre o correo |
| `components/ui/`, `styles/`, identidad visual UV | Sin cambios. Siguen aplicando `plan/Convenciones_de_color_UV.md` |
| `src-tauri/` | Sin cambios |
| Flujo de sync, `wipe_local_data`, HMAC | Idéntico. `SyncPayload` ya está documentado en Go como "MUST NOT contain PII" |

## 4. Mejoras propuestas y con qué skill se atacan

| # | Mejora | Skill / agente | Cuándo |
|---|---|---|---|
| 1 | Rediseñar el encabezado del dashboard sin nombre propio sin que pierda calidez | **`ui-ux-pro-max`** con los guardrails UV de `SKILLS.md` §3 | F5, tras cerrar §4.2 |
| 2 | Verificar que quitar el campo nombre no rompe etiquetas, foco ni lectura por lector de pantalla en el formulario de registro | agente **`accessibility-auditor`** | F5 |
| 3 | Reescritura del aviso de privacidad y del resto de documentos legales | agente **`legal-compliance-checker`** | F6 (bloquea §2.6) |
| 4 | Revisión del diff de frontend | **`/code-review`** | Cierre de F5 |

> Regla 5 del `CLAUDE.md` global: cualquier agente que se lance debe escribir su
> reporte en su propia sección de `estado_proyecto.md`.

## 5. Criterios de aceptación de F5

1. `grep -rn 'full_name\|fullName' frontend/src/` no devuelve nada.
2. `grep -rn 'phone' frontend/src/lib/schema.ts` no devuelve nada.
3. El registro de un menor sigue funcionando de punta a punta, incluido el
   doble opt-in por correo del tutor.
4. `pnpm build` (o sea `tsc --noEmit` + `vite build`) pasa limpio.
5. `PrivacyPage.tsx` **sigue sin tocarse** y queda anotado como deuda de F6.
6. Se compila **fuera de los contenedores** (regla de RAM heredada de
   `../usbi/CLAUDE.md`) y solo se copia `dist/` al destino.
