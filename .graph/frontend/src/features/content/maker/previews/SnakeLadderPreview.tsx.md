---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Componente de previsualización para el tablero de Serpientes y Escaleras. Dibuja la cuadrícula de celdas marcando inicio, meta y conexiones SVG (líneas rectas para escaleras verdes y curvas de Bézier para serpientes rojas) calculadas en porcentaje.

## Funciones

### SnakeLadderPreview
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente contenedor de la previsualización del juego de serpientes y escaleras.

### SnakeLadderBoardPreview
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Construye la cuadrícula visual del tablero calculando celdas e integrando los trazos SVG de serpientes y escaleras.

### BoardLink
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula y renderiza las trayectorias SVG rectas o curvas que unen las posiciones de inicio y fin de una serpiente o escalera.

## Relaciones

- [[frontend/src/features/content/maker/registry.ts.md|registry.ts]] — previsualizador registrado en el registro de plantillas
- [[frontend/src/features/content/maker/snakesLayout.ts.md|snakesLayout.ts]] — utiliza funciones de cálculo de posiciones de celdas
