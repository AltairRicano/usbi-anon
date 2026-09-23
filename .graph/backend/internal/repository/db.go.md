---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Define la interfaz `DBTX` y la estructura genérica `Queries` generada por `sqlc` para abstraer las ejecuciones SQL sobre conexiones de base de datos o transacciones. Proporciona el soporte base para instanciar repositorios y desacoplar la capa de persistencia. `repository.New(db)` es invocado por [[backend/internal/testdb/testdb.go.md#Setup|testdb/testdb.go#Setup]] para construir `Queries` sobre el esquema aislado de cada prueba de integración, y `WithTx` es usado por prácticamente todo servicio que abre una transacción (p. ej. [[backend/internal/sync/service.go.md#Service.ProcessSync|sync/service.go#Service.ProcessSync]]).
