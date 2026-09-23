---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Suite de pruebas unitarias para el paquete legaltext. Comprueba el correcto parseo de los JSON embebidos, la constancia del checksum SHA256 calculado, la verificación de versión en VerifyVersion y la unicidad del payload retornado por SealPayload ante variaciones de cuenta o tiempo.

## Relaciones

- [[backend/legal/embed.go|embed.go]]: Archivo principal testeado
