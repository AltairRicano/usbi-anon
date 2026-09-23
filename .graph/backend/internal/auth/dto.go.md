---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define las estructuras de transferencia de datos (DTOs) para los flujos de registro en 3 pasos, inicio de sesión, gestión de sesiones y administración de cuentas. Establece el contrato formal de entrada y salida del módulo de autenticación, incluyendo la entrega en claro de contraseñas generadas únicamente tras confirmar el registro. `LoginResponse` envuelve el `domain.User` público (definido en [[backend/internal/domain/models.go.md|models.go]]); `RegisterConfirmResponse` es la respuesta que arma [[backend/internal/auth/service.go.md#Service.RegisterConfirm|Service.RegisterConfirm]] y `MeResponse` la que arma [[backend/internal/auth/service.go.md#Service.Me|Service.Me]], comparando `PrivacyNoticeVersion` contra `legaltext.CurrentVersion` de [[backend/legal/embed.go.md|embed.go]].
