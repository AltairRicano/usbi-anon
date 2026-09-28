---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-28
---

Componente en React para el juego de crucigrama. Se encarga de instanciar y gestionar la suscripción al `CrosswordEngine`, asignar numeración a las palabras compartiendo identificador entre aquellas que inicien en la misma celda, y presentar la interfaz que combina el tablero interactivo Phaser con la lista de pistas horizontales y verticales.

## Funciones

### CrosswordGame
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Renderiza la interfaz del crucigrama, pasa el motor al registro de Phaser, mantiene actualizado el estado con el `CrosswordEngine` y lista las pistas ordenadas identificando cuáles palabras han sido completadas.

## Relaciones

- Usa [[frontend/src/features/games/CrosswordScene.ts|CrosswordScene]]
- Usa [[frontend/src/shared/components/ui/Card.tsx|Card]]
- Usa [[frontend/src/shared/PhaserGame.tsx|PhaserGame]]
