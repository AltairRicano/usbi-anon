---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Store Zustand persistido en localStorage para almacenar preferencias de accesibilidad y apariencia independientemente de la sesión del usuario. Aplica y remueve dinámicamente las clases CSS en document.documentElement al cambiar los estados o rehidratar el almacenamiento.

## Funciones

### applyDocumentClasses
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Función global que actualiza las clases CSS en document.documentElement según la configuración activa de tema, daltonismo, movimiento y texto.

### useSettingsStore
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Hook de estado persistido que gestiona las preferencias de usuario (tema, daltonismo, sonidos, movimiento, escala) e invoca applyDocumentClasses ante variaciones de estado.
