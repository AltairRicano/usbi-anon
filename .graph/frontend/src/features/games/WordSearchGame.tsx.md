---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Componente contenedor React para la sopa de letras. Se encarga de instanciar el `WordSearchEngine`, integrar el componente `PhaserGame` asignando la escena `WordSearchScene`, y desplegar un panel accesible que refleja en tiempo real las palabras encontradas y la puntuación.

## Funciones

### WordSearchGame
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicializa el motor de la sopa de letras, gestiona su suscripción de estado y renderiza la estructura principal con la barra lateral de palabras y el lienzo del juego.

### WordSearchGame.handleGameReady
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Callback invocado cuando el canvas Phaser está listo, iniciando la escena `WordSearchScene` con la instancia del motor y el manejador de finalización.

## Relaciones

- Usa [[frontend/src/features/games/WordSearchScene.ts|WordSearchScene]]
- Usa [[frontend/src/shared/components/ui/Card.tsx|Card]]
- Usa [[frontend/src/shared/PhaserGame.tsx|PhaserGame]]
