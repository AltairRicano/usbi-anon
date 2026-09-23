---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Componente de React para el juego de memoria (Memorama). Renderiza una cuadrícula de cartas animadas con giros 3D en CSS, integrando lógica de accesibilidad por teclado y cálculos de contraste dinamicos proporcionados por el `MemoryEngine`.

## Funciones

### MemoryGame
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Muestra la retícula de cartas del juego de memoria, gestiona sus estados de volteo 3D y habilita la interacción mediante clics o teclas Enter/Espacio.

### MemoryGame.handleCardClick
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Voltea la carta seleccionada en el `MemoryEngine` y, al haber dos cartas descubiertas, programa una pausa de 1 segundo para validar si forman pareja.

## Relaciones

(Componente de juego que usa engine de @usbi/engine)
