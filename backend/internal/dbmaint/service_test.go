package dbmaint

import (
	"strings"
	"testing"
	"time"
)

func TestPartitionStatementsCoversBothTablesAndYearRange(t *testing.T) {
	now := time.Date(2027, time.March, 15, 0, 0, 0, 0, time.UTC)
	stmts := partitionStatements(now)

	wantCount := len(partitionedTables) * (yearsAhead + 1)
	if len(stmts) != wantCount {
		t.Fatalf("partitionStatements() returned %d statements, want %d", len(stmts), wantCount)
	}

	want := []string{
		`CREATE TABLE IF NOT EXISTS level_attempts_2027 PARTITION OF level_attempts FOR VALUES FROM ('2027-01-01') TO ('2028-01-01')`,
		`CREATE TABLE IF NOT EXISTS level_attempts_2029 PARTITION OF level_attempts FOR VALUES FROM ('2029-01-01') TO ('2030-01-01')`,
		`CREATE TABLE IF NOT EXISTS daily_streak_2027 PARTITION OF daily_streak FOR VALUES FROM ('2027-01-01') TO ('2028-01-01')`,
	}
	for _, w := range want {
		found := false
		for _, s := range stmts {
			if s == w {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("partitionStatements() missing expected statement: %q\ngot: %s", w, strings.Join(stmts, "\n"))
		}
	}
}
