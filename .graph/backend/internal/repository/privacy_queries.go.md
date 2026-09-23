---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Gestiona operaciones relativas a la privacidad del usuario y derechos ARCO. Implementa la anonimización de bitácoras de eventos mediante procedimientos de la base de datos para no-repudio, la cancelación de cuentas con seudonimización de apodo y la eliminación selectiva del progreso personal. Único consumidor: [[backend/internal/privacy/privacy.go.md|privacy/privacy.go]], que orquesta estas cuatro consultas dentro de la transacción de cancelación de cuenta.

## Funciones

### Queries.NullUserInPseudonymizableLedgers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Ejecuta una función en la base de datos que desvincula el ID del usuario en bitácoras de auditoría conservando las filas para fines de no-repudio.

### Queries.DeactivateAccount
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Desactiva una cuenta reemplazando el apodo por un texto aleatorio para permitir su reutilización, marcando su estado como eliminado e incrementando la versión del token de autenticación.

### Queries.PurgeAccountQuizAnswers
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina las respuestas del cuestionario del usuario llamando al procedimiento con privilegios `SECURITY DEFINER` `purge_account_quiz_answers` dentro de la transacción de cancelación.

### Queries.PurgeUserProgressData
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina los datos de progreso personal del usuario de las tablas de progreso, intentos, rachas y medallas, excluyendo las bitácoras de auditoría.
