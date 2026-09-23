---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona el subcomando `doctor` para auditar la seguridad y matriz de permisos en los tres pools de la base de datos (`usbi_app`, `usbi_moderador`, `usbi_dbmaint`). Evalúa tanto operaciones permitidas como prohibidas dentro de transacciones con reversión obligatoria para garantizar la ausencia de privilegios indebidos sin alterar datos.

## Funciones

### newSuggestionArgs
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera una lista de argumentos de prueba con un UUID fresco para la inserción temporal de filas en la tabla `suggestions`.

### runDoctor
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Establece la lista de chequeos para cada pool de base de datos, ejecuta la suite de comprobaciones y fuerza un estado de salida fallido si detecta algún fallo de permisos o una prohibición no cumplida.

### runPoolCheck
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Se conecta al pool configurado, verifica los usuarios de sesión y ejecuta secuencialmente las consultas SQL permitidas y prohibidas registrando los hallazgos.

### execInRolledBackTx
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicia una transacción en la base de datos, ejecuta la sentencia especificada y garantiza que la transacción siempre termine en un `Rollback` para no persistir datos de prueba.

## Relaciones

- [[backend/cmd/usbictl/main.go|usbictl]]: Subcomando montado desde main
- [[backend/sql/00_roles_unificado.sql|roles]]: Define los roles y GRANT verificados por doctor
