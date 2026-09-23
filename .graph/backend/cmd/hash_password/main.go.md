---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Herramienta de línea de comandos independiente destinada a generar el hash Argon2id de la contraseña del administrador inicial sin depender de red ni base de datos. Mediante la opción `--seal`, permite además generar un UUID V7 y calcular el sello HMAC correspondiente a la aceptación de los avisos de privacidad.

## Funciones

### main
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Procesa los parámetros de entrada, valida los campos requeridos para la generación del sello de privacidad si la bandera `--seal` está activa, y emite por consola el hash Argon2id y la firma HMAC calculada.

## Relaciones

- [[backend/cmd/usbictl/admin.go|admin]]: Subcomando que integra esta funcionalidad directamente
