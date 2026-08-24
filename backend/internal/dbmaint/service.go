// Package dbmaint keeps the RANGE-partitioned tables (level_attempts,
// daily_streak — see migrations 0007 and 0008) supplied with explicit
// per-year partitions well ahead of time, so the DEFAULT partition added in
// migration 0008 stays empty in practice and never needs to absorb a burst
// of ordinary traffic. This is a structural database concern, independent of
// the legal/privacy retention jobs in internal/maintenance.
package dbmaint

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// yearsAhead controls how many years beyond the current one get an explicit
// partition on every run. 2 years of lead time comfortably survives any
// reasonable gap between scheduler runs or deploys.
const yearsAhead = 2

var partitionedTables = []string{"level_attempts", "daily_streak"}

// Service ensures yearly range partitions exist for the partitioned tables.
// It talks to the raw *sql.DB (not the sqlc-generated Queries) because
// CREATE TABLE ... PARTITION OF is schema DDL, not a query sqlc models.
type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// EnsurePartitions creates (idempotently) a partition for `now`'s year and
// each of the following yearsAhead years, for every partitioned table.
func (s *Service) EnsurePartitions(ctx context.Context, now time.Time) error {
	for _, stmt := range partitionStatements(now) {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("executing %q: %w", stmt, err)
		}
	}
	return nil
}

// partitionStatements builds the idempotent DDL for every partitioned table
// and every year from now's year through now's year + yearsAhead. Table
// names and years are drawn from a fixed internal list/range, never from
// user input, so building SQL with fmt.Sprintf is safe here — there is no
// parameter-binding form for DDL identifiers/ranges in database/sql anyway.
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

// StartScheduler runs EnsurePartitions once immediately, then on every tick
// of interval, until ctx is cancelled. Failures are logged, not fatal — the
// next tick retries, mirroring internal/maintenance.StartScheduler.
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
