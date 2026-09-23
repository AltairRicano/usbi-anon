---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Registro central de plantillas para el editor de contenidos. Asocia cada tipo de minijuego con su esquema Zod, componentes de formulario y previsualización, y generadores de valores iniciales, incluyendo validaciones avanzadas de conectividad e inmutabilidad de respuestas en crucigramas.

## Funciones

### MakerCrosswordSchema
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Esquema Zod con refinamiento personalizado que valida que las respuestas del crucigrama no contengan duplicados y puedan formar una cuadrícula interconectada.

### triviaDefaults
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera la estructura de datos por defecto para iniciar la creación de una trivia con preguntas en blanco.

### snakeLadderDefaults
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Genera la configuración inicial por defecto para un tablero de serpientes y escaleras de 6x6 con enlaces calculados por semilla.

## Relaciones

- [[frontend/src/features/content/maker/LevelMakerForm.tsx.md|LevelMakerForm]] — consumidor principal del registro
- [[frontend/src/features/content/maker/forms/TriviaForm.tsx.md|TriviaForm]] — formulario para trivia
- [[frontend/src/features/content/maker/forms/CrosswordForm.tsx.md|CrosswordForm]] — formulario para crucigrama
- [[frontend/src/features/content/maker/forms/WordSearchForm.tsx.md|WordSearchForm]] — formulario para sopa de letras
- [[frontend/src/features/content/maker/forms/PuzzleForm.tsx.md|PuzzleForm]] — formulario para rompecabezas
- [[frontend/src/features/content/maker/forms/FakeNewsForm.tsx.md|FakeNewsForm]] — formulario para noticias falsas
- [[frontend/src/features/content/maker/forms/MemoryForm.tsx.md|MemoryForm]] — formulario para memorama
- [[frontend/src/features/content/maker/forms/SnakeLadderForm.tsx.md|SnakeLadderForm]] — formulario para serpientes y escaleras
- [[frontend/src/features/content/maker/previews/TriviaPreview.tsx.md|TriviaPreview]] — previsualizador para trivia
- [[frontend/src/features/content/maker/previews/CrosswordPreview.tsx.md|CrosswordPreview]] — previsualizador para crucigrama
- [[frontend/src/features/content/maker/previews/WordSearchPreview.tsx.md|WordSearchPreview]] — previsualizador para sopa de letras
- [[frontend/src/features/content/maker/previews/PuzzlePreview.tsx.md|PuzzlePreview]] — previsualizador para rompecabezas
- [[frontend/src/features/content/maker/previews/FakeNewsPreview.tsx.md|FakeNewsPreview]] — previsualizador para noticias falsas
- [[frontend/src/features/content/maker/previews/MemoryPreview.tsx.md|MemoryPreview]] — previsualizador para memorama
- [[frontend/src/features/content/maker/previews/SnakeLadderPreview.tsx.md|SnakeLadderPreview]] — previsualizador para serpientes y escaleras
- [[frontend/src/features/content/maker/snakesLayout.ts.md|snakesLayout.ts]] — genera configuración de enlace para serpientes y escaleras
