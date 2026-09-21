# Despliegue de USBI-Anon

Guía operativa para desplegar los tres servicios (`db`, `api`, `web`) desde
cero en un servidor nuevo, con `docker-compose.yml` (M3 del plan de
maduración, `plan/00_Plan_de_maduracion.md`). Escrita mientras se ejecutaba
el ensayo real de esta misma secuencia — no es un procedimiento teórico:
cada paso, incluidos los dos bugs que se mencionan explícitamente, se
encontró y corrigió corriendo esto de verdad.

**Estado de las cuatro decisiones abiertas que afectan este documento:**
tamaño del VPS ya confirmado (1 vCPU, 1 GB RAM, 20 GB disco — los valores de
memoria de Postgres más abajo están calculados para eso). Nombre/dominio
definitivo, área de la UV responsable y ubicación del centro de datos
siguen sin decidirse; por eso este documento no incluye TLS todavía (sección
"Pendiente antes de exponer a internet real").

## Prerrequisitos

- Docker Engine con el plugin `docker compose` (v2).
- Git.
- El repositorio clonado en el servidor:
  ```
  git clone <url-del-repositorio> usbi-anon
  cd usbi-anon
  ```

No hace falta Go ni Node instalados en el servidor: todo se compila dentro
de los propios contenedores durante `docker compose build`.

## 1. Generar `usbictl` (una sola vez, antes de que exista el stack)

`usbictl secrets init` es lo primero que hay que correr, pero vive dentro de
la imagen `api` — que todavía no existe la primera vez. Se compila aparte,
una sola vez, con un contenedor Go temporal:

```sh
docker run --rm \
  -v "$(pwd)/backend:/src:ro" \
  -v /tmp:/out \
  -w /src \
  golang:1.22-bookworm \
  sh -c "go build -o /out/usbictl ./cmd/usbictl"
```

Esto deja el binario en `/tmp/usbictl` (fuera del repositorio: es una
herramienta de operación, no código versionado como artefacto compilado).
Para las corridas posteriores (`migrate`, `admin create`, `doctor`) no hace
falta repetir esto — se usan a través de `docker compose run` contra la
imagen `api` ya construida (ver más abajo).

## 2. Generar los secretos

```sh
/tmp/usbictl secrets init --out ./.env
```

Pide por consola **una sola cosa**: la contraseña del primer administrador
(o léela de `ADMIN_BOOTSTRAP_PASSWORD` en el entorno si prefieres no
teclearla interactivamente, por ejemplo en un script no interactivo:
`ADMIN_BOOTSTRAP_PASSWORD='...' /tmp/usbictl secrets init --out ./.env`).

Genera y escribe en `.env` (con permisos 600, gitignored):

- `JWT_SECRET`, `HMAC_SECRET`
- `DB_SUPERUSER_PASSWORD` (superusuario `postgres`, solo para el bootstrap
  del clúster — nunca lo usa el backend)
- `DB_APP_PASSWORD`, `DB_MODERADOR_PASSWORD`, `DB_MIGRATE_PASSWORD`,
  `DB_DBMAINT_PASSWORD` (los cuatro roles de aplicación)

Es idempotente: si `.env` ya existe, no lo pisa. Para rotar una sola clave:
`usbictl secrets init --rotate <CLAVE> --out ./.env`.

**Nunca versionar este archivo.** `.gitignore` ya excluye `.env`/`.env.*`
salvo `.env.example`.

## 3. Configurar el resto de `.env`

Copia los valores no secretos de `.env.example` a tu `.env` (puertos,
`CORS_ALLOWED_ORIGIN`, red interna, memoria de Postgres) y ajústalos si
hace falta. Con los valores por defecto no hace falta tocar nada más — son
los mismos puertos que ya se usan en desarrollo (8092/8093/5435).

## 4. Renderizar la configuración de Postgres

```sh
/tmp/usbictl pgconf render \
  --total-ram-mb 1024 \
  --network 172.28.0.0/16 \
  --out-dir ./backend/deploy/postgres/rendered
```

