---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Módulo para la identificación y registro automático de dispositivos cliente. Aplica una heurística basada en User-Agent y dimensiones de pantalla para clasificar el tipo de dispositivo y realiza un registro silencioso en segundo plano tras iniciar sesión sin bloquear la navegación.

## Funciones

### guessDeviceKind
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Infiere la clasificación del dispositivo ('movil', 'tablet', 'escritorio', 'laptop' u 'otro') examinando la cadena del User-Agent y el ancho de ventana.

### registerCurrentDevice
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Registra o actualiza el dispositivo cliente en `/devices` de manera asíncrona tras iniciar sesión, almacenando el ID resultante en `localStorage` e ignorando fallos para no interrumpir la experiencia de usuario.

## Relaciones

- Usa [[frontend/src/shared/apiClient.ts|apiClient]]
- Usa [[frontend/src/features/offline-processes/schemas.ts|schemas]]
