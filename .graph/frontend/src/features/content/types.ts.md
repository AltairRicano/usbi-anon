---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Módulo que define los tipos TypeScript del dominio de contenido y proporciona funciones de normalización. Garantiza que las estructuras de datos recibidas para cada minijuego se conviertan a formatos seguros y válidos para la lógica de juego.

## Funciones

### templateTypeLabel
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Retorna el nombre legible en español correspondiente al identificador de tipo de plantilla.

### normalizeTriviaContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Filtra y sanitiza la lista de preguntas de trivia verificando que posean texto, al menos dos opciones y un índice correcto dentro del rango.

### normalizeMemoryContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Normaliza los pares de cartas configurados para el juego de memorama, retornando solo aquellos que sean jugables.

### normalizeMemoryBackColorContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene y valida la cadena de color hexadecimal para el reverso del memorama, aplicando un valor por defecto si es inválido.

### normalizeFakeNewsContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Sanitiza los elementos de noticias falsas asegurando la presencia de título y contenido válido.

### normalizeWordSearchContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Filtra las palabras de la sopa de letras manteniendo solo aquellas con 2 o más caracteres válidos.

### normalizePuzzleContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Normaliza los datos del rompecabezas devolviendo `null` si la frase no está definida.

### normalizeCrosswordContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Sanitiza la lista de palabras de crucigrama filtrando entradas con longitud menor a 2 caracteres o sin pista.

### normalizeSnakesContent
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida y normaliza la configuración del juego de serpientes y escaleras, retornando `null` si faltan las dimensiones o posiciones esenciales del tablero.

## Relaciones

- OfficialLevelPage — usa funciones de normalización para procesar contenido de nivel
- LocalLevelPage — usa funciones de normalización para procesar niveles locales
- SectionLevelsPage — usa `templateTypeLabel` para etiquetado
- AdminContentPage — usa tipos DTO y `templateTypeLabel`
- DashboardPage — usa tipos y `templateTypeLabel`
