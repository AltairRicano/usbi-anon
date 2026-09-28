---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-28
---

Componente React estilo tarjetas deslizables para evaluar noticias falsas o verdaderas mediante animaciones de `framer-motion`. Permite al usuario calificar artículos arrastrando las tarjetas horizontalmente (izquierda para Verdadero, derecha para Falso) o mediante botones dedicados.

## Funciones

### FakeNewsGame
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Renderiza las tarjetas interactiva con soporte para gestos táctiles de arrastre y botones de clasificación, comunicando las respuestas al `FakeNewsEngine`.

### FakeNewsGame.handleSwipe
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Procesa la clasificación enviada (Verdadero/Falso), consulta el resultado al motor e intercambia la tarjeta actual por la siguiente noticia o concluye el juego.

## Relaciones

(Componente de juego que usa engine de @usbi/engine)
