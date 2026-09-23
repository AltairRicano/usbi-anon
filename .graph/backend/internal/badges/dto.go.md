---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define los objetos de transferencia de datos (DTO) para la serialización y deserialización JSON en los endpoints de administración de insignias. Incluye las estructuras de petición de creación y actualización, así como la respuesta envuelta en la propiedad `items`. `toResponse` traduce el modelo `repository.Badge` (definido en [[backend/internal/repository/badge_queries.go.md|badge_queries.go]]) al DTO público; lo consume [[backend/internal/badges/admin_service.go.md|admin_service.go]].
