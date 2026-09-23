---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Implementa la capa HTTP para la administración de incidentes de seguridad expuestos mediante la API v1. Se encarga de validar los claims del token JWT de la solicitud, decodificar el cuerpo y parámetros de consulta, y mapear los errores del servicio de negocio a respuestas estandarizadas RFC 7807 (Problem Details). Explicitamente no expone ningún endpoint de eliminación.

Obtiene la IP del cliente con [[backend/internal/httputil/clientip.go.md#ClientIP|ClientIP]] para pasarla como evidencia de auditoría a [[backend/internal/incidents/service.go.md#Service.CreateIncident|Service.CreateIncident]] y a [[backend/internal/incidents/admin_read.go.md#Service.Update|Service.Update]] (definido en `admin_read.go`, que también resuelve `List` y `Get`).

## Funciones

### Handler.CreateIncident
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Maneja peticiones HTTP POST para crear incidentes, validando los claims JWT, decodificando el cuerpo JSON de forma estricta y delegando la creación al servicio junto con metadatos de cliente.

### Handler.List
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Maneja peticiones HTTP GET para listar incidentes paginados, extrae y valida el tamaño de página (entre 1 y 50) y el cursor de la URL antes de invocar el servicio.

### Handler.Get
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Maneja peticiones HTTP GET para obtener un incidente específico, parseando y validando que el parámetro de ruta sea un UUID válido.

### Handler.Update
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Maneja peticiones HTTP PATCH para modificar un incidente existente, validando el UUID de la ruta y decodificando la solicitud antes de invocar la actualización en el servicio.
