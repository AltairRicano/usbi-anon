---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Página pública de accesibilidad y configuración de apariencia. Permite modificar temas (claro/oscuro), escala de texto, filtros de daltonización (deuteranopía, protanopía, tritanopía), reducción de animaciones y sonido de juego sin requerir un inicio de sesión previo.

## Funciones

### SettingsPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente principal de la vista de configuración de apariencia y accesibilidad.

## Relaciones

- Usa [[frontend/src/shared/components/ui/HomeButton.tsx|HomeButton]]
- Usa [[frontend/src/features/settings/useSettingsStore.ts|useSettingsStore]]

### SegmentButton
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente de interfaz para grupos de botones con estilo de radio seleccionable (tema, escala de texto).

### ToggleRow
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente de interfaz con soporte táctil para casillas de verificación de opciones binarias (animaciones y sonido).
