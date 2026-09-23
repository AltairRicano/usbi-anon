---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define las estructuras de datos DTO, constantes de límites del dominio y errores estandarizados compartidos por [[backend/internal/quiz/admin_service.go.md|admin_service.go]] y [[backend/internal/quiz/player_service.go.md|player_service.go]]. Incluye la función de muestreo aleatorio no ponderado para la selección de preguntas de registro y `logAudit`, que envuelve [[backend/internal/audit/audit.go.md#Log|audit.Log]] para las mutaciones del banco de preguntas.

## Funciones

### selectRandom
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Efectúa un barajado aleatorio completo (shuffle) sobre el conjunto de preguntas activas empleando `math/rand` y retorna un subconjunto delimitado por el máximo solicitado. La invoca [[backend/internal/quiz/player_service.go.md|PlayerService]] al armar el cuestionario de registro.