Genera `pg_hba.conf` (scram-sha-256, restringido a la red interna del
compose) y `postgresql.conf` (este segundo no lo usa el compose — los
valores de memoria van como flags `-c` en `docker-compose.yml`, ya resueltos
desde `.env`; queda disponible para un despliegue sin compose). Ajusta
`--total-ram-mb` en cuanto cambie el tamaño real del VPS contratado.

## 5. Levantar el stack

```sh
docker compose up -d --build
```

Construye las dos imágenes (`api` multi-etapa Go, `web` multi-etapa
Node→nginx) y arranca los tres servicios. `db` tiene `healthcheck`; `api`
espera a que esté sano antes de arrancar.

**Límite de memoria por contenedor.** Los tres servicios traen
`mem_limit: 250m` en `docker-compose.yml` — 750 MB combinados en ejecución,
con margen dentro del VPS de referencia de 1 GB. Docker Compose no tiene
forma nativa de compartir un único tope entre varios contenedores (cada uno
es su propio cgroup), así que es un límite por servicio, no un pool
conjunto. **Este límite aplica solo en ejecución, no durante el build**: el
paso `--build` de este mismo comando compila el frontend con Vite dentro
del multi-stage de `web`, y ese build por sí solo puede picar a 600–700 MB
(el mismo patrón que ya obligó, en el contenedor de desarrollo de un solo
servicio, a generar `dist/` en el host — ver limitaciones conocidas en
`CLAUDE.md`). En un VPS de 1 GB total, construir las imágenes ahí mismo
sigue siendo un punto de riesgo de OOM aunque los contenedores ya corran
acotados a 250 MB cada uno; si el build falla por memoria, la alternativa
es construir las imágenes en otra máquina y subirlas (`docker save`/`load`
o un registro) en vez de construir en el VPS.

**La primera vez que arranca `db`** (volumen de datos vacío), la imagen
oficial de Postgres corre automáticamente
`backend/sql/00_init_cluster.sh` vía `docker-entrypoint-initdb.d`: crea los
cuatro roles con las contraseñas de `.env`, los une (`usbi_migrate` como
miembro de `usbi_moderador`, necesario para la migración 0004) y crea
`usbi_anon_db`. Esto **no vuelve a correr** en arranques posteriores del
mismo volumen — es una operación de clúster, no una migración.

Verifica que los tres servicios estén arriba:

```sh
docker compose ps
docker compose logs api --tail 20
```

Si `api` muestra `password authentication failed`, el volumen de `db`
quedó con contraseñas de una corrida anterior con un `.env` distinto —
bájalo con `docker compose down -v` (esto borra los datos de esa base:
solo hacerlo en un entorno todavía sin datos reales) y vuelve a levantar.

## 6. Aplicar las migraciones

El esquema no se aplica solo al arrancar (M3.3: subcomando explícito, nunca
automático). Con los tres servicios ya arriba:

```sh
docker compose run --rm --entrypoint /usbictl api migrate up
```

Corre como el rol `usbi_migrate` (dueño del esquema), nunca como
superusuario. Para ver en qué versión quedó la base:

```sh
docker compose run --rm --entrypoint /usbictl api migrate version
```

## 7. Crear el primer administrador

```sh
docker compose run --rm --entrypoint /usbictl \
  -e ADMIN_BOOTSTRAP_PASSWORD="$(grep ADMIN_BOOTSTRAP_PASSWORD .env | cut -d= -f2)" \
  api admin create --nickname admin01
```

Llama exactamente al mismo código que `POST /admin/accounts` — la
validación de nickname/password nunca puede divergir entre este bootstrap
y el uso normal de la aplicación.

## 8. Verificar que las credenciales realmente sirven

```sh
docker compose run --rm --entrypoint /usbictl api doctor
```

Confirma, para los tres pools (`usbi_app`, `usbi_moderador`,
`usbi_dbmaint`): identidad (`current_user`/`session_user`), una operación
que sí debe funcionar, y una que debe fallar. Si algo que debería estar
prohibido tiene éxito, `doctor` termina con código de salida distinto de
cero — no lo ignores, es la señal de un `GRANT` de más.

