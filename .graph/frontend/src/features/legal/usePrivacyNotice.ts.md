---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Hook personalizado para la obtención pública del aviso de privacidad vigente desde `/legal/privacy-notice`. Mantiene el resultado en una variable de módulo como memoria caché para evitar peticiones repetidas al navegar entre páginas.

## Funciones

### usePrivacyNotice
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Realiza la petición HTTP a la API legal, valida la estructura del aviso devuelto mediante Zod y almacena el resultado en caché.

## Relaciones

- Usa [[frontend/src/shared/apiClient.ts|apiClient]]
- Usa [[frontend/src/shared/errorMessage.ts|errorMessage]]
- Usa [[frontend/src/features/legal/schemas.ts|schemas]]
