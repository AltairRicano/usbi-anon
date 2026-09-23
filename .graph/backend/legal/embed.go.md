---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-22
---

Empotra los textos de aviso de privacidad en el binario usando `go:embed` para mantenerlos versionados en el repositorio. Parsea las secciones estructuradas, genera un hash SHA256 del contenido servido y expone funciones para validar la versión activa y construir el payload de sellado HMAC.

## Funciones

### mustBuildCurrent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Parsea los archivos JSON embebidos para el aviso simplificado e integral (lanzando panic si están malformados), calcula el hash SHA256 sobre ambos contenidos y retorna la estructura Notice inicializada.

### VerifyVersion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida que la cadena de versión proporcionada no esté vacía y coincida exactamente con CurrentVersion, asegurando que el cliente no intente registrar aceptaciones con versiones desactualizadas.

### SealPayload
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Construye la cadena delimitada por tuberías con el UUID de la cuenta, versión del aviso, checksum SHA256 y fecha de aceptación, lista para ser firmada con HMAC y sellar el registro legal.

## Relaciones

- [[backend/legal/embed_test.go|embed_test.go]]: Pruebas unitarias
- [[backend/legal/aviso_simplificado_v1.json|aviso_simplificado_v1.json]]: Texto embebido
- [[backend/legal/aviso_integral_v1.json|aviso_integral_v1.json]]: Texto embebido
- [[backend/internal/legal/service.go|legal.Service]]: Servicio que consume este paquete
