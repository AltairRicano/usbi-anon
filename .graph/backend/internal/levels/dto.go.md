---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define los objetos de transferencia de datos (DTOs) utilizados en las solicitudes y respuestas de la API de niveles, secciones, progreso del jugador, historial e insignias. Incluye además el mapa constante `AllowedTemplateTypes` con las plantillas de juego soportadas, que usa `validateLevelInput` en [[backend/internal/levels/service.go.md|service.go]] para rechazar tipos de plantilla desconocidos.

Estos DTOs son la forma de entrada/salida de [[backend/internal/levels/admin_service.go.md|admin_service.go]], [[backend/internal/levels/player_service.go.md|player_service.go]] y [[backend/internal/levels/handler.go.md|handler.go]].
