---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define la configuración central de rutas HTTP y middlewares de seguridad usando Go Chi para la API backend. Es responsable del enrutamiento de endpoints públicos, autenticados y de administración, aplicando validación de tokens JWT con revocación en base de datos, políticas CORS por origen, cabeceras de seguridad y control de tasa (rate limiting) diferencial por IP para prevenir ataques de fuerza bruta y denegación de servicio.

`RouterDependencies` cablea los handlers de todos los paquetes de negocio: [[backend/internal/auth/handler.go.md|auth/handler.go]], [[backend/internal/quiz/admin_service.go.md|quiz/admin_service.go]] (banco de preguntas), [[backend/internal/sync/handler.go.md|sync/handler.go]], [[backend/internal/levels/handler.go.md|levels/handler.go]], [[backend/internal/incidents/handler.go.md|incidents/handler.go]], [[backend/internal/devices/handler.go.md|devices/handler.go]], [[backend/internal/interestlinks/handler.go.md|interestlinks/handler.go]], [[backend/internal/suggestions/handler.go.md|suggestions/handler.go]], [[backend/internal/legal/handler.go.md|legal/handler.go]], [[backend/internal/badges/handler.go.md|badges/handler.go]] y [[backend/internal/auditlog/handler.go.md|auditlog/handler.go]]. `jwtAuthMiddleware` revalida cada petición autenticada contra [[backend/internal/repository/account_queries.go.md#Queries.GetAccountByID|repository/account_queries.go#Queries.GetAccountByID]].

## Funciones

### ClaimsFromContext
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Extrae los claims JWT inyectados en el contexto de la petición HTTP; devuelve nil si no están presentes.

### SetupRoutes
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Configura el árbol de rutas HTTP y la cadena global de middlewares (RequestID, logger, recoverer, timeout, security headers, CORS y límite de cuerpo). Registra endpoints públicos, autenticados y de administración, instala stubs HTTP 501 para dependencias no configuradas y retorna la función de limpieza para las goroutines de rate limiting.

### maxBodyBytesMiddleware
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Restringe el tamaño máximo del cuerpo de las peticiones HTTP (6 MB por defecto mediante MaxBytesReader) para prevenir ataques de agotamiento de memoria.

### requestLogger
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Middleware que registra información estructurada (slog) por cada solicitud HTTP, incluyendo método, ruta, estado HTTP, bytes escritos, latencia en milisegundos, ID de petición e IP del cliente.

### readyHandler
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Handler de verificación de disponibilidad que ejecuta la función de comprobación de dependencias con un tiempo límite de 2 segundos, respondiendo con HTTP 503 en caso de fallo sin exponer detalles internos del driver.

### notImplementedHandler
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Devuelve una respuesta de error HTTP 501 Not Implemented estructurada bajo RFC 7807 (problem details) para operaciones pendientes de implementación.

### securityHeaders
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Establece cabeceras HTTP de seguridad conservadoras (X-Content-Type-Options, X-Frame-Options DENY, Referrer-Policy y Content-Security-Policy), aplicando HSTS únicamente cuando la conexión TLS finaliza en el proceso.

### corsMiddleware
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Aplica políticas CORS verificando el encabezado Origin contra una lista explícita o comodín, absteniéndose de inyectar cabeceras si el origen no coincide y respondiendo a solicitudes preflight OPTIONS.

### jwtAuthMiddleware
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Middleware que valida la cabecera Authorization Bearer y la firma del JWT mediante [[backend/internal/crypto/jwt.go.md#ValidateToken|crypto/jwt.go#ValidateToken]]. Revalida en base de datos la versión del token (token_version) y el estado activo de la cuenta mediante [[backend/internal/repository/account_queries.go.md#Queries.GetAccountByID|repository/account_queries.go#Queries.GetAccountByID]] para permitir revocación inmediata en cierres de sesión o suspensiones.

### visitorTracker.getLimiter
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene o inicializa de forma concurrente el limitador de tasa (token bucket) asociado a la IP de un cliente, actualizando su estampa de última actividad.

### visitorTracker.expireIdle
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina del registro de visitantes aquellos limitadores por IP que hayan superado el tiempo máximo de inactividad especificado.

### rateLimiters.warnIfLikelyProxyMisconfigured
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Emite una advertencia única en logs si tras 50 peticiones se detecta una sola IP de origen, indicando una posible omisión de la bandera TrustProxyHeaders tras un proxy inverso.

### rateLimiters.generalMiddleware
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Aplica el control de tasa por IP para rutas autenticadas generales (10 RPS, ráfaga 20), devolviendo HTTP 429 Too Many Requests si se sobrepasa el límite.

### rateLimiters.authMiddleware
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Aplica control de tasa estricto (0.5 RPS, ráfaga 5) por IP sobre las rutas públicas de autenticación para mitigar ataques de fuerza bruta.
