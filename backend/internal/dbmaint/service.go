// Package dbmaint mantiene las tablas particionadas por RANGO (level_attempts,
// daily_streak) provistas con particiones explícitas por año con anticipación,
// para que la partición DEFAULT se mantenga vacía en la práctica y nunca necesite
// absorber tráfico ordinario. Esto es un asunto estructural de la base de datos,
// independiente de los trabajos de retención legal/privacidad en internal/maintenance.
//
// Para mantener la separación DML/DDL, este Service interactúa mediante su propio pool
// autenticado como usbi_dbmaint —un rol sin privilegios sobre tablas, que únicamente
// posee EXECUTE sobre la función SECURITY DEFINER ensure_yearly_partition.
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
// brecha razonable entre ejecuciones del programador o despliegues.
const yearsAhead = 2

var partitionedTables = []string{"level_attempts", "daily_streak"}

// Service asegura que existan particiones de rango anuales para las tablas
// particionadas, llamando a ensure_yearly_partition en vez de ejecutar DDL
// directo — db debe ser el pool de usbi_dbmaint, no el de la aplicación.
type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// partitionTarget identifica una (tabla, año) a asegurar.
type partitionTarget struct {
	Table string
	Year  int
}

// yearlyPartitionTargets calcula qué particiones deben existir para `now`:
// el año actual y cada uno de los yearsAhead siguientes, para cada tabla
// particionada. Función pura (sin acceso a BD) para que el rango de años sea
// verificable sin un Postgres real — la ejecución real vive en
// EnsurePartitions, que llama a ensure_yearly_partition por cada target.
func yearlyPartitionTargets(now time.Time) []partitionTarget {
	startYear := now.UTC().Year()
	targets := make([]partitionTarget, 0, len(partitionedTables)*(yearsAhead+1))
	for _, table := range partitionedTables {
		for year := startYear; year <= startYear+yearsAhead; year++ {
			targets = append(targets, partitionTarget{Table: table, Year: year})
		}
	}
	return targets
}

// EnsurePartitions crea (de forma idempotente) una partición para el año de `now` y
// cada uno de los años en yearsAhead, para cada tabla particionada.
func (s *Service) EnsurePartitions(ctx context.Context, now time.Time) error {
	for _, target := range yearlyPartitionTargets(now) {
		if _, err := s.db.ExecContext(ctx, `SELECT ensure_yearly_partition($1, $2)`, target.Table, target.Year); err != nil {
			return fmt.Errorf("ensuring partition %s_%d: %w", target.Table, target.Year, err)
		}
	}
	return nil
}

// StartScheduler ejecuta EnsurePartitions una vez inmediatamente, luego en cada tick
// de interval, hasta que se cancele ctx. Los fallos se registran, no son fatales — el
// siguiente tick reintenta, reflejando internal/maintenance.StartScheduler.
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
