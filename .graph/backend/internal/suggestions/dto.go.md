---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define las estructuras de datos (DTOs) necesarias para la transferencia de información en el envío y consulta del buzón de sugerencias anónimas. Garantiza el principio de anonimato al no incorporar ningún campo de identidad ni de relación con el usuario en los modelos de solicitud o respuesta. `toResponse` proyecta el tipo generado por sqlc [[backend/internal/repository/interest_link_queries.go.md|repository.Suggestion]] hacia `SuggestionResponse`, usado por admin_service.go#AdminService.List.
