---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define las estructuras de datos (DTOs) para las solicitudes y respuestas JSON de categorías y enlaces de interés consumidas por los endpoints de jugador y administración. Asegura la consistencia del contrato API envolviendo los listados en la clave `items` e inicializando los slices de enlaces para evitar nulos en las respuestas hacia el frontend.

`categoryToResponse` y `linkToResponse` convierten los modelos `repository.InterestLinkCategory`/`repository.InterestLink` (ver [[backend/internal/repository/interest_link_queries.go.md|interest_link_queries.go]]) a estos DTOs; los consumen tanto [[backend/internal/interestlinks/admin_service.go.md|admin_service.go]] como [[backend/internal/interestlinks/player_service.go.md#PlayerService.ListGrouped|PlayerService.ListGrouped]].
