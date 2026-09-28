---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-28
---

Página de prueba y ejecución de niveles locales guardados en `localStorage` desde el maker. Se encarga de normalizar los contenidos de cada tipo de plantilla y garantiza como decisión de negocio que las partidas locales no registren progreso, puntos de experiencia (XP) ni medallas en la cuenta del usuario.

## Funciones

### LocalLevelPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente que recupera la configuración del nivel local desde `localStorage`, inicializa la plantilla interactiva correspondiente en modo sandbox y administra la interfaz de resultados sin acumular XP.

### hasPlayableContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Verifica que el objeto de datos del nivel cumpla con los requerimientos mínimos de contenido establecidos para su tipo de plantilla antes de permitir su renderizado.

## Relaciones

- [[frontend/src/features/content/types.ts.md|types.ts]] — define tipo `LevelDTO` y funciones de normalización por plantilla

- [[frontend/src/features/games/components/FakeNewsGame.tsx.md|FakeNewsGame]] — importado vía lazy
- [[frontend/src/features/games/components/MemoryGame.tsx.md|MemoryGame]] — importado vía lazy
