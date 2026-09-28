---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-28
---

Vista correspondiente a la pestaña "Más" del panel de control. Presenta los carruseles de enlaces de interés agrupados por categoría y un buzón para el envío de sugerencias anónimas que no quedan vinculadas a la cuenta del usuario.

## Funciones

### MoreTab
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente contenedor que obtiene los grupos de enlaces desde `/interest-links` y renderiza la sección de sugerencias.

### MoreTab.submitSuggestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Envía la sugerencia redactada al endpoint anónimo `/suggestions` y limpia el campo de texto tras recibir la respuesta de confirmación.

## Relaciones

- Usa [[frontend/src/features/dashboard/InterestLinkCarousel.tsx|InterestLinkCarousel]]
- Usa [[frontend/src/features/dashboard/interestLinksSchemas.ts|interestLinksSchemas]]
- Usa [[frontend/src/shared/components/ui/Button.tsx|Button]]
- Usa [[frontend/src/shared/apiClient.ts|apiClient]]
- Usa [[frontend/src/shared/errorMessage.ts|errorMessage]]
