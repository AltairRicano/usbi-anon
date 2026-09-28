---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-28
---

Página para jugar niveles oficiales cargados desde la API remota. Al finalizar la partida envía la puntuación, tiempo y respuestas al servidor para calcular y registrar la experiencia (XP) obtenida, rachas y medallas desbloqueadas; en niveles no publicados opera únicamente como vista previa sin alterar el progreso oficial.

## Funciones

### OfficialLevelPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente principal que consulta el nivel oficial al servidor, renderiza dinámicamente el juego asociado y registra los resultados de la partida consumiendo el endpoint de finalización.

### hasPlayableContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Evalúa la validez del contenido del nivel oficial comprobando que satisfaga la estructura mínima requerida según la plantilla del juego.

## Relaciones

- [[frontend/src/features/content/schemas.ts.md|schemas.ts]] — valida la estructura del nivel con `LevelDTOSchema`
- [[frontend/src/features/content/types.ts.md|types.ts]] — define tipos como `LevelDTO`, `CompleteLevelResponse` y funciones de normalización por plantilla

- [[frontend/src/features/games/components/FakeNewsGame.tsx.md|FakeNewsGame]] — importado vía lazy
- [[frontend/src/features/games/components/MemoryGame.tsx.md|MemoryGame]] — importado vía lazy
