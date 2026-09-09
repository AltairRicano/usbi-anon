// Package dbmaint mantiene las tablas particionadas por RANGO (level_attempts,
// daily_streak — ver migraciones 0007 y 0008) provistas con particiones explícitas
// por año con mucha antelación, para que la partición DEFAULT añadida en la
// migración 0008 se mantenga vacía en la práctica y nunca necesite absorber una ráfaga
// de tráfico ordinario. Esto es un asunto estructural de la base de datos, independiente de
// los trabajos de retención legal/privacidad en internal/maintenance.
//
// F3 (2026-09-09): antes ejecutaba el DDL crudo (CREATE TABLE ... PARTITION
// OF) directamente sobre el pool compartido de la aplicación. Ni usbi_app ni
// usbi_moderador pueden tener privilegios de DDL sin romper la separación
// DML/DDL que 00_roles_unificado.sql documenta como deliberada, así que este
// Service ahora habla con su propio pool, autenticado como usbi_dbmaint —un
// rol sin ningún GRANT de tabla, solo EXECUTE sobre la función SECURITY
// DEFINER ensure_yearly_partition (migración 0005). (Útil)
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

// Service asegura que existan particiones de rango anuales para las tablas
// particionadas, llamando a ensure_yearly_partition en vez de ejecutar DDL
// directo — db debe ser el pool de usbi_dbmaint, no el de la aplicación. (Útil)
type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// partitionTarget identifica una (tabla, año) a asegurar. (Relleno)
type partitionTarget struct {
	Table string
	Year  int
}

// yearlyPartitionTargets calcula qué particiones deben existir para `now`:
// el año actual y cada uno de los yearsAhead siguientes, para cada tabla
// particionada. Función pura (sin acceso a BD) para que el rango de años sea
// verificable sin un Postgres real — la ejecución real vive en
// EnsurePartitions, que llama a ensure_yearly_partition por cada target. (Útil)
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
// cada uno de los años en yearsAhead, para cada tabla particionada. (Relleno)
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
