#!/bin/bash
# Ejecutado UNA SOLA VEZ por la imagen oficial de Postgres, al inicializar un
# volumen de datos vacío (docker-entrypoint-initdb.d) — nunca en arranques
# posteriores del contenedor. Crear roles y la base es una operación de
# clúster (M3.3), no una migración: golang-migrate (usbictl migrate) parte
# de que usbi_anon_db y sus cuatro roles ya existen.
#
# Traduce las contraseñas generadas por `usbictl secrets init` (variables de
# entorno del servicio `db` en docker-compose.yml) a los parámetros psql que
# 00_roles_unificado.sql espera. El .sql vive montado con extensión .sql.src
# (no .sql) para que el escáner automático de la imagen no intente
# ejecutarlo también por su cuenta.
set -euo pipefail

: "${DB_APP_PASSWORD:?falta DB_APP_PASSWORD — usbictl secrets init lo genera}"
: "${DB_MODERADOR_PASSWORD:?falta DB_MODERADOR_PASSWORD}"
: "${DB_MIGRATE_PASSWORD:?falta DB_MIGRATE_PASSWORD}"
: "${DB_DBMAINT_PASSWORD:?falta DB_DBMAINT_PASSWORD}"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres \
  -v app_password="'${DB_APP_PASSWORD}'" \
  -v moderador_password="'${DB_MODERADOR_PASSWORD}'" \
  -v migrate_password="'${DB_MIGRATE_PASSWORD}'" \
  -v dbmaint_password="'${DB_DBMAINT_PASSWORD}'" \
  -f /docker-entrypoint-initdb.d/00_roles_unificado.sql.src
