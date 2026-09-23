---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Empotra todos los archivos de migración SQL (*.sql) dentro del binario compilado de Go utilizando embed.FS. Garantiza que el proceso de despliegue transporte exactamente las versiones de migraciones correspondientes a la versión del código sin depender del sistema de archivos local ni de recursos externos.

## Relaciones

- [[backend/cmd/usbictl/migrate.go|migrate.go]]: Carga FS para aplicar migraciones
- [[backend/migrations/0001_esquema_unificado.up.sql|0001 up]]: Primera migración baseline
- [[backend/migrations/0001_esquema_unificado.down.sql|0001 down]]: Reversión baseline
- [[backend/migrations/0002_roles_player_admin.up.sql|0002 up]]: Migraciones posteriores
- [[backend/migrations/0002_roles_player_admin.down.sql|0002 down]]
- [[backend/migrations/0003_enlaces_interes_y_sugerencias.up.sql|0003 up]]
- [[backend/migrations/0003_enlaces_interes_y_sugerencias.down.sql|0003 down]]
- [[backend/migrations/0004_procedimientos_purga_moderador.up.sql|0004 up]]
- [[backend/migrations/0004_procedimientos_purga_moderador.down.sql|0004 down]]
- [[backend/migrations/0005_particion_dbmaint.up.sql|0005 up]]
- [[backend/migrations/0005_particion_dbmaint.down.sql|0005 down]]
- [[backend/migrations/0006_b1_b4_lectura_admin.up.sql|0006 up]]
- [[backend/migrations/0006_b1_b4_lectura_admin.down.sql|0006 down]]
- [[backend/migrations/0007_verificacion_resultado_nivel.up.sql|0007 up]]
- [[backend/migrations/0007_verificacion_resultado_nivel.down.sql|0007 down]]
- [[backend/migrations/0008_matriz_permisos.up.sql|0008 up]]
- [[backend/migrations/0008_matriz_permisos.down.sql|0008 down]]
