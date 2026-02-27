package databases_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/LuckyGuessServices/planets/internal/databases"
	"github.com/LuckyGuessServices/planets/internal/domain/planets"
	"github.com/LuckyGuessServices/planets/internal/domain/snapshot/storage"
	"github.com/LuckyGuessServices/planets/internal/general/types"
	"github.com/LuckyGuessServices/planets/internal/test_tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestMain(m *testing.M) {
	test_tools.RunTestMain(m)
}

// Tests no prepared statements are made via both master and readonly database pools.
//
// See DBPrimaryConfig.DSN
func TestPGXDoesNotPrepareStatementsInsideDbPool(t *testing.T) {
	tests := []struct {
		name         string
		isDBReadonly bool
	}{
		{name: "CORE", isDBReadonly: false},
		{name: "CORE-readonly", isDBReadonly: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				var dbPool *bun.DB
				if test.isDBReadonly {
					dbPool = databases.CoreReadonly()
				} else {
					dbPool = databases.Core()
				}

				ctx := context.Background()

				var notUsed int
				if err := dbPool.NewSelect().ColumnExpr("? + ?", 1, 1).Scan(ctx, &notUsed); err != nil {
					t.Fatalf("Failed to make a query: %v", err)
				}

				var count int
				errSelectPrepared := dbPool.NewSelect().
					TableExpr("pg_prepared_statements").
					ColumnExpr("count(*)").
					Scan(ctx, &count)
				if errSelectPrepared != nil {
					t.Fatalf("Failed to query pg_prepared_statements data: %v", errSelectPrepared)
				}
				if count > 0 {
					t.Fatalf("Expected 0 prepared statements, found %d", count)
				}
			})
		})
	}
}

// Ensures the same database contents between tests (sync bubbles).
//
// See test_tools.RunInSyncBubble, databases.EnableTestPools, databases.Core
func TestDatabaseRollbacksEachTest(t *testing.T) {
	snapshotDate := types.NewDate(time.Now().UTC())

	// Run this test at least twice.
	// Subsequent iteration(s) will fail only if the database is not rolled back after each sync bubble.
	for testIteration := 1; testIteration <= 2; testIteration++ {
		t.Run("iteration-"+strconv.Itoa(testIteration), func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				ctx := context.Background()
				// We can utilize any table here.
				snapshotTable := storage.SnapshotRepository()

				// At first, we should not find any record (we haven't added one yet).
				snapshots, errSelect := snapshotTable.BySnapshotDate(ctx, snapshotDate)
				require.NoErrorf(t, errSelect, "Failed to select snapshots by date '%s': %v", snapshotDate, errSelect)
				require.Emptyf(t, snapshots, "There should be no records initially.")

				// Add a new row.
				newSnapshot := snapshotTable.NewSnapshot()
				newSnapshot.SetSnapshotDate(snapshotDate)
				newSnapshot.SetPlanetIndex(planets.Mercury().Index())
				newSnapshot.SetIsRetrograde(true)
				require.NoErrorf(t, newSnapshot.Save(ctx), "Failed to insert a new snapshot.")

				// Re-request the now newly added record again
				// and compare its contents with the original inserted structure:
				snapshots, errSelect = snapshotTable.BySnapshotDate(ctx, snapshotDate)
				require.NoErrorf(t, errSelect, "Failed to select snapshots by date '%s': %v", snapshotDate, errSelect)
				require.Len(t, snapshots, 1)
				assert.NotSame(t, newSnapshot, snapshots[0])
				assert.Equalf(t, newSnapshot.ID(), snapshots[0].ID(), "It should be the same record added previously.")
			})
		})
	}
}
