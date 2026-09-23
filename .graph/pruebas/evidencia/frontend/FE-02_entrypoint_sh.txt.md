---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Contiene el script de arranque para la instancia única de stress-test que ejecuta el frontend, backend Go y PostgreSQL 15 en un mismo contenedor de 500 MB RAM. Define la inicialización de la base de datos, migraciones, privilegios del usuario usbi, compilación e inicio del backend y servido del frontend reutilizando dist/ previo para no agotar la memoria.

[[pruebas/03_frontend/FE-02_dependencias_y_build.md|Documento de análisis FE-02]]
