---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa la gestión de migraciones de la base de datos mediante el subcomando `migrate`. Conecta exclusivamente con el rol dedicado `usbi_migrate` utilizando las migraciones embebidas en el binario, evitando la ejecución automática al iniciar el servidor para prevenir condiciones de carrera y permitir la inspección previa.

## Funciones

### newMigrator
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicializa el controlador de migraciones desde el sistema de archivos embebido y establece la conexión a la base de datos utilizando el usuario de migración.

### runMigrate
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Interpreta las acciones de migración (`up`, `down`, `force`, `version`), requiriendo que la reversión (`down`) especifique el número de pasos para evitar reversiones accidentales completas.

## Relaciones

- [[backend/cmd/usbictl/main.go|usbictl]]: Subcomando montado desde main
- [[backend/migrations/embed.go|migrations]]: Empotra los archivos SQL de migración
- [[backend/sql/00_roles_unificado.sql|roles]]: Define el rol `usbi_migrate` utilizado para las migraciones
