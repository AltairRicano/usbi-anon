---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa la lógica de negocio para la lectura de registros de auditoría por parte de administradores. Garantiza la trazabilidad forense al registrar automáticamente en la misma transacción cada consulta realizada (`audit_log.read`) conservando los filtros aplicados sin almacenar los datos devueltos, e implementa paginación por cursor opaco basado en marca de tiempo y UUID.

## Funciones

### AdminService.List
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Ejecuta la consulta paginada de registros de auditoría aplicando los filtros provistos, llamando a [[backend/internal/repository/content_queries.go.md#Queries.ListAuditLog|Queries.ListAuditLog]]. Inicia una transacción de base de datos, consulta `pageSize + 1` filas para evaluar la existencia de una página siguiente, genera el cursor opaco correspondiente, traduce cada fila con [[backend/internal/auditlog/dto.go.md|dto.go]] y registra de forma atómica la entrada de auditoría `audit_log.read` con [[backend/internal/audit/audit.go.md#Log|audit.Log]] (los filtros utilizados, nunca los resultados) antes de realizar el commit. Es el único punto de entrada que llama Handler.List.

### parseCursor
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Decodifica la cadena de cursor opaco con formato `<RFC3339Nano>_<UUID>` obteniendo la marca de tiempo `time.Time` y el UUID correspondiente para la paginación. Retorna `ErrValidation` si la sintaxis es inválida.

### encodeCursor
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Construye la cadena de cursor opaco concatenando la marca de tiempo formateada en RFC3339Nano y el UUID del registro.

### auditReadFilters
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera un mapa con los parámetros de filtro aplicados en la consulta para ser registrado en el log de auditoría de lecturas privileged, omitiendo deliberadamente los datos devueltos en la respuesta.

### parsePageSize
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Parsea la cadena del parámetro `page_size` retornando el valor por defecto (20) si está vacía, o validando que el entero resultante esté dentro del rango permitido (entre 1 y 50).
