---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Documento de evidencia de auditoría de seguridad que analiza el nivel de confianza del servidor respecto a la puntuación enviada por el cliente. Destaca que la única validación efectuada en el backend es verificar que el puntaje no sea negativo (req.Score < 0), sin validar límites superiores contextuales según el nivel.

[[pruebas/03_frontend/FE-06_superficie_de_ataque_cliente.md|Documento de análisis FE-06]]
