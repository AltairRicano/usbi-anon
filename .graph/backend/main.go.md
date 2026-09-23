---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Punto de entrada principal de la aplicación HTTP en Go. Inicializa variables de entorno y registros de auditoría estructurados `slog`, gestiona la apertura de tres pools de conexiones segregados por rol PostgreSQL (`usbi_app`, `usbi_moderador`, `usbi_dbmaint`), instancia servicios y controladores en el enrutador chi, activa tareas en segundo plano y maneja la cancelación limpia por señales `SIGINT`/`SIGTERM`.

## Funciones

### main
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicializa variables de entorno, el sistema de logs, establece las conexiones de tres pools de base de datos aislados por privilegios, inyecta dependencias en servicios y controladores HTTP, lanza programadores en segundo plano y coordina el apagado controlado (*graceful shutdown*) del servidor TLS.

### openPool
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Abre y valida mediante `Ping` una conexión a PostgreSQL configurando límites de conexiones abiertas, inactivas y tiempos máximos de vida/inactividad especificados por variables de entorno.

### readyCheck
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Crea una función de verificación de salud que comprueba la disponibilidad operativa ejecutando pings con contexto a los pools de base de datos de jugador y moderador.

### newLogger
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Configura e instancia un logger estructurado de `slog` con salida en formato JSON o texto plano y nivel de log ajustable según las variables `LOG_LEVEL` y `LOG_FORMAT`.

## Relaciones

- [[backend/internal/config/config.go|config]]: Carga variables de entorno
- [[backend/internal/transport/router.go|router]]: Enruta SetupRoutes
- [[backend/internal/auth|auth]]: Inyecta authSvc
- [[backend/internal/quiz|quiz]]: Inyecta quizPlayerSvc y quizAdminSvc
- [[backend/internal/levels|levels]]: Inyecta levelsPlayerSvc y levelsAdminSvc
- [[backend/internal/devices|devices]]: Inyecta devicesSvc
- [[backend/internal/incidents|incidents]]: Inyecta incidentsSvc
- [[backend/internal/badges|badges]]: Inyecta badgesAdminSvc
- [[backend/internal/auditlog|auditlog]]: Inyecta auditLogAdminSvc
- [[backend/internal/interestlinks|interestlinks]]: Inyecta interestLinksPlayerSvc y interestLinksAdminSvc
- [[backend/internal/legal|legal]]: Inyecta legalSvc
- [[backend/internal/dbmaint|dbmaint]]: Inicia scheduler de particiones
