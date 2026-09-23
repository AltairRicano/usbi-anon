---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Este archivo sirve como el documento único de planeación vigente del proyecto, sustituyendo y consolidando todos los planes maestros e históricos previos en 5 fases de maduración (M1 a M5). Modela la arquitectura de producción en Hostinger, el recálculo y veracidad de resultados de juegos en el backend, la integración de avisos de privacidad y el binario de control `usbictl`. Garantiza invariantes críticas como la seudonimización y minimización de datos sin PII directa, reglas estrictas de mitigación de trampas en puntajes y despliegues orquestados vía Compose con roles Postgres aislados (`scram-sha-256`).

Relacionado con:
- [[DEPLOY.md.md|DEPLOY.md]] — guía operativa de despliegue (M3)
- [[docker-compose.yml.md|docker-compose.yml]] — orquestación de servicios (M3.2/M3.8)
- [[plan/Convenciones_de_color_UV.md.md|Convenciones de color UV]] — identidad visual institucional
- [[estado_proyecto.md.md|estado_proyecto.md]] — bitácora histórica de decisiones (2026-08-24 → 2026-09-17)
