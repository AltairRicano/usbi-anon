---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Script SQL de inserción manual para registrar al primer usuario administrador en la base de datos tras ejecutar las migraciones. Requiere reemplazar previamente los marcadores de posición con los hashes y sellos generados por el binario `cmd/hash_password`.

## Relaciones

- [[backend/cmd/hash_password/main.go|hash_password]]: Genera valores insertados aquí
- [[backend/cmd/usbictl/admin.go|admin]]: Subcomando que reemplaza esta inserción manual
