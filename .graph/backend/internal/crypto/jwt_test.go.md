---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Suite de pruebas de integración criptográfica de extremo a extremo para [[backend/internal/crypto/jwt.go.md#GenerateToken|GenerateToken]]/[[backend/internal/crypto/jwt.go.md#ValidateToken|ValidateToken]] que valida el ciclo completo de emisión y parseo, el rechazo de tokens expirados o con firma/secreto incorrecto, la detección de alteración de payload y la protección contra ataques de confusión de algoritmos (como firmas alg=none).
