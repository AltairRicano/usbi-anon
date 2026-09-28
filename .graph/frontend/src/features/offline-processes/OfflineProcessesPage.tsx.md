---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-28
---

Vista de administración de procesos fuera de línea que permite visualizar los dispositivos cliente registrados por el usuario, consultar el historial paginado de eventos de sincronización mediante cursores y revocar registros de dispositivos advirtiendo sobre el impacto en la sincronización local.

## Funciones

### OfflineProcessesPage
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Componente principal que gestiona el estado de los dispositivos vinculados, filtros seleccionados y el historial de sincronización.

### OfflineProcessesPage.loadDevices
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la lista de dispositivos registrados por la cuenta autenticada realizando una petición GET a `/devices`.

### OfflineProcessesPage.loadEvents
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Consulta el historial de eventos en `/sync/events` filtrando opcionalmente por dispositivo y gestionando la paginación basada en cursor.

### OfflineProcessesPage.confirmRevoke
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina el dispositivo mediante una llamada DELETE a `/devices/:id` y limpia la selección si el dispositivo revocado estaba siendo filtrado.

## Relaciones

- Usa [[frontend/src/shared/components/ui/Button.tsx|Button]]
- Usa [[frontend/src/shared/components/ui/HomeButton.tsx|HomeButton]]
- Usa [[frontend/src/shared/apiClient.ts|apiClient]]
- Usa [[frontend/src/shared/errorMessage.ts|errorMessage]]
- Usa [[frontend/src/features/offline-processes/schemas.ts|schemas]]
