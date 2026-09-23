---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Suite de pruebas unitarias que verifica el comportamiento de [[backend/internal/crypto/hash.go.md#HashPassword|HashPassword]] y [[backend/internal/crypto/hash.go.md#VerifyPassword|VerifyPassword]] con Argon2id, asegurando que se utilice un único hilo de ejecución (p=1) para la conservación de recursos y validando el flujo completo de generación y verificación correcta de contraseñas.
