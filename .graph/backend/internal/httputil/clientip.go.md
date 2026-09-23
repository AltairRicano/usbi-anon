---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Utilidad HTTP compartida que extrae la dirección IP del socket peer a partir de la solicitud `r.RemoteAddr`. Su principal decisión de seguridad es ignorar deliberadamente los encabezados de reenvío enviados por el cliente para evitar la suplantación de identidad en funciones críticas como rate limiting y registro de auditoría, delegando la confianza en proxies al middleware aguas arriba.

## Funciones

### ClientIP
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la IP del cliente separando el host del puerto en `r.RemoteAddr` mediante `net.SplitHostPort` o realizando un fallback al valor sin espacios si la dirección carece de puerto. No consulta encabezados HTTP provistos por el cliente para impedir falsificaciones de IP, dejando la gestión de proxies confiables al middleware `RealIP` aguas arriba.

La usan los handlers que necesitan la IP como evidencia de auditoría: [[backend/internal/incidents/handler.go.md#Handler.CreateIncident|Handler.CreateIncident]] y [[backend/internal/incidents/handler.go.md#Handler.Update|Handler.Update]] la pasan a [[backend/internal/incidents/service.go.md#Service.CreateIncident|Service.CreateIncident]] y [[backend/internal/incidents/admin_read.go.md#Service.Update|Service.Update]] para sellar el registro de `audit_log`.
