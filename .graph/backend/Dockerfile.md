---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define la compilación multi-etapa para construir las imágenes de producción del backend y la herramienta CLI (`usbictl`). Utiliza Go 1.22 sobre Debian para compilar binarios estáticos sin CGO (`CGO_ENABLED=0`) y los traslada a un contenedor minimalista `distroless` sin shell ni gestor de paquetes para reducir drásticamente la superficie de ataque.

Usado en: [[docker-compose.yml.md|docker-compose.yml]] (servicio `api`).
