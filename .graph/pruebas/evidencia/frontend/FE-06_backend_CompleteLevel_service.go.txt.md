---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Extracto del servicio backend en Go que procesa la finalización de niveles por parte de un usuario. Ejecuta la lógica transaccional con aislamiento serializable para calcular experiencia (XP), registrar el intento, actualizar el progreso acumulado, mantener la racha diaria de actividad y otorgar insignias elegibles.

[[pruebas/03_frontend/FE-06_superficie_de_ataque_cliente.md|Documento de análisis FE-06]]
[[backend/internal/levels/player_service.go.md|Código backend: Service.CompleteLevel]]

## Funciones

### Service.CompleteLevel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida identificadores y puntaje no negativo, inicia una transacción serializable para evitar condiciones de carrera en intentos simultáneos, calcula los XP según dificultad e intento, registra el intento y actualiza el progreso, la racha diaria de actividad, el historial de experiencia y otorga insignias aplicables.
