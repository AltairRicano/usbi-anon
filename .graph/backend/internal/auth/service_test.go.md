---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Suite de pruebas unitarias para el servicio de autenticación. Verifica el comportamiento de [[backend/internal/auth/service.go.md#Service.RegisterAnswers|Service.RegisterAnswers]] ante distintas versiones del aviso de privacidad, comparando contra `legaltext.CurrentVersion` de [[backend/legal/embed.go.md|embed.go]] — asegurando que se rechacen peticiones con versiones obsoletas o vacías y se acepten únicamente las que coincidan con la versión vigente.
