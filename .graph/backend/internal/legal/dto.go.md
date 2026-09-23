---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define las estructuras de transferencia de datos (DTO) utilizadas para exponer las respuestas del aviso de privacidad y el registro de su aceptación. Previene la exposición directa de `legaltext.Section` (ver [[backend/legal/embed.go.md|embed.go]]) hacia la API pública — [[backend/internal/legal/service.go.md#Service.CurrentNotice|Service.CurrentNotice]] es quien mapea de uno a otro.
