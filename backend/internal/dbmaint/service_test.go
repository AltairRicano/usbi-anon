package dbmaint

import (
	"testing"
	"time"
)

func TestYearlyPartitionTargetsCoversBothTablesAndYearRange(t *testing.T) {
	now := time.Date(2027, time.March, 15, 0, 0, 0, 0, time.UTC)
	targets := yearlyPartitionTargets(now)

	wantCount := len(partitionedTables) * (yearsAhead + 1)
	if len(targets) != wantCount {
		t.Fatalf("yearlyPartitionTargets() returned %d targets, want %d", len(targets), wantCount)
	}

	want := []partitionTarget{
		{Table: "level_attempts", Year: 2027},
		{Table: "level_attempts", Year: 2029},
		{Table: "daily_streak", Year: 2027},
	}
	for _, w := range want {
		found := false
		for _, target := range targets {
			if target == w {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("yearlyPartitionTargets() missing expected target: %+v\ngot: %+v", w, targets)
		}
	}
}
