// Package dbmaint mantiene las tablas particionadas por RANGO (level_attempts,
// daily_streak — ver migraciones 0007 y 0008) provistas con particiones explícitas
// por año con mucha antelación, para que la partición DEFAULT añadida en la
// migración 0008 se mantenga vacía en la práctica y nunca necesite absorber una ráfaga
// de tráfico ordinario. Esto es un asunto estructural de la base de datos, independiente de
// los trabajos de retención legal/privacidad en internal/maintenance. (Útil)
package dbmaint

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// yearsAhead controla cuántos años más allá del actual obtienen una partición
// explícita en cada ejecución. 2 años de antelación sobrevive cómodamente a cualquier
// brecha razonable entre ejecuciones del programador o despliegues. (Útil)
const yearsAhead = 2

var partitionedTables = []string{"level_attempts", "daily_streak"}

// Service asegura que existan particiones de rango anuales para las tablas particionadas.
// Habla con el *sql.DB crudo (no las Queries generadas por sqlc) porque
// CREATE TABLE ... PARTITION OF es un DDL de esquema, no una consulta que modele sqlc. (Útil)
type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// EnsurePartitions crea (de forma idempotente) una partición para el año de `now` y
// cada uno de los años en yearsAhead, para cada tabla particionada. (Relleno)
func (s *Service) EnsurePartitions(ctx context.Context, now time.Time) error {
	for _, stmt := range partitionStatements(now) {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("executing %q: %w", stmt, err)
		}
	}
	return nil
}

// partitionStatements construye el DDL idempotente para cada tabla particionada
// y cada año desde el año actual hasta el año actual + yearsAhead. Los nombres de
// tablas y años se extraen de una lista/rango interno fijo, nunca de la
// entrada del usuario, así que construir SQL con fmt.Sprintf es seguro aquí — de todas formas
// no hay forma de vinculación de parámetros para identificadores/rangos DDL en database/sql. (Útil)
func partitionStatements(now time.Time) []string {
	startYear := now.UTC().Year()
	stmts := make([]string, 0, len(partitionedTables)*(yearsAhead+1))
	for _, table := range partitionedTables {
		for year := startYear; year <= startYear+yearsAhead; year++ {
			stmts = append(stmts, fmt.Sprintf(
				`CREATE TABLE IF NOT EXISTS %s_%d PARTITION OF %s FOR VALUES FROM ('%d-01-01') TO ('%d-01-01')`,
				table, year, table, year, year+1,
			))
		}
	}
	return stmts
}

// StartScheduler ejecuta EnsurePartitions una vez inmediatamente, luego en cada tick
// de interval, hasta que se cancele ctx. Los fallos se registran, no son fatales — el
// siguiente tick reintenta, reflejando internal/maintenance.StartScheduler. (Útil)
func StartScheduler(ctx context.Context, svc *Service, interval time.Duration, logger *log.Logger) {
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	if logger == nil {
		logger = log.Default()
	}

	go func() {
		run := func() {
			if err := svc.EnsurePartitions(ctx, time.Now()); err != nil {
				logger.Printf("[WARN] db partition maintenance failed: %v", err)
			}
		}

		run()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}
