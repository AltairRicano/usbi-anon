---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Encapsula el ciclo de vida del motor de juegos Phaser 3 dentro de React mediante `forwardRef` y `useLayoutEffect`. Previene instancias duplicadas durante el montaje doble de React 18 StrictMode en desarrollo, sincroniza la escena activa en cada paso del loop y provee accesibilidad ARIA para el lienzo.

## Funciones

### PhaserGame
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicializa la instancia de `Phaser.Game` asociada al contenedor DOM, notifica cuando el juego está listo, actualiza la referencia a la escena activa en el evento `step` y realiza la destrucción limpia del lienzo al desmontar.

## Relaciones

- Usa Phaser 3 como motor de juegos
