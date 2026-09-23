---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Provee la lógica de negocio enfocada en el jugador final, garantizando que nunca se exponga contenido borrador o no publicado. Gestiona la consulta de catálogo publicado, el procesamiento transaccional serializable al completar niveles (cálculo de XP, reintentos concurrentes, rachas e insignias) y la obtención del progreso del perfil.

## Funciones

### PlayerService.GetLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene un nivel por ID retornando error de no encontrado si el nivel no está publicado.

### PlayerService.ListLevels
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Consulta de forma paginada por cursor los niveles limitándose estrictamente a los publicados.

### PlayerService.ListSections
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Lista únicamente las secciones que se encuentran publicadas.

### PlayerService.CompleteLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Procesa en una transacción serializable la verificación de respuestas con [[backend/internal/levels/verify.go.md#verifyAnswers|verifyAnswers]], el cálculo de XP otorgado según el intento con [[backend/internal/levels/service.go.md#CalculateXP|CalculateXP]], registro de intento, actualización de racha diaria con [[backend/internal/levels/service.go.md#calculateCurrentStreak|calculateCurrentStreak]] y concesión de insignias elegibles.

### PlayerService.GetProfileProgress
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Consolida y retorna el resumen de progreso del jugador, incluyendo XP total, niveles completados, racha actual, insignias obtenidas y estado por nivel.
