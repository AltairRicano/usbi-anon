---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Contiene las declaraciones de errores de dominio del paquete de insignias y las reglas de validación de entradas. Define límites para nombres, exige umbrales de XP no negativos y valida que las claves de icono cumplan con un patrón regex específico (`^[a-z][a-z0-9_]{2,39}$`) para su mapeo dinámico en el cliente. Lo consumen [[backend/internal/badges/admin_service.go.md#AdminService.Create|AdminService.Create]] y [[backend/internal/badges/admin_service.go.md#AdminService.Update|AdminService.Update]]; la concesión automática de insignias por umbral de XP y la lectura del jugador viven en `internal/levels` en vez de aquí.

## Funciones

### validateBadgeInput
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida que el nombre no esté vacío ni exceda 100 caracteres, que el umbral de XP no sea negativo y que la clave del icono coincida con la expresión regular `^[a-z][a-z0-9_]{2,39}$`.