## 9. Verificación final

```sh
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:${HOST_WEB_PORT:-8092}/
curl -s -X POST http://localhost:${HOST_WEB_PORT:-8092}/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"nickname":"admin01","password":"<la que elegiste en el paso 2>"}'
```

El primer comando debe responder `200`; el segundo debe devolver un
`access_token`. Si ambos funcionan, el stack está operativo de punta a
punta: nginx sirviendo el frontend, proxy de `/api` funcionando, y los tres
pools de Postgres autenticando correctamente.

## Actualizar un despliegue existente

```sh
git pull
docker compose up -d --build         # reconstruye api/web si el código cambió
docker compose run --rm --entrypoint /usbictl api migrate up   # si hay migraciones nuevas
```

`db` no necesita reconstruirse (usa la imagen oficial `postgres:15` sin
cambios); su configuración (`pg_hba.conf`, memoria) solo cambia si vuelves
a correr `usbictl pgconf render` y reinicias ese servicio.

## Solución de problemas

| Síntoma | Causa probable | Qué hacer |
|---|---|---|
| `api` en `Restarting`, log `password authentication failed` | El volumen de `db` tiene contraseñas de un `.env` anterior | `docker compose down -v` (⚠️ borra los datos) y volver a levantar con el `.env` actual |
| `api` en `Restarting`, log `signal: killed` en `go build` o en Vite | Memoria insuficiente durante el build — no debería pasar en este flujo (los binarios ya vienen compilados en la imagen), señal de que algo está corriendo `go run`/`vite dev` en vez de los binarios de producción | Revisar que no haya un `entrypoint.sh` viejo parchado a mano en la imagen (ver nota histórica en `estado_proyecto.md`, incidente 2026-09-17) |
| `doctor` reporta una operación prohibida con éxito | Un `GRANT` de más, probablemente por una migración nueva que no revisó la matriz de permisos | Revisar `backend/migrations/0008_matriz_permisos.up.sql` y la migración nueva; nunca ampliar un `GRANT` sin revisar `doctor` después |
| `migrate up` falla en la migración 0004 con `must be member of role usbi_moderador` | El volumen de `db` es de antes de que `00_roles_unificado.sql` incluyera el `GRANT usbi_moderador TO usbi_migrate` | Reinicializar el volumen (`docker compose down -v`) — este archivo ya está corregido en el repositorio, el error solo aparece contra un volumen viejo |

## Pendiente antes de exponer a internet real

Estos puntos son deliberadamente distintos de lo que hace este documento
hoy — no son errores, son las piezas que dependen de decisiones todavía
abiertas (ver `CLAUDE.md`, sección "Decisiones cerradas y pendientes"):

- **TLS.** `frontend/deploy/nginx.conf` sirve hoy en HTTP plano
  (`server_name _`, puerto 80). En cuanto exista un dominio definitivo,
  agregar un `server{}` en 443 con certificado y redirigir el bloque de 80.
- **Cerrar los puertos de `db` y `api`.** Hoy ambos están publicados al
  host (`HOST_DB_PORT`, `HOST_API_PORT`) porque así se usa en desarrollo.
  Antes de una IP pública real, comenta esos dos bloques `ports` en
  `docker-compose.yml` (o pon los valores a nada) — nada dentro de la red
  `internal` del compose necesita que salgan del host; el frontend ya
  proxya `/api` sobre el mismo origen.
- **Respaldos.** `pg_dump` programado, política de retención, y una
  restauración de prueba ya ejecutada — fuera de alcance de M3 (ver M3.11
  en el plan de maduración).
- **Renovación automática de TLS**, una vez exista el certificado inicial.
- **Bitácoras y monitoreo** de espacio en disco — el VPS de referencia
  tiene 20 GB totales, compartidos entre el volumen de Postgres y las
  imágenes/logs de Docker.
- **CORS_ALLOWED_ORIGIN** debe apuntar al dominio público real, no a
  `localhost`.
