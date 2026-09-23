---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Suite de pruebas unitarias para la función [[backend/internal/httpjson/decode.go.md#DecodeStrict|DecodeStrict]]. Verifica que el decodificador rechace solicitudes con campos no definidos en la estructura de destino y detecte cuerpos de mensaje que contengan múltiples valores JSON concatenados.
