---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Documento histórico de arquitectura del backend en Go bajo el modelo de dos pools de conexión y dos bases de datos. Clasifica la reutilización de código (verbatim vs reescritura), detalla la separación entre `internal/repository` y el paquete `internal/identityrepo`, y especifica la orquestación idempotente de la saga ARCO de cancelación en dos fases. Establece reglas de seguridad strictly scoped como el control de `token_version` en middleware, secretos aislados y los criterios de aceptación para la fase F4.
