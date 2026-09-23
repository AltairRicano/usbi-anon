---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Función utilitaria centralizada para la extracción y formato de mensajes de error de red. Prioriza la propiedad `detail` proveniente del estándar RFC 7807 Problem Details retornado por el servidor en Go sobre los mensajes por defecto de Axios.

## Funciones

### errorMessage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Extrae y devuelve el detalle descriptivo de una falla recibida desde la API, el mensaje de un objeto `Error` o un texto de respaldo especificado.

## Relaciones

- Usa axios para manejo de errores HTTP
