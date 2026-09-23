---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Elimina las funciones almacenadas purge_account_quiz_answers y null_user_in_pseudonymizable_ledgers utilizadas para el proceso de cancelación de cuentas. Revierte los permisos especiales otorgados para el aislamiento de tareas de depuración en la base de datos.

## Relaciones

- [[backend/migrations/0004_procedimientos_purga_moderador.up.sql|0004 up]]: Migración forward de este cambio
