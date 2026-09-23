---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Escena de Phaser 3 que dibuja la cuadrícula interactiva del crucigrama y gestiona la entrada de datos. Implementa la sincronización con un elemento `<input>` HTML real y transparente fuera de pantalla para forzar la apertura del teclado virtual en navegadores móviles (iOS/Android) durante la interacción táctil.

## Funciones

### CrosswordScene.init
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Recupera las instancias de `CrosswordEngine` y la función de finalización desde los datos de escena o el registro de Phaser.

### CrosswordScene.create
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Calcula el escalado y centrado de la cuadrícula, crea los contenedores gráficos de celdas, inicializa el campo de texto DOM para el teclado virtual y se suscribe a los cambios del motor.

### CrosswordScene.drawState
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Actualiza visualmente el lienzo modificando el color de relleno de celdas según estén bloqueadas (correctas) o seleccionadas, y actualiza los caracteres ingresados.

### CrosswordScene.createDomInput
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Crea y conecta al DOM un elemento `<input>` invisible y fuera de pantalla para capturar la entrada de texto y eventos de teclado de manera síncrona en dispositivos móviles.

### CrosswordScene.handleKeyDown
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Procesa la navegación mediante las flechas del teclado, la tecla Backspace para borrar y la entrada de letras para enviarlas al motor de juego.

### CrosswordScene.shutdown
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Cancela las suscripciones al motor, elimina los oyentes de eventos de teclado y remueve el elemento de texto DOM para prevenir fugas de memoria.

## Relaciones

(Escena Phaser que usa @usbi/engine CrosswordEngine)
