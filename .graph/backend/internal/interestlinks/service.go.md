---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Declara las constantes de límites de texto, errores de dominio del paquete y funciones auxiliares de validación para categorías y enlaces. Espeja en la capa de aplicación las restricciones CHECK de la base de datos (longitudes máximas, formato hex de color y esquemas de URL HTTP/HTTPS) para devolver errores 422 legibles antes de interaccionar con la base de datos.

`validateCategoryInput` y `validateLinkInput` son invocadas por [[backend/internal/interestlinks/admin_service.go.md#AdminService.CreateCategory|AdminService.CreateCategory]]/`UpdateCategory` y `CreateLink`/`UpdateLink` en [[backend/internal/interestlinks/admin_service.go.md|admin_service.go]] antes de tocar la base de datos.

## Funciones

### validateCategoryInput
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Sanitiza espacios en blanco y verifica que el nombre de la categoría no esté vacío ni supere los 200 caracteres permitidos.

### validateLinkInput
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Sanitiza y valida las restricciones de longitud de título (máx 50) y descripción (máx 100), la sintaxis del color hexadecimal (`#RRGGBB`) y la validez del esquema URL.

### isHTTPURL
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Parsea la dirección URI para asegurar que utilice un esquema `http` o `https` y posea un host válido, rechazando esquemas potencialmente inseguros como `javascript:` o `data:`.
