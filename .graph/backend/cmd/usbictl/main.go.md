---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Punto de entrada principal del binario unificado de operaciones `usbictl`. Carga las variables de entorno de la aplicación y despacha la ejecución al subcomando correspondiente (`migrate`, `secrets`, `admin`, `doctor`, `pgconf`), ofreciendo una interfaz centralizada de administración.

## Funciones

### main
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Inicializa el entorno, interpreta los argumentos de la línea de comandos y enruta la invocación hacia la función encargada de ejecutar el subcomando solicitado.

## Relaciones

- [[backend/cmd/usbictl/migrate.go|migrate]]: Subcomando para aplicar/revertir migraciones
- [[backend/cmd/usbictl/admin.go|admin]]: Subcomando para crear el administrador inicial
- [[backend/cmd/usbictl/doctor.go|doctor]]: Subcomando para verificar permisos y salud de pools
- [[backend/cmd/usbictl/pgconf.go|pgconf]]: Subcomando para renderizar configuración de Postgres
