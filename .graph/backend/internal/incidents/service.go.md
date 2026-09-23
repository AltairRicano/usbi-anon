---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define el servicio central del módulo de incidentes, responsable de la creación de registros de seguridad y la inicialización de la estructura de servicio. Garantiza el principio de No-Repudio al sellar criptográficamente cada incidente con HMAC usando un secreto obligatorio, exigiendo permisos de administrador y validando rangos de tiempo y límites de tamaño en los campos narrativos.

## Funciones

### NewService
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Constructor del servicio de incidentes que valida la presencia obligatoria de la clave secreta HMAC, lanzando un panic si se encuentra vacía para evitar operar sin garantías criptográficas.

### Service.CreateIncident
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Registra un nuevo incidente de seguridad para administradores validando severidad, límites de texto y que la fecha de detección no sea futura, generando la firma HMAC de evidencia con [[backend/internal/crypto/hash.go.md#GenerateHMAC|GenerateHMAC]], persistiendo con [[backend/internal/repository/security_incident_queries.go.md#Queries.InsertSecurityIncident|Queries.InsertSecurityIncident]] y registrando el evento con [[backend/internal/audit/audit.go.md#Log|audit.Log]] dentro de la misma transacción.
