---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Suite de pruebas unitarias para el servicio legal. Valida que [[backend/internal/legal/service.go.md#Service.CurrentNotice|Service.CurrentNotice]] recupere y mapee adecuadamente los campos de versión, fecha efectiva, secciones simplificadas y completas, así como la concordancia del checksum contra [[backend/legal/embed.go.md|legaltext.Current()]], el aviso legal embebido.
